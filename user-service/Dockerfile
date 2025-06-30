FROM golang:1.24 AS builder

WORKDIR /app

# Copy all modules and workspace files
COPY go.work go.work.sum ./
COPY email-service ./email-service
COPY user-service ./user-service
COPY common ./common

# Download dependencies for the workspace
RUN go work sync

# Build the binary from workspace root
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/user-service ./user-service/cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /usr/local/bin/user-service .

CMD ["./user-service"]
