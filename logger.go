package dy

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the logging level
type Level int

const (
	// DebugLevel logs detailed information for debugging
	DebugLevel Level = iota
	// InfoLevel logs informational messages
	InfoLevel
	// WarnLevel logs warnings that might need attention
	WarnLevel
	// ErrorLevel logs error conditions
	ErrorLevel
	// FatalLevel logs critical errors and exits
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry for JSON output
type LogEntry struct {
	Timestamp   string      `json:"timestamp,omitempty"`
	Level       string      `json:"level"`
	Message     string      `json:"message"`
	Prefix      string      `json:"prefix,omitempty"`
	NestLevel   int         `json:"nest_level,omitempty"`
	Caller      *CallerInfo `json:"caller,omitempty"`
	TraceType   string      `json:"trace_type,omitempty"` // "entry" or "exit" for trace logs
	ElapsedTime string      `json:"elapsed_time,omitempty"`
}

// CallerInfo contains information about the caller of the log function
type CallerInfo struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Logger represents a logger with configurable outputs and level
type Logger struct {
	mu           sync.Mutex
	out          io.Writer
	level        Level
	prefix       string
	timestamp    bool
	nestingLevel int
	traceEnabled bool
	indentString string
	jsonFormat   bool
	callerInfo   bool
}

// Option is a function that modifies a Logger
type Option func(*Logger)

// WithOutput sets the output writer
func WithOutput(w io.Writer) Option {
	return func(l *Logger) {
		l.out = w
	}
}

// WithLevel sets the minimum log level
func WithLevel(level Level) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// WithPrefix sets a prefix for all log messages
func WithPrefix(prefix string) Option {
	return func(l *Logger) {
		l.prefix = prefix
	}
}

// WithTimestamp enables or disables timestamp in log messages
func WithTimestamp(enable bool) Option {
	return func(l *Logger) {
		l.timestamp = enable
	}
}

// WithTrace enables or disables function call tracing
func WithTrace(enable bool) Option {
	return func(l *Logger) {
		l.traceEnabled = enable
	}
}

// WithIndentString sets the string used for indentation in nested function logs
func WithIndentString(indent string) Option {
	return func(l *Logger) {
		l.indentString = indent
	}
}

// WithJSONFormat enables or disables JSON output format
func WithJSONFormat(enable bool) Option {
	return func(l *Logger) {
		l.jsonFormat = enable
	}
}

// WithCallerInfo enables or disables including caller information (file, line, function)
func WithCallerInfo(enable bool) Option {
	return func(l *Logger) {
		l.callerInfo = enable
	}
}

// New creates a new Logger with the given options
func New(options ...Option) *Logger {
	l := &Logger{
		out:          os.Stdout,
		level:        InfoLevel,
		timestamp:    true,
		nestingLevel: 0,
		traceEnabled: false,
		indentString: "  ",  // Default to two spaces
		jsonFormat:   false, // Default to text format
		callerInfo:   false, // Default to no caller info
	}

	for _, option := range options {
		option(l)
	}

	return l
}

// DefaultLogger is the default logger used by package-level functions
var DefaultLogger = New()

// log writes a log message if the level is sufficient
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Format the message
	msg := fmt.Sprintf(format, args...)

	// Acquire lock only for reading state
	l.mu.Lock()
	nestingLevel := l.nestingLevel
	hasPrefix := l.prefix != ""
	prefixValue := l.prefix
	hasTimestamp := l.timestamp
	traceEnabled := l.traceEnabled
	indentStr := l.indentString
	useJSON := l.jsonFormat
	includeCaller := l.callerInfo
	out := l.out // Keep a reference to output
	l.mu.Unlock()

	// Get caller info if enabled
	var caller *CallerInfo
	if includeCaller {
		caller = getCaller(3) // skip log, calling method, and actual caller
	}

	// Current time for timestamp
	now := time.Now()
	timestampStr := now.Format("2006-01-02 15:04:05.000")

	if useJSON {
		// Create a structured log entry
		entry := LogEntry{
			Level:     level.String(),
			Message:   msg,
			NestLevel: nestingLevel,
		}

		if hasTimestamp {
			entry.Timestamp = timestampStr
		}

		if hasPrefix {
			entry.Prefix = prefixValue
		}

		if includeCaller {
			entry.Caller = caller
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(entry)
		if err != nil {
			// Fallback to plain text if JSON marshaling fails
			fmt.Fprintf(out, "ERROR marshaling log entry to JSON: %v\n", err)
		} else {
			fmt.Fprintln(out, string(jsonData))
		}
	} else {
		// Original text format
		var prefix string
		if hasPrefix {
			prefix = prefixValue + " "
		}

		var timestamp string
		if hasTimestamp {
			timestamp = timestampStr + " "
		}

		var indent string
		if traceEnabled && nestingLevel > 0 {
			indent = strings.Repeat(indentStr, nestingLevel)
		}

		// Add caller info if enabled
		var callerInfo string
		if includeCaller && caller != nil {
			callerInfo = fmt.Sprintf(" [%s:%d %s] ", caller.File, caller.Line, caller.Function)
		}

		fmt.Fprintf(out, "%s%s[%s]%s %s%s\n", timestamp, prefix, level.String(), callerInfo, indent, msg)
	}

	if level == FatalLevel {
		os.Exit(1)
	}
}

// getCaller returns information about the calling function
func getCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return &CallerInfo{
			Function: "unknown",
			File:     "unknown",
			Line:     0,
		}
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	funcName := fn.Name()

	// Shorten the file path to just filename
	fileName := filepath.Base(file)

	return &CallerInfo{
		Function: funcName,
		File:     fileName,
		Line:     line,
	}
}

