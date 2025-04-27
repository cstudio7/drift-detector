package logger

import (
	"log"
	"os"
)

// Logger defines the interface for logging.
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// StdLogger is a standard implementation of the Logger interface.
type StdLogger struct {
	logger *log.Logger
}

// NewStdLogger creates a new StdLogger.
func NewStdLogger() *StdLogger {
	return &StdLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// NewTestLogger creates a new StdLogger for testing purposes.
func NewTestLogger() *StdLogger {
	// For testing, we can use the same StdLogger implementation
	// Alternatively, we could create a no-op logger or a logger that writes to a buffer
	return &StdLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs an info message.
func (l *StdLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[INFO] %s %v", msg, keysAndValues)
}

// Warn logs a warning message.
func (l *StdLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[WARN] %s %v", msg, keysAndValues)
}

// Error logs an error message.
func (l *StdLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("[ERROR] %s %v", msg, keysAndValues)
}
