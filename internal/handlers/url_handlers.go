package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Brijesh-09/internal/models"
	. "github.com/Brijesh-09/storage"
)

type URLHandler struct {
	storage *PostgresStorage
}

func NewURLHandler(store *PostgresStorage) *URLHandler {
	return &URLHandler{storage: store}
}

func (h *URLHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Create a context with 2 second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel() // Always call cancel to free resources!

	// Try to ping database
	if err := h.storage.Ping(ctx); err != nil {
		http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server and Database Healthy"))
}

func (h *URLHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON
	var urlData models.URL
	if err := json.NewDecoder(r.Body).Decode(&urlData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate
	if urlData.OriginalURL == "" || urlData.ShortCode == "" {
		http.Error(w, "original_url and short_code are required", http.StatusBadRequest)
		return
	}

	// Create context with 5 second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Save to database
	if err := h.storage.Save(ctx, urlData); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			http.Error(w, "Short code already exists", http.StatusConflict)
			return
		}
		log.Printf("Error saving URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Short URL created",
		"short_code": urlData.ShortCode,
		"short_url":  "http://localhost:9000/" + urlData.ShortCode,
	})
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	if code == "" || code == "health" || code == "create" {
		return
	}

	// Create context with 3 second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Get from database
	originalURL, err := h.storage.Get(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
