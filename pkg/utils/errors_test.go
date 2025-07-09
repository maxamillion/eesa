package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAppError(ErrorCodeAPITimeout, "API timeout occurred", cause)
	
	assert.Equal(t, ErrorCodeAPITimeout, err.Code)
	assert.Equal(t, "API timeout occurred", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.True(t, err.Retryable)
	assert.NotEmpty(t, err.Context.File)
	assert.NotZero(t, err.Context.Line)
}

func TestNewAppErrorWithContext(t *testing.T) {
	cause := errors.New("underlying error")
	context := ErrorContext{
		Operation: "test_operation",
		Service:   "test_service",
		UserID:    "user123",
		RequestID: "req456",
		Extra: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}
	
	err := NewAppErrorWithContext(ErrorCodeJiraError, "Jira API error", cause, context)
	
	assert.Equal(t, ErrorCodeJiraError, err.Code)
	assert.Equal(t, "Jira API error", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.False(t, err.Retryable) // Jira errors are not retryable by default
	assert.Equal(t, "test_operation", err.Context.Operation)
	assert.Equal(t, "test_service", err.Context.Service)
	assert.Equal(t, "user123", err.Context.UserID)
	assert.Equal(t, "req456", err.Context.RequestID)
	assert.Equal(t, "value1", err.Context.Extra["key1"])
	assert.Equal(t, 42, err.Context.Extra["key2"])
	assert.NotEmpty(t, err.Context.File)
	assert.NotZero(t, err.Context.Line)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "error with cause",
			err: &AppError{
				Code:    ErrorCodeAPITimeout,
				Message: "API timeout",
				Cause:   errors.New("connection timeout"),
			},
			expected: "API_TIMEOUT: API timeout (caused by: connection timeout)",
		},
		{
			name: "error without cause",
			err: &AppError{
				Code:    ErrorCodeDataInvalid,
				Message: "Invalid data format",
			},
			expected: "DATA_INVALID: Invalid data format",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAppError_ChainedMethods(t *testing.T) {
	err := NewAppError(ErrorCodeAPITimeout, "API timeout", nil).
		WithOperation("get_user_activities").
		WithService("jira").
		WithUserID("user123").
		WithRequestID("req456").
		WithExtra("endpoint", "/rest/api/2/search").
		WithExtra("retry_count", 3)
	
	assert.Equal(t, "get_user_activities", err.Context.Operation)
	assert.Equal(t, "jira", err.Context.Service)
	assert.Equal(t, "user123", err.Context.UserID)
	assert.Equal(t, "req456", err.Context.RequestID)
	assert.Equal(t, "/rest/api/2/search", err.Context.Extra["endpoint"])
	assert.Equal(t, 3, err.Context.Extra["retry_count"])
}

func TestAppError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected bool
	}{
		{"API timeout", ErrorCodeAPITimeout, true},
		{"API rate limit", ErrorCodeAPIRateLimit, true},
		{"API server error", ErrorCodeAPIServerError, true},
		{"Network error", ErrorCodeNetworkError, true},
		{"Timeout error", ErrorCodeTimeoutError, true},
		{"Auth failed", ErrorCodeAuthFailed, false},
		{"Data invalid", ErrorCodeDataInvalid, false},
		{"Config invalid", ErrorCodeConfigInvalid, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError(tt.code, "test message", nil)
			assert.Equal(t, tt.expected, err.IsRetryable())
		})
	}
}

func TestAppError_IsType(t *testing.T) {
	err := NewAppError(ErrorCodeAPITimeout, "API timeout", nil)
	
	assert.True(t, err.IsType(ErrorCodeAPITimeout))
	assert.False(t, err.IsType(ErrorCodeAuthFailed))
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAppError(ErrorCodeAPITimeout, "API timeout", cause)
	
	assert.Equal(t, cause, err.Unwrap())
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		originalErr  error
		code         ErrorCode
		message      string
		expectNil    bool
		expectAppErr bool
	}{
		{
			name:      "nil error",
			originalErr: nil,
			code:      ErrorCodeAPITimeout,
			message:   "test message",
			expectNil: true,
		},
		{
			name:         "generic error",
			originalErr:  errors.New("generic error"),
			code:         ErrorCodeAPITimeout,
			message:      "wrapped error",
			expectAppErr: true,
		},
		{
			name:         "existing AppError",
			originalErr:  NewAppError(ErrorCodeDataInvalid, "data error", nil),
			code:         ErrorCodeAPITimeout,
			message:      "wrapped error",
			expectAppErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.originalErr, tt.code, tt.message)
			
			if tt.expectNil {
				assert.Nil(t, result)
			} else if tt.expectAppErr {
				require.NotNil(t, result)
				assert.Equal(t, tt.code, result.Code)
				assert.Equal(t, tt.message, result.Message)
				assert.Equal(t, tt.originalErr, result.Cause)
			}
		})
	}
}

