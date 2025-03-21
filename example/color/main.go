package main

import (
	"os"

	logger "github.com/zakirkun/dy"
)

func main() {
	// Create a logger with color output
	log := logger.New(
		logger.WithColor(true),
		logger.WithLevel(logger.DebugLevel),
	)

	// Show all log levels with colors
	log.Debug("This is a DEBUG message (blue)")
	log.Info("This is an INFO message (green)")
	log.Warn("This is a WARN message (yellow)")
	log.Error("This is an ERROR message (red)")

	// Don't actually call Fatal since it will exit the program
	// log.Fatal("This is a FATAL message (bold red)")

	// Example with colors disabled
	noColorLog := logger.New(
		logger.WithColor(false),
		logger.WithOutput(os.Stdout),
	)

	noColorLog.Info("This message has no colors")

	// Example with trace function and colors
	log.EnableTrace()
	defer log.TraceFunction("with colors")()

	log.Info("Function tracing with colors works too!")
}
