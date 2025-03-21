package dy

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// ErrorData contains extended information about an error
type ErrorData struct {
	Message    string                 `json:"message"`
	Type       string                 `json:"type,omitempty"`
	Stack      []StackFrame           `json:"stack,omitempty"`
	Cause      *ErrorData             `json:"cause,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Code       string                 `json:"code,omitempty"`
}

// StackFrame represents a single frame in the error stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// WithError creates a new logger with detailed error information in context
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}

	// Create the error data structure
	errData := extractErrorData(err, 3) // Skip 3 frames to get to the actual caller

	// Create a new logger with the error data in context
	return l.WithContext("error", errData)
}

// WithErrorCode adds an error code to a logger with error
func (l *Logger) WithErrorCode(code string) *Logger {
	if l == nil {
		return nil
	}

	// Check if there's an existing error context
	l.mu.Lock()
	context := l.context
	l.mu.Unlock()

	if context == nil {
		// No context, just add the code
		return l.WithContext("error_code", code)
	}

	// Look for error data in existing context
	var foundError bool
	var errData ErrorData

	for _, field := range context.Fields {
		if field.Key == "error" {
			if data, ok := field.Value.(ErrorData); ok {
				errData = data
				errData.Code = code
				foundError = true
				break
			}
		}
	}

	if foundError {
		// Create a new logger with updated error data
		newLogger := l.WithoutContext("error")
		return newLogger.WithContext("error", errData)
	}

	// No error found, just add the code
	return l.WithContext("error_code", code)
}

// extractErrorData extracts structured data from an error
func extractErrorData(err error, skip int) ErrorData {
	if err == nil {
		return ErrorData{}
	}

	// Create the base error data
	errData := ErrorData{
		Message:    err.Error(),
		Type:       fmt.Sprintf("%T", err),
		Attributes: make(map[string]interface{}),
	}

	// Capture stack trace if enabled
	errData.Stack = captureStack(skip)

	// Handle wrapped errors (from Go 1.13+)
	var cause error
	if errors.Unwrap(err) != nil {
		cause = errors.Unwrap(err)
		causeData := extractErrorData(cause, 0) // Don't skip frames for cause
		errData.Cause = &causeData
	}

	// Extract additional attributes from custom error types
	extractErrorAttributes(&errData, err)

	return errData
}

// captureStack captures the current stack trace
func captureStack(skip int) []StackFrame {
	const depth = 32
	var pcs [depth]uintptr

	// +2 to skip captureStack and extractErrorData
	n := runtime.Callers(skip+2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	stack := make([]StackFrame, 0, n)

	for {
		frame, more := frames.Next()

		// Skip runtime and standard library frames
		if !strings.Contains(frame.File, "runtime/") && !strings.HasPrefix(frame.File, "/usr/local/go/") {
			stack = append(stack, StackFrame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			})
		}

		if !more || len(stack) >= 16 { // Limit to 16 frames
			break
		}
	}

	return stack
}

// extractErrorAttributes extracts additional attributes from custom error types
func extractErrorAttributes(data *ErrorData, err error) {
	// Check for common error interfaces and extract useful data

	// Check for errors with Code method
	type coder interface {
		Code() string
	}
	if ce, ok := err.(coder); ok {
		data.Code = ce.Code()
	}

	// Check for errors with Timeout method
	type timeoutError interface {
		Timeout() bool
	}
	if te, ok := err.(timeoutError); ok {
		data.Attributes["timeout"] = te.Timeout()
	}

	// Check for errors with Temporary method
	type temporaryError interface {
		Temporary() bool
	}
	if te, ok := err.(temporaryError); ok {
		data.Attributes["temporary"] = te.Temporary()
	}

	// Check for HTTP status code
	type statusCoder interface {
		StatusCode() int
	}
	if sc, ok := err.(statusCoder); ok {
		data.Attributes["status_code"] = sc.StatusCode()
	}

	// Check for errors with additional context fields
	type fielder interface {
		Fields() map[string]interface{}
	}
	if fe, ok := err.(fielder); ok {
		for k, v := range fe.Fields() {
			data.Attributes[k] = v
		}
	}
}

// Define some error interfaces for users to implement

// ErrorWithCode is an interface for errors that have a code
type ErrorWithCode interface {
	error
	Code() string
}

// ErrorWithFields is an interface for errors with additional context fields
type ErrorWithFields interface {
	error
	Fields() map[string]interface{}
}

// SimpleError is a basic implementation of error with code and fields
type SimpleError struct {
	msg    string
	code   string
	fields map[string]interface{}
}

// NewError creates a new error with code and optional fields
func NewError(message string, code string, fields map[string]interface{}) *SimpleError {
	return &SimpleError{
		msg:    message,
		code:   code,
		fields: fields,
	}
}

// Error implements the error interface
func (e *SimpleError) Error() string {
	return e.msg
}

// Code returns the error code
func (e *SimpleError) Code() string {
	return e.code
}

// Fields returns the error context fields
func (e *SimpleError) Fields() map[string]interface{} {
	return e.fields
}

// WrapError wraps an existing error with additional context
func WrapError(err error, message string, code string, fields map[string]interface{}) error {
	if err == nil {
		return nil
	}

	wrapped := &SimpleError{
		msg:    fmt.Sprintf("%s: %s", message, err.Error()),
		code:   code,
		fields: fields,
	}

	return fmt.Errorf("%w", wrapped)
}
