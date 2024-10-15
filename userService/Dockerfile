FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o userService .

EXPOSE 8080

CMD ["./userService"]