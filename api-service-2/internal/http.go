package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"

	"toDoList/models"
	e "toDoList/pkg/errors"
)

var validate = validator.New()

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		log.Println("Create(): Error decoding body:", err)
		e.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	if err := validate.Struct(list); err != nil {
		log.Println("Create(): Error validating struct:", err)
		e.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	b, err := Create(list.Title, list.Description)
	if err != nil {
		log.Println("Create(): Error creating list:", err)
		e.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response:", err)
	}

	if err = produceEvent("todo-events", "User successfully created task"); err != nil {
		log.Println("ListHandler(): Error producing event: ", err)
	}

	log.Println("successfully created")
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	lists, err := GetAllTasks()
	if err != nil {
		log.Println("ListHandler(): Error getting all tasks:", err)
		e.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	b, err := json.MarshalIndent(lists, "", "	")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response:", err)
	}

	if err = produceEvent("todo-events", "User successfully listed all tasks"); err != nil {
		log.Println("ListHandler(): Error producing event:", err)
	}

	log.Println("successfully lists all tasks")
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	var id struct {
		ID string `json:"id" validate:"required,min=6,max=6"`
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		log.Println("DeleteHandler(): Error decoding body:", err)
		e.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	if err := validate.Struct(id); err != nil {
		log.Println("DeleteHandler(): Error validating struct:", err)
		e.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	b, err := Delete(id.ID)
	if errors.Is(err, e.ErrNotExists) {
		log.Println("DeleteHandler(): Error deleting task:", err)
		e.SendJSONError(w, err, http.StatusNotFound)

		return
	}
	if err != nil {
		log.Println("DeleteHandler(): Error deleting task:", err)
		e.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response:", err)
	}

	if err = produceEvent("todo-events", "User deleted task"); err != nil {
		log.Println("ListHandler(): Error producing event:", err)
	}

	log.Println("successfully deleted")
}

func DoneHandler(w http.ResponseWriter, r *http.Request) {
	var id struct {
		ID string `json:"id" validate:"required,min=6,max=6"`
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		log.Println("DoneHandler(): Error decoding body: ", err)
		e.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	if err := validate.Struct(id); err != nil {
		log.Println("DoneHandler(): Error validating struct:", err)
		e.SendJSONError(w, err, http.StatusBadRequest)
		return
	}

	b, err := CompleteTask(id.ID)
	if errors.Is(err, e.ErrNotExists) {
		log.Println("DoneHandler(): Error completing task:", err)
		e.SendJSONError(w, err, http.StatusNotFound)

		return
	}
	if err != nil {
		log.Println("DoneHandler(): Error completing task:", err)
		e.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response:", err)
	}

	if err = produceEvent("todo-events", "User successfully completed task"); err != nil {
		log.Println("ListHandler(): Error producing event:", err)
	}

	log.Println("successfully done")
}

func produceEvent(topic, message string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:6703"},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	return writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(time.Now().Format(time.RFC3339)),
			Value: []byte("[" + time.Now().Format("15:04:05 02.01.2006") + "]" + ", Message: " + message),
		},
	)
}
