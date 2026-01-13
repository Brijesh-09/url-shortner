package main

import (
	"log"
	"net/http"

	"github.com/Brijesh-09/internal/handlers"
	"github.com/Brijesh-09/storage"
)

func main() {

	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize handlers
	handler := handlers.NewURLHandler(store)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/create", handler.Create)
	mux.HandleFunc("/", handler.Redirect)

	// Start server
	log.Println("Starting server on :9000")
	log.Fatal(http.ListenAndServe(":9000", mux))

}
