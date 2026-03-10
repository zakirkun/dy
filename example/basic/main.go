package main

import (
	"time"

	logger "github.com/zakirkun/dy"
)

func main() {
	// Create a logger with JSON format and caller info
	jsonLogger := logger.New(
		logger.WithJSONFormat(true),
		logger.WithCallerInfo(true),
		logger.WithTrace(true),
		logger.WithLevel(logger.DebugLevel),
	)

	// Regular logging with JSON format
	jsonLogger.Info("Application started with JSON logging")
	jsonLogger.Warn("This is a warning message with detailed info")

	// Trace function calls with JSON output
	processUserRequest(jsonLogger, "user123", "view_profile")

	// You can also use the default logger with JSON
	logger.EnableJSONFormat()
	logger.EnableCallerInfo()
	logger.EnableTrace()
	// logger.SetLevel(logger.DebugLevel)
	logger.SetLevel(logger.ParseLevel("DEBUG"))

	// This will now output in JSON format
	logger.Info("Switched default logger to JSON format")

	// Trace a function with the default logger
	defer logger.TraceFunction("args:", "sample")()

	// Do some work
	time.Sleep(50 * time.Millisecond)

	// Log an error
	logger.Error("Something went wrong: %s", "connection timeout")
}

func processUserRequest(log *logger.Logger, userID, action string) {
	// Trace this function with parameters
	defer log.TraceFunction("userID:", userID, "action:", action)()

	log.Info("Processing request for user %s", userID)

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Call nested function
	validateUserAccess(log, userID, action)

	log.Info("Request processed successfully")
}

func validateUserAccess(log *logger.Logger, userID, action string) {
	// Trace this function
	defer log.TraceFunction()()

	log.Debug("Validating user access rights")

	// Simulate database lookup
	time.Sleep(75 * time.Millisecond)

	// Log some information
	log.Info("User %s has access to %s", userID, action)
}
