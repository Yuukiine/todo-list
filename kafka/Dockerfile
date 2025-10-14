FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o kafka-app ./kafka/main.go

FROM golang:1.24
WORKDIR /app
COPY --from=builder /app/kafka-app .
CMD ["./kafka-app"]