// getFunctionName returns the name of the calling function
func getFunctionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}

	// Get full function name (including package path)
	fullName := runtime.FuncForPC(pc).Name()

	// Extract just the function name (without package path)
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// TraceFunction logs entry and exit of a function with proper nesting
// It returns a function that should be deferred to log the exit
func (l *Logger) TraceFunction(args ...interface{}) func() {
	if !l.traceEnabled || DebugLevel < l.level {
		return func() {}
	}

	// Get calling function name and location
	funcName := getFunctionName(2) // skip TraceFunction and caller
	var caller *CallerInfo
	if l.callerInfo {
		caller = getCaller(2)
	}

	// Prepare the entry message outside the lock
	var entryMsg string
	if len(args) == 0 {
		entryMsg = fmt.Sprintf("→ Entering %s", funcName)
	} else {
		// Convert all arguments to a simple string
		argsStr := fmt.Sprint(args...)
		entryMsg = fmt.Sprintf("→ Entering %s %s", funcName, argsStr)
	}

	// Lock only for the minimal necessary operations
	l.mu.Lock()
	currentLevel := l.nestingLevel
	l.nestingLevel++
	useJSON := l.jsonFormat
	includeCaller := l.callerInfo
	hasPrefix := l.prefix != ""
	prefixValue := l.prefix
	hasTimestamp := l.timestamp
	traceEnabled := l.traceEnabled
	indentStr := l.indentString
	out := l.out
	l.mu.Unlock()

	// Record start time for elapsed time calculation
	startTime := time.Now()
	timestampStr := startTime.Format("2006-01-02 15:04:05.000")

	// Log after releasing the lock to avoid potential deadlock
	if DebugLevel >= l.level {
		if useJSON {
			// Create a structured log entry
			entry := LogEntry{
				Level:     DebugLevel.String(),
				Message:   entryMsg,
				NestLevel: currentLevel,
				TraceType: "entry",
			}

			if hasTimestamp {
				entry.Timestamp = timestampStr
			}

			if hasPrefix {
				entry.Prefix = prefixValue
			}

			if includeCaller && caller != nil {
				entry.Caller = caller
			}

			// Marshal to JSON
			jsonData, err := json.Marshal(entry)
			if err != nil {
				// Fallback to plain text if JSON marshaling fails
				fmt.Fprintf(out, "ERROR marshaling trace entry to JSON: %v\n", err)
			} else {
				fmt.Fprintln(out, string(jsonData))
			}
		} else {
			// Original text format
			indent := ""
			if traceEnabled && currentLevel > 0 {
				indent = strings.Repeat(indentStr, currentLevel)
			}

			var prefix string
			if hasPrefix {
				prefix = prefixValue + " "
			}

			var timestamp string
			if hasTimestamp {
				timestamp = timestampStr + " "
			}

			// Add caller info if enabled
			var callerInfo string
			if includeCaller && caller != nil {
				callerInfo = fmt.Sprintf(" [%s:%d %s] ", caller.File, caller.Line, caller.Function)
			}

			fmt.Fprintf(out, "%s%s[%s]%s %s%s\n", timestamp, prefix, DebugLevel.String(), callerInfo, indent, entryMsg)
		}
	}

	// Return function to be deferred
	return func() {
		exitMsg := fmt.Sprintf("← Exiting %s", funcName)
		endTime := time.Now()
		elapsed := endTime.Sub(startTime)
		elapsedStr := elapsed.String()

		// Lock only for the minimal necessary operations
		l.mu.Lock()
		currentLevel := l.nestingLevel - 1
		if currentLevel < 0 {
			currentLevel = 0
		}
		l.nestingLevel = currentLevel
		useJSON := l.jsonFormat
		includeCaller := l.callerInfo
		hasPrefix := l.prefix != ""
		prefixValue := l.prefix
		hasTimestamp := l.timestamp
		traceEnabled := l.traceEnabled
		indentStr := l.indentString
		out := l.out
		l.mu.Unlock()

		// Log after releasing the lock
		if DebugLevel >= l.level {
			timestampStr := endTime.Format("2006-01-02 15:04:05.000")

			if useJSON {
				// Get updated caller info for exit
				var exitCaller *CallerInfo
				if includeCaller {
					exitCaller = getCaller(2)
				}

				// Create a structured log entry
				entry := LogEntry{
					Level:       DebugLevel.String(),
					Message:     exitMsg,
					NestLevel:   currentLevel,
					TraceType:   "exit",
					ElapsedTime: elapsedStr,
				}

				if hasTimestamp {
					entry.Timestamp = timestampStr
				}

				if hasPrefix {
					entry.Prefix = prefixValue
				}

				if includeCaller && exitCaller != nil {
					entry.Caller = exitCaller
				}

				// Marshal to JSON
				jsonData, err := json.Marshal(entry)
				if err != nil {
					// Fallback to plain text if JSON marshaling fails
					fmt.Fprintf(out, "ERROR marshaling trace exit to JSON: %v\n", err)
				} else {
					fmt.Fprintln(out, string(jsonData))
				}
			} else {
				// Original text format
				indent := ""
				if traceEnabled && currentLevel > 0 {
					indent = strings.Repeat(indentStr, currentLevel)
				}

				var prefix string
				if hasPrefix {
					prefix = prefixValue + " "
				}

				var timestamp string
				if hasTimestamp {
					timestamp = timestampStr + " "
				}

				// Add caller info and elapsed time for exit
				var exitInfo string
				if includeCaller {
					caller := getCaller(2)
					exitInfo = fmt.Sprintf(" [%s:%d %s] (took %s) ", caller.File, caller.Line, caller.Function, elapsedStr)
				} else {
					exitInfo = fmt.Sprintf(" (took %s) ", elapsedStr)
				}

				fmt.Fprintf(out, "%s%s[%s]%s %s%s\n", timestamp, prefix, DebugLevel.String(), exitInfo, indent, exitMsg)
			}
		}
	}
}

