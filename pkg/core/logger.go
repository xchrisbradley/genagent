package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorCyan    = "\033[36m"
	ColorMagenta = "\033[35m"
	ColorRed     = "\033[31m"
)

// Logger provides structured logging for the application
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
	isDebug     bool
}

// NewLogger creates a new logger instance
func NewLogger(logDir string, debug bool) (*Logger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("genagent_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)
	errorWriter := io.MultiWriter(file, os.Stderr)

	return &Logger{
		debugLogger: log.New(multiWriter, ColorMagenta+"DEBUG: "+ColorReset, log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(multiWriter, ColorCyan+"INFO:  "+ColorReset, log.Ldate|log.Ltime),
		errorLogger: log.New(errorWriter, ColorRed+"ERROR: "+ColorReset, log.Ldate|log.Ltime|log.Lshortfile),
		isDebug:     debug,
	}, nil
}

// Debug logs debug messages when debug mode is enabled
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.isDebug {
		l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info logs informational messages
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Output(2, fmt.Sprintf(format, v...))
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Output(2, fmt.Sprintf(format, v...))
}
