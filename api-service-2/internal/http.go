package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"toDoList/models"
)

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		log.Println("Create(): Error decoding body: ", err)
		models.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	b, err := Create(list.Title, list.Description, list.Completed)
	if err != nil {
		log.Println("Create(): Error creating list: ", err)
		models.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response: ", err)
	}
	log.Println("successfully created")
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	lists, err := GetAllTasks()
	if err != nil {
		log.Println("ListHandler(): Error getting all tasks: ", err)
		models.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	b, err := json.MarshalIndent(lists, "", "	")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response: ", err)
	}
	log.Println("successfully lists all tasks")
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	var id struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		log.Println("DeleteHandler(): Error decoding body: ", err)
		models.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	b, err := Delete(id.ID)
	if errors.Is(err, models.ErrNotExists) {
		log.Println("DeleteHandler(): Error deleting task: ", err)
		models.SendJSONError(w, err, http.StatusNotFound)

		return
	}
	if err != nil {
		log.Println("DeleteHandler(): Error deleting task: ", err)
		models.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response: ", err)
	}
	log.Println("successfully deleted")
}

func DoneHandler(w http.ResponseWriter, r *http.Request) {
	var id struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		log.Println("DoneHandler(): Error decoding body: ", err)
		models.SendJSONError(w, err, http.StatusBadRequest)

		return
	}

	b, err := CompleteTask(id.ID)
	if errors.Is(err, models.ErrNotExists) {
		log.Println("DoneHandler(): Error completing task: ", err)
		models.SendJSONError(w, err, http.StatusNotFound)

		return
	}
	if err != nil {
		log.Println("DoneHandler(): Error completing task: ", err)
		models.SendJSONError(w, err, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		log.Println("failed to write response: ", err)
	}
	log.Println("successfully done")
}
