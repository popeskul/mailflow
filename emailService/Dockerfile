FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o emailService .

EXPOSE 9090

CMD ["./emailService"]