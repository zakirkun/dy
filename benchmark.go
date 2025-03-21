package dy

import (
	"io"
	"testing"
)

func BenchmarkLoggerInfo(b *testing.B) {
	l := New(WithOutput(io.Discard), WithTimestamp(true))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("This is a benchmark test message")
	}
}

func BenchmarkLoggerInfoFormatted(b *testing.B) {
	l := New(WithOutput(io.Discard), WithTimestamp(true))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("This is a benchmark test message: %d", i)
	}
}

func BenchmarkLoggerInfoFilteredOut(b *testing.B) {
	l := New(WithOutput(io.Discard), WithLevel(ErrorLevel))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("This message should be filtered out")
	}
}

func BenchmarkLoggerWithPrefix(b *testing.B) {
	l := New(WithOutput(io.Discard), WithPrefix("BENCHMARK"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("Prefixed benchmark message")
	}
}

func BenchmarkLoggerWithoutTimestamp(b *testing.B) {
	l := New(WithOutput(io.Discard), WithTimestamp(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("No timestamp benchmark message")
	}
}

func BenchmarkDefaultLogger(b *testing.B) {
	// Save the original default logger to restore it after the test
	originalLogger := DefaultLogger
	DefaultLogger = New(WithOutput(io.Discard))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Default logger benchmark message")
	}

	// Restore the original default logger
	DefaultLogger = originalLogger
}

func BenchmarkConcurrentLogging(b *testing.B) {
	l := New(WithOutput(io.Discard))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("Concurrent benchmark message")
		}
	})
}
