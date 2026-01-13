package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Brijesh-09/internal/models"
	"github.com/Brijesh-09/storage"
)

// URLHandler handles URL shortening requests
type URLHandler struct {
	storage *storage.MemoryStorage
}

// NewURLHandler creates a new URL handler
func NewURLHandler(store *storage.MemoryStorage) *URLHandler {
	return &URLHandler{storage: store}
}

// Health checks if server is running
func (h *URLHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server Healthy and running"))
}

// Create handles URL shortening creation
func (h *URLHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var urlData models.URL
	if err := json.NewDecoder(r.Body).Decode(&urlData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if urlData.OriginalURL == "" || urlData.ShortCode == "" {
		http.Error(w, "original_url and short_code are required", http.StatusBadRequest)
		return
	}

	// Save to storage
	if err := h.storage.Save(urlData); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			http.Error(w, "Short code already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Short URL created",
		"short_code": urlData.ShortCode,
		"short_url":  "http://localhost:9000/" + urlData.ShortCode,
	})
}

// Redirect handles redirection to original URL
func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	// Skip known routes
	if code == "" || code == "health" || code == "create" {
		return
	}

	// Get original URL
	originalURL, err := h.storage.Get(code)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/delete/")

	if code == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	// Delete from storage
	if err := h.storage.Delete(code); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Short URL deleted",
		"short_code": code,
	})
}
