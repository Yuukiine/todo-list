package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"toDoList/api-service-2/internal"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/create", internal.CreateHandler)
	router.Get("/list", internal.ListHandler)
	router.Delete("/delete", internal.DeleteHandler)
	router.Put("/done", internal.DoneHandler)

	log.Println("Starting internal-service...")
	if err := http.ListenAndServe(":6702", router); err != nil {
		log.Fatal("failed to start server: ", err.Error())
	}
}
