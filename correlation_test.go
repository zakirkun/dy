package dy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Test error types
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

type testErrorWithCode struct {
	msg  string
	code string
}

func (e *testErrorWithCode) Error() string {
	return e.msg
}

func (e *testErrorWithCode) Code() string {
	return e.code
}

type testErrorWithFields struct {
	msg    string
	fields map[string]interface{}
}

func (e *testErrorWithFields) Error() string {
	return e.msg
}

func (e *testErrorWithFields) Fields() map[string]interface{} {
	return e.fields
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Simple error
	err := errors.New("something went wrong")
	l.WithError(err).Error("Operation failed")

	output := buf.String()

	// Check that error message is included
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("Expected error message in output, got: %s", output)
	}

	// Should include error type
	if !strings.Contains(output, "*errors.errorString") {
		t.Errorf("Expected error type in output, got: %s", output)
	}

	// Should include stack trace
	if !strings.Contains(output, "Stack:") {
		t.Errorf("Expected stack trace in output, got: %s", output)
	}
}

func TestWithErrorCode(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Error with code
	err := &testErrorWithCode{
		msg:  "authorization failed",
		code: "AUTH_ERROR",
	}

	l.WithError(err).Error("Permission denied")

	output := buf.String()

	// Check that error code is included
	if !strings.Contains(output, "AUTH_ERROR") {
		t.Errorf("Expected error code in output, got: %s", output)
	}

	// Also test explicitly setting code
	buf.Reset()
	l.WithError(errors.New("generic error")).WithErrorCode("CUSTOM_CODE").Error("Another error")

	output = buf.String()
	if !strings.Contains(output, "CUSTOM_CODE") {
		t.Errorf("Expected custom error code in output, got: %s", output)
	}
}

func TestWithErrorFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Error with additional fields
	err := &testErrorWithFields{
		msg: "database query failed",
		fields: map[string]interface{}{
			"query":       "SELECT * FROM users",
			"duration_ms": 150,
			"user_id":     42,
		},
	}

	l.WithError(err).Error("Database error")

	output := buf.String()

	// Check that fields are included
	if !strings.Contains(output, "query: SELECT * FROM users") {
		t.Errorf("Expected query field in output, got: %s", output)
	}

	if !strings.Contains(output, "duration_ms: 150") {
		t.Errorf("Expected duration field in output, got: %s", output)
	}

	if !strings.Contains(output, "user_id: 42") {
		t.Errorf("Expected user_id field in output, got: %s", output)
	}
}

func TestErrorUnwrapping(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// Create wrapped errors
	err1 := errors.New("original error")
	err2 := fmt.Errorf("intermediate error: %w", err1)
	err3 := fmt.Errorf("outer error: %w", err2)

	l.WithError(err3).Error("Wrapped error occurred")

	output := buf.String()

	// Check that the error chain is captured
	if !strings.Contains(output, "outer error") {
		t.Errorf("Expected outer error in output, got: %s", output)
	}

	if !strings.Contains(output, "Caused by") {
		t.Errorf("Expected 'Caused by' in output, got: %s", output)
	}

	if !strings.Contains(output, "intermediate error") {
		t.Errorf("Expected intermediate error in output, got: %s", output)
	}

	if !strings.Contains(output, "original error") {
		t.Errorf("Expected original error in output, got: %s", output)
	}
}

func TestJSONErrorOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithOutput(&buf),
		WithTimestamp(false),
		WithJSONFormat(true),
	)

	// Error with code and fields
	err := WrapError(
		errors.New("network timeout"),
		"API request failed",
		"API_ERROR",
		map[string]interface{}{
			"endpoint": "/api/data",
			"timeout":  true,
		},
	)

	l.WithError(err).Error("Operation failed")

	// Parse the JSON output
	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check context field contains error data
	context, ok := entry["context"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected context field in JSON output")
	}

	errorData, ok := context["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected error field in context")
	}

	// Check error properties
	if message, ok := errorData["message"].(string); !ok || !strings.Contains(message, "API request failed") {
		t.Errorf("Expected error message in JSON output, got: %v", errorData["message"])
	}

	if code, ok := errorData["code"].(string); !ok || code != "API_ERROR" {
		t.Errorf("Expected error code in JSON output, got: %v", errorData["code"])
	}

	attributes, ok := errorData["attributes"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected attributes in error data")
	}

	if endpoint, ok := attributes["endpoint"].(string); !ok || endpoint != "/api/data" {
		t.Errorf("Expected endpoint in attributes, got: %v", attributes["endpoint"])
	}

	if timeout, ok := attributes["timeout"].(bool); !ok || !timeout {
		t.Errorf("Expected timeout in attributes, got: %v", attributes["timeout"])
	}
}

func TestNewError(t *testing.T) {
	// Test creating a simple error
	err := NewError("something failed", "TEST_ERROR", map[string]interface{}{
		"importance": "high",
		"retry":      false,
	})

	// Check basic properties
	if err.Error() != "something failed" {
		t.Errorf("Expected error message 'something failed', got: %s", err.Error())
	}

	if err.Code() != "TEST_ERROR" {
		t.Errorf("Expected error code 'TEST_ERROR', got: %s", err.Code())
	}

	fields := err.Fields()
	if importance, ok := fields["importance"].(string); !ok || importance != "high" {
		t.Errorf("Expected importance field, got: %v", fields["importance"])
	}

	if retry, ok := fields["retry"].(bool); !ok || retry {
		t.Errorf("Expected retry field to be false, got: %v", fields["retry"])
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original problem")
	wrapped := WrapError(originalErr, "context info", "WRAP_ERROR", map[string]interface{}{
		"context_key": "context_value",
	})

	// Check basic properties
	if !strings.Contains(wrapped.Error(), "context info") {
		t.Errorf("Expected wrapped error to contain context message, got: %s", wrapped.Error())
	}

	if !strings.Contains(wrapped.Error(), "original problem") {
		t.Errorf("Expected wrapped error to contain original message, got: %s", wrapped.Error())
	}

	// Check that it implements our interfaces
	if coder, ok := wrapped.(ErrorWithCode); !ok || coder.Code() != "WRAP_ERROR" {
		t.Errorf("Expected wrapped error to implement ErrorWithCode")
	}

	if fielder, ok := wrapped.(ErrorWithFields); !ok {
		t.Errorf("Expected wrapped error to implement ErrorWithFields")
	} else {
		fields := fielder.Fields()
		if value, ok := fields["context_key"].(string); !ok || value != "context_value" {
			t.Errorf("Expected context field in wrapped error, got: %v", fields)
		}
	}

	// Check that unwrapping works
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped.Error() != originalErr.Error() {
		t.Errorf("Expected unwrapped error to be original, got: %s", unwrapped.Error())
	}
}
