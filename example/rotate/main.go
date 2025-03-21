package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	logger "github.com/zakirkun/dy"
)

func main() {
	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		return
	}

	// Create a logger with rotation
	log := logger.New(
		logger.WithRotateWriter(filepath.Join(logsDir, "app.log"),
			logger.WithMaxSize(5),                   // 5MB max size
			logger.WithMaxBackups(3),                // Keep 3 backups
			logger.WithBackupInterval(24*time.Hour), // Daily rotation
			logger.WithCompress(true),               // Compress old logs
		),
		logger.WithLevel(logger.DebugLevel),
		logger.WithTimestamp(true),
		logger.WithColor(true),
		logger.WithPrefix("ROTATE-DEMO"),
	)
	defer log.Close() // Important: close logger to flush buffers

	// Simulate logging over time
	log.Info("Application started with log rotation enabled")
	log.Info("Configuration: max size=5MB, max backups=3, rotate daily")

	// Demonstrate logging at different levels
	log.Debug("This is a debug message")
	log.Info("This is an info message")
	log.Warn("This is a warning message")
	log.Error("This is an error message")

	// Simulate heavy logging to trigger rotation by size
	fmt.Println("Generating log data to demonstrate rotation...")
	for i := 0; i < 100000; i++ {
		log.Info("Log entry %d: This is a longer message to help fill up the log file faster. "+
			"In a real application, logs would accumulate over time naturally.", i)

		// Every 10000 entries, print progress
		if i > 0 && i%10000 == 0 {
			fmt.Printf("Generated %d log entries...\n", i)
		}
	}

	// Force a rotation to demonstrate
	fmt.Println("Forcing log rotation...")
	// Direct access to the RotateWriter from the logger's output
	if rw, ok := log.GetOutput().(*logger.RotateWriter); ok {
		rw.ForceRotate()
	} else {
		fmt.Println("Could not access RotateWriter")
	}

	// Log some more after rotation
	log.Info("Logging after rotation")
	log.Info("Check the logs directory to see rotated files")

	// Display the logs directory content
	files, err := filepath.Glob(filepath.Join(logsDir, "app.log*"))
	if err != nil {
		log.Error("Failed to list log files: %v", err)
		return
	}

	fmt.Println("\nLog files created:")
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("- %s (error getting info: %v)\n", file, err)
			continue
		}
		fmt.Printf("- %s (size: %.2f KB)\n", file, float64(info.Size())/1024)
	}

	fmt.Println("\nLog rotation demo completed successfully!")
}
