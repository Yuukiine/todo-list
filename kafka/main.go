package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

func ConsumeMessage(r *kafka.Reader, w *bufio.Writer) {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Error reading message:", err)
			continue
		}

		log.Println("Received: ", m.Value)

		_, err = w.WriteString(string(m.Value) + "\n")
		if err != nil {
			log.Println("Error writing to file:", err)
		}

		w.Flush()
	}
}

func main() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "todo-events",
		GroupID: "file-writer-group",
	})

	file, err := os.OpenFile("events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	ConsumeMessage(r, writer)
}
