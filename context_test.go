package dy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Create a child logger with context
	childLogger := l.WithContext("request_id", "abc-123")
	childLogger.Info("Request received")

	output := buf.String()
	if !strings.Contains(output, "request_id: abc-123") {
		t.Errorf("Expected context field in output, got: %s", output)
	}
}

func TestLoggerWithMultipleContexts(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Create nested context loggers
	requestLogger := l.WithContext("request_id", "req-123")
	userLogger := requestLogger.WithContext("user_id", "user-456")
	txLogger := userLogger.WithContext("transaction_id", "tx-789")

	// Log with the most nested logger
	txLogger.Info("Transaction started")

	output := buf.String()

	// Check that all context fields are present
	if !strings.Contains(output, "request_id: req-123") {
		t.Errorf("Expected request_id in output, got: %s", output)
	}
	if !strings.Contains(output, "user_id: user-456") {
		t.Errorf("Expected user_id in output, got: %s", output)
	}
	if !strings.Contains(output, "transaction_id: tx-789") {
		t.Errorf("Expected transaction_id in output, got: %s", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Add multiple fields at once
	fieldsLogger := l.WithFields(map[string]interface{}{
		"service": "api",
		"version": "1.0",
		"region":  "us-west",
	})

	fieldsLogger.Info("Service starting")

	output := buf.String()

	// Check that all fields are present
	if !strings.Contains(output, "service: api") {
		t.Errorf("Expected service field in output, got: %s", output)
	}
	if !strings.Contains(output, "version: 1.0") {
		t.Errorf("Expected version field in output, got: %s", output)
	}
	if !strings.Contains(output, "region: us-west") {
		t.Errorf("Expected region field in output, got: %s", output)
	}
}

func TestLoggerWithoutContext(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Create a logger with multiple contexts
	contextLogger := l.WithContext("request_id", "req-123").
		WithContext("user_id", "user-456").
		WithContext("temporary", "value")

	// Remove one context key
	reducedLogger := contextLogger.WithoutContext("temporary")

	reducedLogger.Info("Processing request")

	output := buf.String()

	// Check that removed field is not present
	if strings.Contains(output, "temporary: value") {
		t.Errorf("Expected 'temporary' field to be removed, got: %s", output)
	}

	// Check that other fields are still present
	if !strings.Contains(output, "request_id: req-123") {
		t.Errorf("Expected request_id field to be present, got: %s", output)
	}
	if !strings.Contains(output, "user_id: user-456") {
		t.Errorf("Expected user_id field to be present, got: %s", output)
	}
}

func TestLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Create an error and add it to context
	err := fmt.Errorf("something went wrong")
	errorLogger := l.WithError(err)

	errorLogger.Error("Operation failed")

	output := buf.String()

	// Check that error is in context
	if !strings.Contains(output, "error: something went wrong") {
		t.Errorf("Expected error in context, got: %s", output)
	}
}

func TestLoggerWithJSONContext(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithOutput(&buf),
		WithTimestamp(false),
		WithJSONFormat(true),
	)

	// Create a logger with context and log in JSON format
	contextLogger := l.WithContext("request_id", "req-123").
		WithContext("user_id", "user-456")

	contextLogger.Info("Processing request")

	// Parse the JSON output
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check that context fields are in the JSON
	if entry.Context == nil {
		t.Fatalf("Expected context in JSON output, got nil")
	}

	if requestID, ok := entry.Context["request_id"]; !ok || requestID != "req-123" {
		t.Errorf("Expected request_id in JSON context, got: %v", entry.Context)
	}

	if userID, ok := entry.Context["user_id"]; !ok || userID != "user-456" {
		t.Errorf("Expected user_id in JSON context, got: %v", entry.Context)
	}
}

func TestContextIsIndependent(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Create first context logger
	firstLogger := l.WithContext("type", "first")

	// Create second context logger from base
	secondLogger := l.WithContext("type", "second")

	// Log with both loggers
	buf.Reset()
	firstLogger.Info("First logger")
	firstOutput := buf.String()

	buf.Reset()
	secondLogger.Info("Second logger")
	secondOutput := buf.String()

	// Check that contexts are independent
	if !strings.Contains(firstOutput, "type: first") {
		t.Errorf("Expected 'type: first' in first logger, got: %s", firstOutput)
	}

	if !strings.Contains(secondOutput, "type: second") {
		t.Errorf("Expected 'type: second' in second logger, got: %s", secondOutput)
	}

	if strings.Contains(firstOutput, "type: second") {
		t.Errorf("First logger should not contain second logger's context")
	}

	if strings.Contains(secondOutput, "type: first") {
		t.Errorf("Second logger should not contain first logger's context")
	}
}
