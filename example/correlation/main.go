package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	logger "github.com/zakirkun/dy"
)

// Example custom error types
type DBError struct {
	Err        error
	Query      string
	Params     []interface{}
	RetryCount int
}

func (e *DBError) Error() string {
	return fmt.Sprintf("database error: %s", e.Err)
}

func (e *DBError) Unwrap() error {
	return e.Err
}

// Implement Fields() to add contextual information
func (e *DBError) Fields() map[string]interface{} {
	return map[string]interface{}{
		"query":       e.Query,
		"retry_count": e.RetryCount,
	}
}

// Implement Code() for error code
func (e *DBError) Code() string {
	if errors.Is(e.Err, sql.ErrNoRows) {
		return "DB_NOT_FOUND"
	}
	return "DB_ERROR"
}

// HTTP Error with status code
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s", e.Code, e.Message)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}

func (e *HTTPError) Fields() map[string]interface{} {
	return map[string]interface{}{
		"http_status": e.Code,
	}
}

func main() {
	// Create a logger with color output
	log := logger.New(
		logger.WithLevel(logger.DebugLevel),
		logger.WithColor(true),
		logger.WithCallerInfo(true),
	)

	// Simulate a request handler
	handleRequest(log, "/api/users/123")
}

func handleRequest(log *logger.Logger, path string) {
	// Add request context
	reqLog := log.WithContext("path", path).
		WithContext("method", "GET").
		WithContext("request_id", "req-abc-123")

	reqLog.Info("Processing request")

	// Try to get user from database
	user, err := getUserFromDB(123)
	if err != nil {
		// Log with detailed error information - automatically captures stack trace,
		// error type, and extracts fields from our custom error type
		reqLog.WithError(err).Error("Failed to get user")

		// We can also add an error code if needed
		reqLog.WithError(err).WithErrorCode("USER_RETRIEVAL_FAILED").Error("User operation failed")
		return
	}

	// Try to call an external API
	err = callExternalAPI(context.Background())
	if err != nil {
		// Log with wrapped error
		wrappedErr := logger.WrapError(err, "API integration failed", "EXT_API_ERROR", map[string]interface{}{
			"timeout_ms": 500,
			"endpoint":   "https://api.example.com/data",
		})

		reqLog.WithError(wrappedErr).Error("External system integration failed")
		return
	}

	reqLog.WithContext("user_id", user.ID).Info("Request completed successfully")
}

// Simulate getting a user from database with error
func getUserFromDB(id int) (*User, error) {
	// Simulate database error
	dbErr := sql.ErrConnDone

	// Wrap in our custom error type that adds context
	return nil, &DBError{
		Err:        dbErr,
		Query:      "SELECT * FROM users WHERE id = ?",
		Params:     []interface{}{id},
		RetryCount: 2,
	}
}

// Simulate calling an external API
func callExternalAPI(ctx context.Context) error {
	// Simulate a timeout
	time.Sleep(100 * time.Millisecond)

	// Create a context timeout
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	// Simulate HTTP error
	return &HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "External API is currently unavailable",
		Err:     context.DeadlineExceeded,
	}
}

// User represents a database user
type User struct {
	ID    int
	Name  string
	Email string
}
