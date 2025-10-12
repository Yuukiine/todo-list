package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"toDoList/storage/storage"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file: ", err)
		return
	}

	connStr := os.Getenv("DB_URL")

	s, err := storage.New(connStr)
	if err != nil {
		log.Println("Unable to initialize storage: ", err)
		return
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/create", s.Create)
	router.Get("/all", s.AllIDs)
	router.Get("/list", s.GetAllTasks)
	router.Post("/delete", s.DeleteTask)
	router.Post("/done", s.CompleteTask)

	log.Println("Starting storage-service...")
	if err = http.ListenAndServe(":6701", router); err != nil {
		log.Fatal("failed to start server: ", err.Error())
	}
}
