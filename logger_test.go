package dy

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{Level(99), "UNKNOWN"},
	}

	for _, test := range tests {
		if got := test.level.String(); got != test.expected {
			t.Errorf("Level(%d).String() = %q, want %q", test.level, got, test.expected)
		}
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	l.Info("test message")
	expected := "[INFO] test message\n"
	if got := buf.String(); got != expected {
		t.Errorf("Logger.Info() output = %q, want %q", got, expected)
	}
}

func TestLoggerLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithLevel(WarnLevel), WithTimestamp(false))

	// These should not output anything due to level filtering
	l.Debug("debug message")
	l.Info("info message")

	if buf.Len() > 0 {
		t.Errorf("Expected no output for messages below WarnLevel, got %q", buf.String())
	}

	// This should output
	l.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Errorf("Expected warn message in output, got %q", buf.String())
	}
}

func TestLoggerPrefix(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithPrefix("TEST"), WithTimestamp(false))

	l.Info("test message")
	expected := "TEST [INFO] test message\n"
	if got := buf.String(); got != expected {
		t.Errorf("Logger.Info() with prefix output = %q, want %q", got, expected)
	}
}

func TestLoggerTimestamp(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(true))

	l.Info("test message")

	// Check that the output contains a timestamp in the expected format
	got := buf.String()
	if !strings.Contains(got, time.Now().Format("2006-01-02")) {
		t.Errorf("Expected output to contain today's date, got %q", got)
	}
}

func TestLoggerFormatting(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	l.Info("Hello, %s!", "world")
	expected := "[INFO] Hello, world!\n"
	if got := buf.String(); got != expected {
		t.Errorf("Logger.Info() with formatting output = %q, want %q", got, expected)
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithLevel(InfoLevel), WithTimestamp(false))

	// This should not output
	l.Debug("debug message")
	if buf.Len() > 0 {
		t.Errorf("Expected no output for Debug message, got %q", buf.String())
	}

	// Change level
	l.SetLevel(DebugLevel)

	// Now this should output
	l.Debug("debug message after level change")
	if !strings.Contains(buf.String(), "debug message after level change") {
		t.Errorf("Expected debug message in output after level change, got %q", buf.String())
	}
}

// This test is a bit tricky as Fatal calls os.Exit
// In a real test, you might want to use a custom ExitFunc that can be overridden
func TestLoggerErrorOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	l.Error("error: %v", "test error")
	expected := "[ERROR] error: test error\n"
	if got := buf.String(); got != expected {
		t.Errorf("Logger.Error() output = %q, want %q", got, expected)
	}
}

func TestDefaultLogger(t *testing.T) {
	// Save the original default logger to restore it after the test
	originalLogger := DefaultLogger

	var buf bytes.Buffer
	DefaultLogger = New(WithOutput(&buf), WithTimestamp(false))

	Info("default logger test")
	expected := "[INFO] default logger test\n"
	if got := buf.String(); got != expected {
		t.Errorf("Info() with default logger output = %q, want %q", got, expected)
	}

	// Restore the original default logger
	DefaultLogger = originalLogger
}

func TestConcurrentAccess(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false))

	// This is a very basic test for concurrent access
	// A more thorough test would use sync.WaitGroup to ensure completion
	for i := 0; i < 10; i++ {
		go func(n int) {
			l.Info("concurrent message %d", n)
		}(i)
	}

	// Give goroutines time to complete
	time.Sleep(100 * time.Millisecond)

	// Check that there was some output
	if buf.Len() == 0 {
		t.Error("Expected some output from concurrent logging")
	}
}

func TestTraceFunctionDisabled(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false), WithTrace(false))

	// Call a function with tracing
	func() {
		defer l.TraceFunction()()
		l.Info("Inside function")
	}()

	// Should only see the Info log, not the trace logs
	output := buf.String()
	if strings.Contains(output, "Entering") || strings.Contains(output, "Exiting") {
		t.Errorf("Expected no trace logs when tracing is disabled, got: %s", output)
	}
	if !strings.Contains(output, "Inside function") {
		t.Errorf("Expected to see info log, got: %s", output)
	}
}

func TestTraceFunctionEnabled(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithOutput(&buf),
		WithTimestamp(false),
		WithTrace(true),
		WithLevel(DebugLevel),
	)

	// Call a function with tracing
	func() {
		defer l.TraceFunction()()
		l.Info("Inside function")
	}()

	// Should see both trace logs and the info log
	output := buf.String()
	if !strings.Contains(output, "→ Entering") {
		t.Errorf("Expected to see function entry trace, got: %s", output)
	}
	if !strings.Contains(output, "← Exiting") {
		t.Errorf("Expected to see function exit trace, got: %s", output)
	}
	if !strings.Contains(output, "Inside function") {
		t.Errorf("Expected to see info log, got: %s", output)
	}
}

func TestTraceFunctionNesting(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithOutput(&buf),
		WithTimestamp(false),
		WithTrace(true),
		WithLevel(DebugLevel),
	)

	// Create nested function calls with tracing
	func() {
		defer l.TraceFunction()()
		l.Info("Level 1")

		func() {
			defer l.TraceFunction()()
			l.Info("Level 2")

			func() {
				defer l.TraceFunction()()
				l.Info("Level 3")
			}()
		}()
	}()

	// Check output
	output := buf.String()
	lines := strings.Split(output, "\n")

	// Check for proper indentation in logs
	foundLevel2Indent := false
	foundLevel3Indent := false

	for _, line := range lines {
		if strings.Contains(line, "Level 2") {
			if strings.Contains(line, l.indentString) {
				foundLevel2Indent = true
			}
		}
		if strings.Contains(line, "Level 3") {
			if strings.Contains(line, l.indentString+l.indentString) {
				foundLevel3Indent = true
			}
		}
	}

	if !foundLevel2Indent {
		t.Errorf("Expected to see Level 2 log with indentation")
	}
	if !foundLevel3Indent {
		t.Errorf("Expected to see Level 3 log with double indentation")
	}
}

func TestTraceFunctionWithArgs(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithOutput(&buf),
		WithTimestamp(false),
		WithTrace(true),
		WithLevel(DebugLevel),
	)

	// Call a function with args in trace
	func() {
		defer l.TraceFunction("test arg:", "value")()
		l.Info("Inside function")
	}()

	// Should see args in entry trace
	output := buf.String()
	if !strings.Contains(output, "test arg: value") {
		t.Errorf("Expected to see function args in trace, got: %s", output)
	}
}

func TestEnableDisableTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithOutput(&buf),
		WithTimestamp(false),
		WithTrace(false),
		WithLevel(DebugLevel),
	)

	// Initially disabled
	func() {
		defer l.TraceFunction()()
		l.Info("First call")
	}()

	firstOutput := buf.String()
	buf.Reset()

	// Enable tracing
	l.EnableTrace()

	func() {
		defer l.TraceFunction()()
		l.Info("Second call")
	}()

	secondOutput := buf.String()
	buf.Reset()

	// Disable tracing again
	l.DisableTrace()

	func() {
		defer l.TraceFunction()()
		l.Info("Third call")
	}()

	thirdOutput := buf.String()

	// Check outputs
	if strings.Contains(firstOutput, "Entering") {
		t.Errorf("Expected no trace logs initially, got: %s", firstOutput)
	}

	if !strings.Contains(secondOutput, "Entering") {
		t.Errorf("Expected trace logs after enabling, got: %s", secondOutput)
	}

	if strings.Contains(thirdOutput, "Entering") {
		t.Errorf("Expected no trace logs after disabling, got: %s", thirdOutput)
	}
}
