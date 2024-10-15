#!/bin/bash

API_PATH="api"
OUTPUT_FILE="swagger/swagger.swagger.json"

# Function to merge JSON files
merge_json() {
    jq -s '
    reduce .[] as $item ({};
        . * $item
        | .paths += $item.paths
        | .definitions += $item.definitions
        | .securityDefinitions += $item.securityDefinitions
    )' "$@"
}

# Find all Swagger files
swagger_files=$(find "$API_PATH" -name "*.swagger.json")

# Combine all Swagger files
merge_json $swagger_files > "$OUTPUT_FILE"

echo "Combined Swagger file created at $OUTPUT_FILE"
