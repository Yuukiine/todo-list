package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"toDoList/models"
	e "toDoList/pkg/errors"
)

const lenOfOrderID = 6

func Create(title, description string, completed bool) ([]byte, error) {
	list := models.List{
		ID:          generateUniqueID(),
		Title:       title,
		Description: description,
		Completed:   completed,
	}

	b, err := json.MarshalIndent(list, "", "	")
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("http://storage-service:6701/create", "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("failed to create a new list")
	}

	return b, nil
}

func GetAllTasks() ([]models.List, error) {
	var lists []models.List
	resp, err := http.Get("http://storage-service:6701/list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&lists); err != nil {
		return nil, err
	}

	return lists, nil
}

func Delete(id string) ([]byte, error) {
	resp, err := http.Post("http://storage-service:6701/delete", "application/json", bytes.NewBuffer([]byte(`
	{
		"id": "`+id+`"
	}`)))
	if resp.StatusCode == http.StatusNotFound {
		return nil, e.ErrNotExists
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return []byte("{}"), nil
}

func CompleteTask(id string) ([]byte, error) {
	resp, err := http.Post("http://storage-service:6701/done", "application/json", bytes.NewBuffer([]byte(`
	{
		"id": "`+id+`"
	}`)))
	if resp.StatusCode == http.StatusNotFound {
		return nil, e.ErrNotExists
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateUniqueID() string {
	var a struct {
		IDs []string `json:"ids"`
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	m := make(map[string]struct{})
	const letters = "1234567890abcdefghijklmnopqrstuvwxyz"

	resp, err := http.Get("http://storage-service:6701/all")
	if err != nil {
		log.Println("failed to get response: ", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a); err != nil {
		log.Println("failed to decode response: ", err)
	}

	for _, n := range a.IDs {
		m[n] = struct{}{}
	}

	for {
		output := make([]byte, lenOfOrderID)

		for i := range lenOfOrderID {
			n := rng.Intn(len(letters))
			if i == 0 && n == 0 {
				n = 1
			}
			output[i] = letters[n]
		}

		if _, exists := m[string(output)]; !exists {
			return string(output)
		}
	}
}
