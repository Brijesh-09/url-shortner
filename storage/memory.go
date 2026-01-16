package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Brijesh-09/internal/models"
	_ "github.com/lib/pq"
)

var (
	ErrNotFound      = errors.New("short code not found")
	ErrAlreadyExists = errors.New("short code already exists")
)

type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL connection
func NewPostgresStorage(connectionString string) (*PostgresStorage, error) {
	// Open database connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Max open connections
	db.SetMaxIdleConns(5)                  // Max idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max connection lifetime

	log.Println("✅ Connected to PostgreSQL")
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// InitSchema creates the URLs table if it doesn't exist
func (s *PostgresStorage) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			short_code VARCHAR(255) PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_created_at ON urls(created_at);
	`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("✅ Database schema ready")
	return nil
}

// NewMemoryStorage creates a new in-memory storage
// Save stores a URL in the database
// ctx is passed in so we can cancel if needed
func (s *PostgresStorage) Save(ctx context.Context, url models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url)
		VALUES ($1, $2)
	`

	// Execute with context - if ctx is cancelled, query stops
	_, err := s.db.ExecContext(ctx, query, url.ShortCode, url.OriginalURL)
	if err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key") {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

// Get retrieves a URL from the database
func (s *PostgresStorage) Get(ctx context.Context, shortCode string) (string, error) {
	query := `
		SELECT original_url FROM urls WHERE short_code = $1
	`

	var originalURL string
	err := s.db.QueryRowContext(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return originalURL, nil
}

// delete
func (s *PostgresStorage) Delete(ctx context.Context, shortCode string) error {
	query := `
		DELETE FROM urls WHERE short_code = $1
	`

	result, err := s.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
