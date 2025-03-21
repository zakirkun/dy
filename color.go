package dy

import (
	"fmt"
	"io"
	"os"
)

// ANSI color codes
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Red        = "\033[31m"
	Green      = "\033[32m"
	Yellow     = "\033[33m"
	Blue       = "\033[34m"
	Magenta    = "\033[35m"
	Cyan       = "\033[36m"
	White      = "\033[37m"
	BoldRed    = "\033[1;31m"
	BoldGreen  = "\033[1;32m"
	BoldYellow = "\033[1;33m"
	BoldBlue   = "\033[1;34m"
	BoldPurple = "\033[1;35m"
	BoldCyan   = "\033[1;36m"
	BoldWhite  = "\033[1;37m"
)

// ColorOption is a function that modifies a Logger to use colors
func WithColor(enable bool) Option {
	return func(l *Logger) {
		l.colorEnabled = enable
	}
}

// Returns the appropriate color code for a log level
func getLevelColor(level Level) string {
	switch level {
	case DebugLevel:
		return Blue
	case InfoLevel:
		return Green
	case WarnLevel:
		return Yellow
	case ErrorLevel:
		return Red
	case FatalLevel:
		return BoldRed
	default:
		return Reset
	}
}

// colorizeLevel returns a colorized level string if colors are enabled
func (l *Logger) colorizeLevel(level Level) string {
	if !l.colorEnabled || !isTerminal(l.out) {
		return level.String()
	}

	return fmt.Sprintf("%s%s%s", getLevelColor(level), level.String(), Reset)
}

// isTerminal checks if the writer is a terminal (to avoid adding color codes to files, etc.)
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return f == os.Stdout || f == os.Stderr
	}
	return false
}
