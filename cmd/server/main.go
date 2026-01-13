package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Brijesh-09/internal/handlers"
	"github.com/Brijesh-09/storage"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[LOG] Started %s %s", r.Method, r.URL.Path)

		next(w, r)

		duration := time.Since(start)
		log.Printf("[LOG] Completed in %v", duration)
	}
}
func main() {

	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize handlers
	handler := handlers.NewURLHandler(store)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", Logger(handler.Health))
	mux.HandleFunc("/create", Logger(handler.Create))
	mux.HandleFunc("/", Logger(handler.Redirect))

	// Start server
	log.Println("Starting server on :9000")
	log.Fatal(http.ListenAndServe(":9000", mux))

}
