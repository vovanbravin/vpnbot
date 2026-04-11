FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
COPY .env ./

RUN go mod download

COPY . .

RUN go build -o vpnbot ./cmd/bot/main.go

CMD ["./vpnbot"]

