package dy

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// RotateWriter implements io.Writer with automatic log rotation
// This type is exported so it can be accessed by users for advanced operations
type RotateWriter struct {
	mu             sync.Mutex
	filename       string        // Log file path
	file           *os.File      // Current file handle
	size           int64         // Current file size
	maxSize        int64         // Maximum size in bytes before rotation
	maxBackups     int           // Maximum number of backups to keep
	backupInterval time.Duration // Time interval for rotation regardless of size
	lastRotate     time.Time     // Time of last rotation
	compress       bool          // Whether to compress backup files
}

// RotateOption defines options for the RotateWriter
type RotateOption func(*RotateWriter)

// WithMaxSize sets the maximum file size before rotation
func WithMaxSize(megabytes int) RotateOption {
	return func(rw *RotateWriter) {
		rw.maxSize = int64(megabytes) * 1024 * 1024
	}
}

// WithMaxBackups sets the maximum number of backup files to keep
func WithMaxBackups(count int) RotateOption {
	return func(rw *RotateWriter) {
		rw.maxBackups = count
	}
}

// WithBackupInterval sets the time interval for regular rotation
func WithBackupInterval(duration time.Duration) RotateOption {
	return func(rw *RotateWriter) {
		rw.backupInterval = duration
	}
}

// WithCompress enables or disables backup compression
func WithCompress(compress bool) RotateOption {
	return func(rw *RotateWriter) {
		rw.compress = compress
	}
}

// NewRotateWriter creates a new rotate writer
func NewRotateWriter(filename string, options ...RotateOption) (*RotateWriter, error) {
	rw := &RotateWriter{
		filename:       filename,
		maxSize:        100 * 1024 * 1024, // Default: 100MB
		maxBackups:     5,                 // Default: keep 5 backup files
		backupInterval: 24 * time.Hour,    // Default: rotate daily
		compress:       true,              // Default: compress backups
		lastRotate:     time.Now(),
	}

	// Apply options
	for _, option := range options {
		option(rw)
	}

	// Open or create the log file
	if err := rw.openFile(); err != nil {
		return nil, err
	}

	return rw, nil
}

// openFile opens or creates the log file
func (rw *RotateWriter) openFile() error {
	// Ensure directory exists
	dir := filepath.Dir(rw.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open the file with append mode
	file, err := os.OpenFile(rw.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get the current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	rw.file = file
	rw.size = info.Size()
	return nil
}

// Write implements io.Writer for logger output
func (rw *RotateWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// If file is not opened, try to open it
	if rw.file == nil {
		if err := rw.openFile(); err != nil {
			return 0, err
		}
	}

	// Check if we need to rotate based on size or time
	if (rw.maxSize > 0 && rw.size+int64(len(p)) > rw.maxSize) ||
		(rw.backupInterval > 0 && time.Since(rw.lastRotate) > rw.backupInterval) {
		if err := rw.rotate(); err != nil {
			return 0, err
		}
	}

	// Write to the file
	n, err = rw.file.Write(p)
	rw.size += int64(n)
	return n, err
}

// Close closes the current file
func (rw *RotateWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.file == nil {
		return nil
	}

	err := rw.file.Close()
	rw.file = nil
	return err
}

// rotate performs the actual log rotation
func (rw *RotateWriter) rotate() error {
	// Close the current file
	if rw.file != nil {
		if err := rw.file.Close(); err != nil {
			return err
		}
		rw.file = nil
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s", rw.filename, timestamp)

	// Rename the current log file to backup name
	if err := os.Rename(rw.filename, backupName); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to rename log file: %w", err)
		}
		// If the file doesn't exist, just continue with creating a new one
	} else {
		// Compress the backup if enabled
		if rw.compress {
			go func(name string) {
				if err := compressFile(name); err != nil {
					// Log error but continue - don't want to block main thread
					fmt.Fprintf(os.Stderr, "Failed to compress backup: %v\n", err)
				}
			}(backupName)
		}
	}

	// Reopen the log file
	if err := rw.openFile(); err != nil {
		return err
	}

	// Update last rotation time
	rw.lastRotate = time.Now()

	// Clean up old backups
	if rw.maxBackups > 0 {
		go rw.cleanupOldBackups()
	}

	return nil
}

// compressFile compresses a file and removes the original
func compressFile(filename string) error {
	// Open the original file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the compressed file
	compressedName := filename + ".gz"
	compressed, err := os.Create(compressedName)
	if err != nil {
		return err
	}
	defer compressed.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(compressed)
	defer gzipWriter.Close()

	// Copy the file contents to the gzip writer
	_, err = io.Copy(gzipWriter, file)
	if err != nil {
		return err
	}

	// Close both writers before removing the original
	gzipWriter.Close()
	compressed.Close()
	file.Close()

	// Remove the original file
	return os.Remove(filename)
}

// cleanupOldBackups removes old backup files exceeding maxBackups
func (rw *RotateWriter) cleanupOldBackups() {
	// Get the base path without extension
	dir := filepath.Dir(rw.filename)
	base := filepath.Base(rw.filename)

	// Get all backup files
	pattern := filepath.Join(dir, base+".????????-??????*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find backup files: %v\n", err)
		return
	}

	// Add .gz files too
	gzPattern := filepath.Join(dir, base+".????????-??????*.gz")
	gzMatches, err := filepath.Glob(gzPattern)
	if err == nil {
		matches = append(matches, gzMatches...)
	}

	// If we don't have too many backups, nothing to do
	if len(matches) <= rw.maxBackups {
		return
	}

	// Sort the backups by modification time (oldest first)
	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Remove excess backups
	for i := 0; i < len(matches)-rw.maxBackups; i++ {
		if err := os.Remove(matches[i]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove old backup: %v\n", err)
		}
	}
}

// ForceRotate forces an immediate log rotation regardless of size or time
func (rw *RotateWriter) ForceRotate() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.rotate()
}

// WithRotateWriter creates a log output option using the rotate writer
// This configures the logger to write to a file with automatic rotation based on
// size and/or time, with backup management.
//
// Example usage:
//
//	logger := dy.New(
//	    dy.WithRotateWriter("logs/app.log",
//	        dy.WithMaxSize(10),             // 10MB
//	        dy.WithMaxBackups(5),           // Keep 5 backups
//	        dy.WithBackupInterval(24*time.Hour), // Daily rotation
//	    ),
//	)
//	defer logger.Close()  // Important!
func WithRotateWriter(filename string, options ...RotateOption) Option {
	return func(l *Logger) {
		rotateWriter, err := NewRotateWriter(filename, options...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create rotate writer: %v, falling back to stderr\n", err)
			l.out = os.Stderr
			return
		}
		l.out = rotateWriter
		// Set up the closer function to properly close the file when Logger.Close() is called
		l.closer = func() error {
			return rotateWriter.Close()
		}
	}
}
