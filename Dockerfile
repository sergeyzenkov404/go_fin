FROM golang:1.27.1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


ENV TODO_PORT=7545 \
    TODO_PASSWORD=12345


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/go_fin

CMD ["/app/go_fin"] 