// EnableTrace enables function call tracing
func (l *Logger) EnableTrace() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.traceEnabled = true
}

// DisableTrace disables function call tracing
func (l *Logger) DisableTrace() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.traceEnabled = false
}

// EnableJSONFormat enables JSON output format
func (l *Logger) EnableJSONFormat() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.jsonFormat = true
}

// DisableJSONFormat disables JSON output format (switches to text format)
func (l *Logger) DisableJSONFormat() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.jsonFormat = false
}

// EnableCallerInfo enables including caller information in logs
func (l *Logger) EnableCallerInfo() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callerInfo = true
}

// DisableCallerInfo disables including caller information in logs
func (l *Logger) DisableCallerInfo() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callerInfo = false
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

// Info logs an informational message using the default logger
func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(format string, args ...interface{}) {
	DefaultLogger.Fatal(format, args...)
}

// SetLevel sets the minimum log level for the default logger
func SetLevel(level Level) {
	DefaultLogger.SetLevel(level)
}

// TraceFunction logs entry and exit of a function with proper nesting using the default logger
// Example usage: defer logger.TraceFunction("param:", value)()
func TraceFunction(args ...interface{}) func() {
	return DefaultLogger.TraceFunction(args...)
}

// EnableTrace enables function call tracing for the default logger
func EnableTrace() {
	DefaultLogger.EnableTrace()
}

// DisableTrace disables function call tracing for the default logger
func DisableTrace() {
	DefaultLogger.DisableTrace()
}

// EnableJSONFormat enables JSON output format for the default logger
func EnableJSONFormat() {
	DefaultLogger.EnableJSONFormat()
}

// DisableJSONFormat disables JSON output format for the default logger
func DisableJSONFormat() {
	DefaultLogger.DisableJSONFormat()
}

// EnableCallerInfo enables including caller information in logs for the default logger
func EnableCallerInfo() {
	DefaultLogger.EnableCallerInfo()
}

// DisableCallerInfo disables including caller information in logs for the default logger
func DisableCallerInfo() {
	DefaultLogger.DisableCallerInfo()
}
