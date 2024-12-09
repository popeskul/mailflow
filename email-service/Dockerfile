FROM golang:1.23.2 AS builder

WORKDIR /app

COPY . .

RUN go work use ./email-service ./user-service

WORKDIR /app/email-service
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/email-service ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /usr/local/bin/email-service .

CMD ["./email-service"]