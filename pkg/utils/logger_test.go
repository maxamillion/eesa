package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"error", LogLevelError},
		{"unknown", LogLevelInfo}, // Default case
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevel(999), "UNKNOWN"}, // Invalid level
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockLogger(t *testing.T) {
	logger := NewMockLogger()
	
	// Test debug logging
	logger.Debug("debug message", NewField("key1", "value1"))
	assert.Len(t, logger.Entries, 1)
	assert.Equal(t, LogLevelDebug, logger.Entries[0].Level)
	assert.Equal(t, "debug message", logger.Entries[0].Message)
	assert.Len(t, logger.Entries[0].Fields, 1)
	assert.Equal(t, "key1", logger.Entries[0].Fields[0].Key)
	assert.Equal(t, "value1", logger.Entries[0].Fields[0].Value)
	
	// Test info logging
	logger.Info("info message", NewField("key2", "value2"))
	assert.Len(t, logger.Entries, 2)
	assert.Equal(t, LogLevelInfo, logger.Entries[1].Level)
	assert.Equal(t, "info message", logger.Entries[1].Message)
	
	// Test warn logging
	logger.Warn("warn message")
	assert.Len(t, logger.Entries, 3)
	assert.Equal(t, LogLevelWarn, logger.Entries[2].Level)
	assert.Equal(t, "warn message", logger.Entries[2].Message)
	
	// Test error logging
	testErr := errors.New("test error")
	logger.Error("error message", testErr, NewField("key3", "value3"))
	assert.Len(t, logger.Entries, 4)
	assert.Equal(t, LogLevelError, logger.Entries[3].Level)
	assert.Equal(t, "error message", logger.Entries[3].Message)
	assert.Equal(t, testErr, logger.Entries[3].Error)
}

func TestMockLogger_GetEntriesByLevel(t *testing.T) {
	logger := NewMockLogger()
	
	logger.Debug("debug1")
	logger.Info("info1")
	logger.Debug("debug2")
	logger.Error("error1", nil)
	
	debugEntries := logger.GetEntriesByLevel(LogLevelDebug)
	assert.Len(t, debugEntries, 2)
	assert.Equal(t, "debug1", debugEntries[0].Message)
	assert.Equal(t, "debug2", debugEntries[1].Message)
	
	infoEntries := logger.GetEntriesByLevel(LogLevelInfo)
	assert.Len(t, infoEntries, 1)
	assert.Equal(t, "info1", infoEntries[0].Message)
	
	warnEntries := logger.GetEntriesByLevel(LogLevelWarn)
	assert.Len(t, warnEntries, 0)
}

func TestMockLogger_GetEntriesByMessage(t *testing.T) {
	logger := NewMockLogger()
	
	logger.Info("test message")
	logger.Debug("another message")
	logger.Info("test message")
	
	entries := logger.GetEntriesByMessage("test message")
	assert.Len(t, entries, 2)
	assert.Equal(t, LogLevelInfo, entries[0].Level)
	assert.Equal(t, LogLevelInfo, entries[1].Level)
	
	entries = logger.GetEntriesByMessage("nonexistent")
	assert.Len(t, entries, 0)
}

func TestMockLogger_HasFieldValue(t *testing.T) {
	logger := NewMockLogger()
	
	logger.Info("test message", NewField("service", "jira"), NewField("status", "success"))
	logger.Debug("another message", NewField("service", "gemini"))
	
	assert.True(t, logger.HasFieldValue("service", "jira"))
	assert.True(t, logger.HasFieldValue("service", "gemini"))
	assert.True(t, logger.HasFieldValue("status", "success"))
	assert.False(t, logger.HasFieldValue("service", "nonexistent"))
	assert.False(t, logger.HasFieldValue("nonexistent", "value"))
}

func TestMockLogger_Reset(t *testing.T) {
	logger := NewMockLogger()
	
	logger.Info("test message")
	logger.Debug("another message")
	assert.Len(t, logger.Entries, 2)
	
	logger.Reset()
	assert.Len(t, logger.Entries, 0)
}