func TestErrorHandler_Handle(t *testing.T) {
	mockLogger := NewMockLogger()
	handler := NewErrorHandler(mockLogger)
	
	// Test handling AppError
	appErr := NewAppError(ErrorCodeAPITimeout, "API timeout", errors.New("connection timeout")).
		WithService("jira").
		WithUserID("user123")
	
	handler.Handle(appErr, "test_operation")
	
	assert.Len(t, mockLogger.Entries, 1)
	assert.Equal(t, LogLevelError, mockLogger.Entries[0].Level)
	assert.Equal(t, "API timeout", mockLogger.Entries[0].Message)
	assert.Equal(t, appErr.Cause, mockLogger.Entries[0].Error)
	assert.True(t, mockLogger.HasFieldValue("error_code", "API_TIMEOUT"))
	assert.True(t, mockLogger.HasFieldValue("operation", "test_operation"))
	assert.True(t, mockLogger.HasFieldValue("service", "jira"))
	assert.True(t, mockLogger.HasFieldValue("user_id", "user123"))
	assert.True(t, mockLogger.HasFieldValue("retryable", true))
	
	// Test handling generic error
	mockLogger.Reset()
	genericErr := errors.New("generic error")
	handler.Handle(genericErr, "test_operation")
	
	assert.Len(t, mockLogger.Entries, 1)
	assert.Equal(t, LogLevelError, mockLogger.Entries[0].Level)
	assert.Equal(t, "Unhandled error", mockLogger.Entries[0].Message)
	assert.Equal(t, genericErr, mockLogger.Entries[0].Error)
	assert.True(t, mockLogger.HasFieldValue("error_code", "UNKNOWN"))
	assert.True(t, mockLogger.HasFieldValue("operation", "test_operation"))
	
	// Test handling nil error
	mockLogger.Reset()
	handler.Handle(nil, "test_operation")
	assert.Len(t, mockLogger.Entries, 0)
}

func TestRecoverFromPanic(t *testing.T) {
	mockLogger := NewMockLogger()
	
	// Test panic recovery
	func() {
		defer RecoverFromPanic(mockLogger)
		panic("test panic")
	}()
	
	assert.Len(t, mockLogger.Entries, 1)
	assert.Equal(t, LogLevelError, mockLogger.Entries[0].Level)
	assert.Equal(t, "Panic recovered", mockLogger.Entries[0].Message)
	assert.NotNil(t, mockLogger.Entries[0].Error)
	assert.True(t, mockLogger.HasFieldValue("error_code", "INTERNAL_ERROR"))
	assert.True(t, mockLogger.HasFieldValue("panic_value", "test panic"))
}

func TestValidationErrors(t *testing.T) {
	var validationErrors ValidationErrors
	
	// Test empty validation errors
	assert.False(t, validationErrors.HasErrors())
	assert.Equal(t, "validation failed", validationErrors.Error())
	
	// Add validation errors
	validationErrors.Add("username", "Username is required", "")
	validationErrors.Add("email", "Invalid email format", "invalid-email")
	validationErrors.Add("username", "Username too short", "ab")
	
	assert.True(t, validationErrors.HasErrors())
	assert.Len(t, validationErrors, 3)
	
	// Test error message
	assert.Equal(t, "validation failed with 3 errors", validationErrors.Error())
	
	// Test single error message
	singleError := ValidationErrors{
		{Field: "username", Message: "Username is required", Value: ""},
	}
	assert.Equal(t, "validation failed: Username is required", singleError.Error())
	
	// Test GetFieldErrors
	usernameErrors := validationErrors.GetFieldErrors("username")
	assert.Len(t, usernameErrors, 2)
	assert.Equal(t, "Username is required", usernameErrors[0].Message)
	assert.Equal(t, "Username too short", usernameErrors[1].Message)
	
	emailErrors := validationErrors.GetFieldErrors("email")
	assert.Len(t, emailErrors, 1)
	assert.Equal(t, "Invalid email format", emailErrors[0].Message)
	
	nonExistentErrors := validationErrors.GetFieldErrors("nonexistent")
	assert.Len(t, nonExistentErrors, 0)
}

func TestValidationErrors_ToAppError(t *testing.T) {
	var validationErrors ValidationErrors
	
	// Test empty validation errors
	appErr := validationErrors.ToAppError()
	assert.Nil(t, appErr)
	
	// Add validation errors
	validationErrors.Add("username", "Username is required", "")
	validationErrors.Add("email", "Invalid email format", "invalid-email")
	
	appErr = validationErrors.ToAppError()
	require.NotNil(t, appErr)
	assert.Equal(t, ErrorCodeValidationError, appErr.Code)
	assert.Equal(t, "validation failed with 2 errors", appErr.Message)
	assert.Equal(t, validationErrors, appErr.Context.Extra["validation_errors"])
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected bool
	}{
		{ErrorCodeAPITimeout, true},
		{ErrorCodeAPIRateLimit, true},
		{ErrorCodeAPIServerError, true},
		{ErrorCodeNetworkError, true},
		{ErrorCodeTimeoutError, true},
		{ErrorCodeAuthFailed, false},
		{ErrorCodeDataInvalid, false},
		{ErrorCodeConfigInvalid, false},
		{ErrorCodeUnknown, false},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.expected, isRetryable(tt.code))
		})
	}
}