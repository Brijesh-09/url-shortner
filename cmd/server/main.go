package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	. "github.com/Brijesh-09/internal/handlers"
	. "github.com/Brijesh-09/storage"
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
	// Database connection string
	// Format: postgres://username:password@localhost:5432/database?sslmode=disable
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}
	// Connect to database
	storage, err := NewPostgresStorage(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer storage.Close()

	// Initialize schema
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := storage.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create handler
	handler := NewURLHandler(storage)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/create", handler.Create)
	mux.HandleFunc("/", handler.Redirect)

	// Start server
	log.Println("🚀 Starting server on :9000")
	log.Fatal(http.ListenAndServe(":9000", mux))
}