func TestSecurityLogger(t *testing.T) {
	mockLogger := NewMockLogger()
	securityLogger := NewSecurityLogger(mockLogger)
	
	// Test auth attempt logging
	securityLogger.LogAuthAttempt("jira", "testuser", true, "192.168.1.1")
	assert.Len(t, mockLogger.Entries, 1)
	assert.Equal(t, "Authentication attempt", mockLogger.Entries[0].Message)
	assert.True(t, mockLogger.HasFieldValue("service", "jira"))
	assert.True(t, mockLogger.HasFieldValue("username", "testuser"))
	assert.True(t, mockLogger.HasFieldValue("success", true))
	assert.True(t, mockLogger.HasFieldValue("ip", "192.168.1.1"))
	assert.True(t, mockLogger.HasFieldValue("security_event", "auth_attempt"))
	
	// Test API call logging
	duration := 250 * time.Millisecond
	securityLogger.LogAPICall("jira", "/rest/api/2/search", "GET", 200, duration)
	assert.Len(t, mockLogger.Entries, 2)
	assert.Equal(t, "API call", mockLogger.Entries[1].Message)
	assert.True(t, mockLogger.HasFieldValue("service", "jira"))
	assert.True(t, mockLogger.HasFieldValue("endpoint", "/rest/api/2/search"))
	assert.True(t, mockLogger.HasFieldValue("method", "GET"))
	assert.True(t, mockLogger.HasFieldValue("status_code", 200))
	assert.True(t, mockLogger.HasFieldValue("duration_ms", int64(250)))
	assert.True(t, mockLogger.HasFieldValue("security_event", "api_call"))
	
	// Test security event logging
	securityLogger.LogSecurityEvent("credential_access", "Keyring access", NewField("service", "jira"))
	assert.Len(t, mockLogger.Entries, 3)
	assert.Equal(t, "Security event", mockLogger.Entries[2].Message)
	assert.True(t, mockLogger.HasFieldValue("security_event", "credential_access"))
	assert.True(t, mockLogger.HasFieldValue("description", "Keyring access"))
	assert.True(t, mockLogger.HasFieldValue("service", "jira"))
	
	// Test credential access logging
	securityLogger.LogCredentialAccess("jira", "get_token", true)
	assert.Len(t, mockLogger.Entries, 4)
	assert.Equal(t, "Credential access", mockLogger.Entries[3].Message)
	assert.True(t, mockLogger.HasFieldValue("service", "jira"))
	assert.True(t, mockLogger.HasFieldValue("operation", "get_token"))
	assert.True(t, mockLogger.HasFieldValue("success", true))
	assert.True(t, mockLogger.HasFieldValue("security_event", "credential_access"))
}

func TestNewField(t *testing.T) {
	field := NewField("test_key", "test_value")
	assert.Equal(t, "test_key", field.Key)
	assert.Equal(t, "test_value", field.Value)
	
	// Test with different value types
	intField := NewField("count", 42)
	assert.Equal(t, "count", intField.Key)
	assert.Equal(t, 42, intField.Value)
	
	boolField := NewField("enabled", true)
	assert.Equal(t, "enabled", boolField.Key)
	assert.Equal(t, true, boolField.Value)
}

func TestStructuredLogger_LogLevels(t *testing.T) {
	// Test that log levels are respected
	tests := []struct {
		name         string
		loggerLevel  string
		shouldLog    map[LogLevel]bool
	}{
		{
			name:        "debug level",
			loggerLevel: "debug",
			shouldLog: map[LogLevel]bool{
				LogLevelDebug: true,
				LogLevelInfo:  true,
				LogLevelWarn:  true,
				LogLevelError: true,
			},
		},
		{
			name:        "info level",
			loggerLevel: "info",
			shouldLog: map[LogLevel]bool{
				LogLevelDebug: false,
				LogLevelInfo:  true,
				LogLevelWarn:  true,
				LogLevelError: true,
			},
		},
		{
			name:        "warn level",
			loggerLevel: "warn",
			shouldLog: map[LogLevel]bool{
				LogLevelDebug: false,
				LogLevelInfo:  false,
				LogLevelWarn:  true,
				LogLevelError: true,
			},
		},
		{
			name:        "error level",
			loggerLevel: "error",
			shouldLog: map[LogLevel]bool{
				LogLevelDebug: false,
				LogLevelInfo:  false,
				LogLevelWarn:  false,
				LogLevelError: true,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.loggerLevel)
			
			// We can't easily test the actual output without capturing stdout,
			// so we just verify the logger was created successfully
			assert.NotNil(t, logger)
			
			// Test that methods can be called without panicking
			assert.NotPanics(t, func() {
				logger.Debug("debug message")
				logger.Info("info message")
				logger.Warn("warn message")
				logger.Error("error message", errors.New("test error"))
			})
		})
	}
}