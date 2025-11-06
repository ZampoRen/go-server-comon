package logger

import (
	"log"
)

// Logger wraps logging functionality
type Logger struct {
	// TODO: Add logger implementation
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	log.Printf("[INFO] %s", msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error) {
	log.Printf("[ERROR] %s: %v", msg, err)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	log.Printf("[DEBUG] %s", msg)
}
