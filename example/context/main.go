package main

import (
	"fmt"
	"time"

	"github.com/zakirkun/dy"
	logger "github.com/zakirkun/dy"
)

func main() {
	// Create a base logger
	log := logger.New(
		logger.WithTimestamp(true),
		logger.WithJSONFormat(false),
		logger.WithLevel(logger.DebugLevel),
	)

	// Start processing an incoming request
	requestLogger := log.WithContext("request_id", "req-abc123")
	requestLogger.Info("Received new request")

	// Authenticate the user
	userLogger := requestLogger.WithContext("user_id", "user-456")
	userLogger.Info("User authenticated")

	// Process a transaction
	processTransaction(userLogger, "tx-789")

	// Another example with plain text format
	textLog := logger.New(
		logger.WithColor(true),
		logger.WithCallerInfo(true),
	)

	// Create service-wide context
	serviceLog := textLog.WithFields(map[string]interface{}{
		"service": "payment-api",
		"version": "2.1",
		"env":     "production",
	})

	// Handle a specific request
	requestLog := serviceLog.WithContext("client_ip", "192.168.1.1").
		WithContext("method", "POST").
		WithContext("path", "/api/v1/payments")

	requestLog.Info("Request started")

	// Try an operation that might fail
	if err := riskyOperation(); err != nil {
		// Automatically add error info to context
		requestLog.WithError(err).Error("Operation failed")
	}

	// Remove sensitive data for certain logs
	limitedLog := requestLog.WithoutContext("client_ip")
	limitedLog.Info("Sending to analytics service")

	// Base logger
	baseLogger := dy.New(dy.WithLevel(dy.InfoLevel))

	// Create request-scoped logger with request ID
	reqLogger := baseLogger.WithContext("request_id", "abc-123")
	reqLogger.Info("Request received") // Logs with request_id=abc-123

	// Create user-scoped child logger that inherits request context
	userLogger = reqLogger.WithContext("user_id", "user-456")
	userLogger.Info("User authenticated") // Logs with request_id=abc-123 AND user_id=user-456

	// Add multiple fields at once
	txLogger := userLogger.WithFields(map[string]interface{}{
		"transaction_id": "tx-789",
		"amount":         99.95,
		"currency":       "USD",
	})
	txLogger.Info("Transaction started") // Logs with ALL context values

	// Remove sensitive context for certain logs
	sanitizedLogger := txLogger.WithoutContext("user_id")
	sanitizedLogger.Info("Metrics collected") // Logs without the user_id

	// Automatically capture errors
	if err := processPayment(); err != nil {
		txLogger.WithError(err).Error("Payment failed")
	}
}

func processPayment() error {
	// Simulate an error
	return fmt.Errorf("payment gateway timeout after 30s")
}

func processTransaction(log *logger.Logger, txID string) {
	// Create transaction-specific logger
	txLog := log.WithContext("transaction_id", txID)

	txLog.Debug("Starting transaction processing")

	// Simulate work
	time.Sleep(100 * time.Millisecond)

	// Add more context as processing continues
	paymentLog := txLog.WithContext("payment_method", "credit_card").
		WithContext("amount", 125.50).
		WithContext("currency", "USD")

	paymentLog.Info("Payment processed successfully")

	// The log contains request_id, user_id, transaction_id, payment_method, amount, and currency
}

func riskyOperation() error {
	// Simulate an error
	return fmt.Errorf("database connection timeout after 30s")
}
