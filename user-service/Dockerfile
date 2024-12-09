FROM golang:1.23.2 AS builder

WORKDIR /app

COPY . .

RUN go work use ./user-service ./email-service

WORKDIR /app/user-service
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/user-service ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /usr/local/bin/user-service .

CMD ["./user-service"]
