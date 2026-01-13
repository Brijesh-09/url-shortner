package storage

import (
	"errors"

	"github.com/Brijesh-09/internal/models"
)

var (
	ErrNotFound      = errors.New("short code not found")
	ErrAlreadyExists = errors.New("short code already exists")
)

// MemoryStorage stores URLs in memory
type MemoryStorage struct {
	urls map[string]string
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls: make(map[string]string),
	}
}

// Save stores a URL mapping
func (s *MemoryStorage) Save(url models.URL) error {
	if _, exists := s.urls[url.ShortCode]; exists {
		return ErrAlreadyExists
	}
	s.urls[url.ShortCode] = url.OriginalURL
	return nil
}

// Get retrieves the original URL by short code
func (s *MemoryStorage) Get(shortCode string) (string, error) {
	originalURL, exists := s.urls[shortCode]
	if !exists {
		return "", ErrNotFound
	}
	return originalURL, nil
}

func (s *MemoryStorage) Delete(shortCode string) error {
	if _, exits := s.urls[shortCode]; !exits {
		return ErrNotFound
	}
	delete(s.urls, shortCode)
	return nil
}
