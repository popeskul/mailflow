#!/bin/bash

# coverage_report.sh - Generate HTML coverage reports for all services
set -e

PROJECT_ROOT="/Users/ppopeskul/dev/mailflow"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
SERVICES=("user-service" "email-service")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📊 Generating Coverage Reports${NC}"
echo "=================================="

# Create coverage directory structure
mkdir -p "$COVERAGE_DIR"/{html,profiles}

# Function to generate coverage for a service
generate_service_coverage() {
    local service=$1
    local service_dir="$PROJECT_ROOT/$service"
    local profile_file="$COVERAGE_DIR/profiles/${service}.out"
    local html_file="$COVERAGE_DIR/html/${service}.html"
    
    echo -e "\n${YELLOW}📦 Processing $service${NC}"
    echo "----------------------------"
    
    cd "$service_dir"
    
    # Generate coverage profile
    echo "Generating coverage profile..."
    go test -coverprofile="$profile_file" -covermode=atomic ./... 2>/dev/null || {
        echo -e "${RED}❌ Failed to generate coverage for $service${NC}"
        return 1
    }
    
    # Check if profile was created
    if [[ ! -f "$profile_file" ]]; then
        echo -e "${RED}❌ No coverage profile generated for $service${NC}"
        return 1
    fi
    
    # Generate HTML report
    echo "Generating HTML report..."
    go tool cover -html="$profile_file" -o="$html_file"
    
    # Calculate coverage percentage
    local coverage_pct=$(go tool cover -func="$profile_file" | grep total | awk '{print $3}')
    
    # Color code based on coverage percentage
    local coverage_num=${coverage_pct%\%}
    local color=$RED
    if (( $(echo "$coverage_num >= 80" | bc -l) )); then
        color=$GREEN
    elif (( $(echo "$coverage_num >= 60" | bc -l) )); then
        color=$YELLOW
    fi
    
    echo -e "${color}✅ $service: $coverage_pct coverage${NC}"
    echo "   📄 Profile: $profile_file"
    echo "   🌐 HTML: $html_file"
    
    return 0
}

# Generate coverage for each service
total_services=${#SERVICES[@]}
successful_services=0

for service in "${SERVICES[@]}"; do
    if generate_service_coverage "$service"; then
        ((successful_services++))
    fi
done

# Generate combined coverage report
echo -e "\n${BLUE}🔗 Generating Combined Coverage Report${NC}"
echo "======================================"

combined_profile="$COVERAGE_DIR/profiles/combined.out"
combined_html="$COVERAGE_DIR/html/combined.html"

# Combine all coverage profiles
{
    echo "mode: atomic"
    for service in "${SERVICES[@]}"; do
        profile_file="$COVERAGE_DIR/profiles/${service}.out"
        if [[ -f "$profile_file" ]]; then
            tail -n +2 "$profile_file" 2>/dev/null || true
        fi
    done
} > "$combined_profile"

# Generate combined HTML report
if [[ -s "$combined_profile" ]]; then
    go tool cover -html="$combined_profile" -o="$combined_html"
    
    # Calculate combined coverage
    combined_coverage=$(go tool cover -func="$combined_profile" | grep total | awk '{print $3}')
    echo -e "${GREEN}✅ Combined Coverage: $combined_coverage${NC}"
    echo "   📄 Profile: $combined_profile"
    echo "   🌐 HTML: $combined_html"
fi

# Generate summary report
echo -e "\n${BLUE}📈 Coverage Summary${NC}"
echo "=================="
echo "Services processed: $successful_services/$total_services"

for service in "${SERVICES[@]}"; do
    profile_file="$COVERAGE_DIR/profiles/${service}.out"
    if [[ -f "$profile_file" ]]; then
        coverage_pct=$(go tool cover -func="$profile_file" | grep total | awk '{print $3}')
        coverage_num=${coverage_pct%\%}
        
        if (( $(echo "$coverage_num >= 80" | bc -l) )); then
            color=$GREEN
            status="🟢"
        elif (( $(echo "$coverage_num >= 60" | bc -l) )); then
            color=$YELLOW
            status="🟡"
        else
            color=$RED
            status="🔴"
        fi
        
        echo -e "$status ${color}$service: $coverage_pct${NC}"
    else
        echo -e "🔴 ${RED}$service: No coverage data${NC}"
    fi
done

# Show file locations
echo -e "\n${BLUE}📁 File Locations${NC}"
echo "================="
echo "Coverage directory: $COVERAGE_DIR"
echo "HTML reports: $COVERAGE_DIR/html/"
echo "Coverage profiles: $COVERAGE_DIR/profiles/"

# Open combined report
if [[ -f "$combined_html" ]]; then
    echo -e "\n${GREEN}🚀 Opening combined coverage report...${NC}"
    if command -v open >/dev/null 2>&1; then
        open "$combined_html"
    elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open "$combined_html"
    else
        echo "Open manually: $combined_html"
    fi
fi

echo -e "\n${GREEN}✅ Coverage report generation completed!${NC}"
