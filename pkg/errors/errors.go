package errors

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

var (
	ErrInternalServer = errors.New("internal server error")
	ErrNotExists      = errors.New("task doesn't exists")
)

type JSONError struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func SendJSONError(w http.ResponseWriter, err error, status int) {
	errJSON := &JSONError{
		Message:   err.Error(),
		Timestamp: time.Now().Format("15:04:05 02.01.2006"),
	}

	b, err := json.MarshalIndent(errJSON, "", "    ")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(status)
	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
	log.Println("JSON Error sent successfully")
}
