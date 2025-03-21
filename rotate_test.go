package dy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotateWriter(t *testing.T) {
	// Create a temporary directory for test log files
	tempDir, err := ioutil.TempDir("", "rotate_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test log file path
	logFile := filepath.Join(tempDir, "test.log")

	// Create a rotate writer with small size limit for testing
	rw, err := NewRotateWriter(logFile,
		WithMaxSize(1),                   // 1MB max size
		WithMaxBackups(3),                // Keep 3 backups
		WithBackupInterval(time.Hour*24), // Daily rotation
		WithCompress(true),               // Compress backups
	)
	if err != nil {
		t.Fatalf("Failed to create rotate writer: %v", err)
	}
	defer rw.Close()

	// Write some test data (enough to trigger rotation)
	data := strings.Repeat("test log line\n", 50000) // ~550KB
	_, err = rw.Write([]byte(data))
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Write more data to trigger rotation
	_, err = rw.Write([]byte(data))
	if err != nil {
		t.Fatalf("Failed to write more data: %v", err)
	}

	// Force a rotation
	if err := rw.ForceRotate(); err != nil {
		t.Fatalf("Failed to force rotation: %v", err)
	}

	// Check that the original log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file doesn't exist after rotation")
	}

	// Check for backup files (allow some time for compression)
	time.Sleep(100 * time.Millisecond)
	matches, err := filepath.Glob(filepath.Join(tempDir, "test.log.*"))
	if err != nil {
		t.Fatalf("Failed to glob backup files: %v", err)
	}

	if len(matches) == 0 {
		t.Errorf("No backup files found after rotation")
	}
}

func TestLoggerWithRotation(t *testing.T) {
	// Create a temporary directory for test log files
	tempDir, err := ioutil.TempDir("", "logger_rotate_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test log file path
	logFile := filepath.Join(tempDir, "app.log")

	// Create a logger with rotation
	logger := New(
		WithRotateWriter(logFile,
			WithMaxSize(1),                // 1MB
			WithMaxBackups(2),             // Keep 2 backups
			WithBackupInterval(time.Hour), // Hourly rotation for testing
		),
		WithLevel(DebugLevel),
	)

	// Log a bunch of messages
	for i := 0; i < 10000; i++ {
		logger.Info("Test log message %d with some padding to make it longer.", i)
	}

	// Force rotation
	rotateWriter, ok := logger.out.(*RotateWriter)
	if !ok {
		t.Fatal("Logger output is not a RotateWriter")
	}
	rotateWriter.ForceRotate()

	// Log a few more messages
	for i := 0; i < 100; i++ {
		logger.Info("Post-rotation test message %d", i)
	}

	// Check that log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file doesn't exist after logging")
	}

	// Read some content to verify logging worked
	content, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Post-rotation test message") {
		t.Errorf("Log file doesn't contain expected post-rotation messages")
	}
}

func TestRotateWriterCleanup(t *testing.T) {
	// Create a temporary directory for test log files
	tempDir, err := ioutil.TempDir("", "rotate_cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test log file path
	logFile := filepath.Join(tempDir, "cleanup.log")

	// Create a rotate writer with low max backups
	rw, err := NewRotateWriter(logFile,
		WithMaxSize(1),      // 1MB
		WithMaxBackups(2),   // Keep only 2 backups
		WithCompress(false), // Don't compress for this test
	)
	if err != nil {
		t.Fatalf("Failed to create rotate writer: %v", err)
	}

	// Create 5 rotated log files
	for i := 0; i < 5; i++ {
		// Write enough to trigger rotation
		data := strings.Repeat(fmt.Sprintf("test data for file %d\n", i), 50000)
		_, err = rw.Write([]byte(data))
		if err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}

		// Force rotation
		if err := rw.ForceRotate(); err != nil {
			t.Fatalf("Failed to force rotation: %v", err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	rw.Close()

	// Allow time for cleanup goroutine
	time.Sleep(200 * time.Millisecond)

	// Check number of backup files - should be 2 + current
	matches, err := filepath.Glob(filepath.Join(tempDir, "cleanup.log*"))
	if err != nil {
		t.Fatalf("Failed to glob backup files: %v", err)
	}

	// Should have the current log file plus 2 backups (3 total)
	if len(matches) > 3 {
		t.Errorf("Expected at most 3 log files, found %d: %v", len(matches), matches)
	}
}
