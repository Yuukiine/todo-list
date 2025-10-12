FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../api-service .

RUN go build -o internal-app ./internal/main.go

FROM golang:1.24
WORKDIR /app
COPY --from=builder /app/api-app .
CMD ["./api-app"]