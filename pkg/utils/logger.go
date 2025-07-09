package utils

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a log level string
func ParseLogLevel(s string) LogLevel {
	switch s {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// Field represents a structured log field
type Field struct {
	Key   string
	Value interface{}
}

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
}

// StructuredLogger implements the Logger interface with structured logging
type StructuredLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new structured logger
func NewLogger(levelStr string) Logger {
	level := ParseLogLevel(levelStr)
	logger := log.New(os.Stdout, "", 0)
	
	return &StructuredLogger{
		level:  level,
		logger: logger,
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, fields ...Field) {
	if l.level <= LogLevelDebug {
		l.log(LogLevelDebug, msg, nil, fields...)
	}
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, fields ...Field) {
	if l.level <= LogLevelInfo {
		l.log(LogLevelInfo, msg, nil, fields...)
	}
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, fields ...Field) {
	if l.level <= LogLevelWarn {
		l.log(LogLevelWarn, msg, nil, fields...)
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, err error, fields ...Field) {
	if l.level <= LogLevelError {
		l.log(LogLevelError, msg, err, fields...)
	}
}

// log formats and outputs a log message
func (l *StructuredLogger) log(level LogLevel, msg string, err error, fields ...Field) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level.String(),
		"message":   msg,
	}
	
	// Add error if present
	if err != nil {
		logEntry["error"] = err.Error()
	}
	
	// Add fields
	for _, field := range fields {
		// Avoid overwriting standard fields
		if field.Key != "timestamp" && field.Key != "level" && field.Key != "message" && field.Key != "error" {
			logEntry[field.Key] = field.Value
		}
	}
	
	// Marshal to JSON
	jsonData, jsonErr := json.Marshal(logEntry)
	if jsonErr != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.logger.Printf("LOG_ERROR: Failed to marshal log entry: %v", jsonErr)
		l.logger.Printf("%s: %s", level.String(), msg)
		return
	}
	
	l.logger.Println(string(jsonData))
}

// SecurityLogger provides security-specific logging functionality
type SecurityLogger struct {
	logger Logger
}

// NewSecurityLogger creates a new security logger
func NewSecurityLogger(logger Logger) *SecurityLogger {
	return &SecurityLogger{logger: logger}
}

// LogAuthAttempt logs an authentication attempt
func (s *SecurityLogger) LogAuthAttempt(service string, username string, success bool, ip string) {
	s.logger.Info("Authentication attempt",
		NewField("service", service),
		NewField("username", username),
		NewField("success", success),
		NewField("ip", ip),
		NewField("security_event", "auth_attempt"),
	)
}

// LogAPICall logs an API call for security auditing
func (s *SecurityLogger) LogAPICall(service string, endpoint string, method string, statusCode int, duration time.Duration) {
	s.logger.Info("API call",
		NewField("service", service),
		NewField("endpoint", endpoint),
		NewField("method", method),
		NewField("status_code", statusCode),
		NewField("duration_ms", duration.Milliseconds()),
		NewField("security_event", "api_call"),
	)
}

// LogSecurityEvent logs a general security event
func (s *SecurityLogger) LogSecurityEvent(eventType string, description string, fields ...Field) {
	allFields := append([]Field{
		NewField("security_event", eventType),
		NewField("description", description),
	}, fields...)
	
	s.logger.Info("Security event", allFields...)
}

// LogCredentialAccess logs credential access events
func (s *SecurityLogger) LogCredentialAccess(service string, operation string, success bool) {
	s.logger.Info("Credential access",
		NewField("service", service),
		NewField("operation", operation),
		NewField("success", success),
		NewField("security_event", "credential_access"),
	)
}

// MockLogger is a mock implementation of Logger for testing
type MockLogger struct {
	Entries []LogEntry
}

// LogEntry represents a log entry for testing
type LogEntry struct {
	Level   LogLevel
	Message string
	Error   error
	Fields  []Field
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Entries: make([]LogEntry, 0),
	}
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   LogLevelDebug,
		Message: msg,
		Fields:  fields,
	})
}

// Info logs an info message
func (m *MockLogger) Info(msg string, fields ...Field) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   LogLevelInfo,
		Message: msg,
		Fields:  fields,
	})
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   LogLevelWarn,
		Message: msg,
		Fields:  fields,
	})
}

// Error logs an error message
func (m *MockLogger) Error(msg string, err error, fields ...Field) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   LogLevelError,
		Message: msg,
		Error:   err,
		Fields:  fields,
	})
}

// Reset clears all logged entries
func (m *MockLogger) Reset() {
	m.Entries = make([]LogEntry, 0)
}

// GetEntriesByLevel returns all log entries with the specified level
func (m *MockLogger) GetEntriesByLevel(level LogLevel) []LogEntry {
	var entries []LogEntry
	for _, entry := range m.Entries {
		if entry.Level == level {
			entries = append(entries, entry)
		}
	}
	return entries
}

// GetEntriesByMessage returns all log entries containing the specified message
func (m *MockLogger) GetEntriesByMessage(message string) []LogEntry {
	var entries []LogEntry
	for _, entry := range m.Entries {
		if entry.Message == message {
			entries = append(entries, entry)
		}
	}
	return entries
}

// HasFieldValue checks if any log entry has a field with the specified key and value
func (m *MockLogger) HasFieldValue(key string, value interface{}) bool {
	for _, entry := range m.Entries {
		for _, field := range entry.Fields {
			if field.Key == key && field.Value == value {
				return true
			}
		}
	}
	return false
}

// NewField creates a new log field
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}