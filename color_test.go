package dy

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestColorization(t *testing.T) {
	// Test with colors enabled but output to a bytes.Buffer (non-terminal)
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false), WithColor(true))

	l.Info("test message")
	if strings.Contains(buf.String(), Green) {
		t.Error("Colors should not be applied to non-terminal output")
	}

	// Test with colors explicitly disabled
	buf.Reset()
	l = New(WithOutput(&buf), WithTimestamp(false), WithColor(false))
	l.Info("test message")
	if strings.Contains(buf.String(), Green) {
		t.Error("Colors should not be applied when disabled")
	}

	// Test colorizeLevel function
	l = New(WithColor(true))
	colored := l.colorizeLevel(InfoLevel)
	uncolored := InfoLevel.String()

	if colored == uncolored {
		t.Error("Expected colored level string to be different from uncolored")
	}

	// Test all levels produce different colors
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
	colors := make(map[string]bool)

	for _, level := range levels {
		color := getLevelColor(level)
		colors[color] = true
	}

	if len(colors) != len(levels) {
		t.Error("Expected each level to have a unique color")
	}
}

func TestIsTerminal(t *testing.T) {
	// Test with os.Stdout
	if !isTerminal(os.Stdout) {
		t.Error("os.Stdout should be detected as a terminal")
	}

	// Test with a buffer
	var buf bytes.Buffer
	if isTerminal(&buf) {
		t.Error("bytes.Buffer should not be detected as a terminal")
	}
}
