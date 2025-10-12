FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o storage-app ./storage/main.go

FROM golang:1.24
WORKDIR /app
COPY --from=builder /app/storage-app .
COPY .env .

CMD ["./storage-app"]