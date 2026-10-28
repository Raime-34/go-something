package main

import (
	"log"
	"net/http"

	"tasks-api/internal/handlers"
	"tasks-api/internal/middleware"
	"tasks-api/internal/storage"
)

func main() {
	store := storage.NewMemory()
	h := handlers.New(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/tasks", h.TasksCollection) // GET, POST
	mux.HandleFunc("/tasks/", h.TaskItem)       // GET, PUT, DELETE

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", middleware.Logging(mux)); err != nil {
		log.Fatal(err)
	}
}
