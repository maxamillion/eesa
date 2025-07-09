package utils

import (
	"fmt"
	"runtime"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	// Configuration errors
	ErrorCodeConfigInvalid    ErrorCode = "CONFIG_INVALID"
	ErrorCodeConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"
	ErrorCodeConfigParseError ErrorCode = "CONFIG_PARSE_ERROR"
	
	// Authentication errors
	ErrorCodeAuthFailed     ErrorCode = "AUTH_FAILED"
	ErrorCodeTokenExpired   ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeTokenInvalid   ErrorCode = "TOKEN_INVALID"
	ErrorCodeCredentialsMissing ErrorCode = "CREDENTIALS_MISSING"
	
	// API errors
	ErrorCodeAPITimeout     ErrorCode = "API_TIMEOUT"
	ErrorCodeAPIRateLimit   ErrorCode = "API_RATE_LIMIT"
	ErrorCodeAPIUnauthorized ErrorCode = "API_UNAUTHORIZED"
	ErrorCodeAPINotFound    ErrorCode = "API_NOT_FOUND"
	ErrorCodeAPIServerError ErrorCode = "API_SERVER_ERROR"
	ErrorCodeAPIBadRequest  ErrorCode = "API_BAD_REQUEST"
	
	// Data processing errors
	ErrorCodeDataInvalid    ErrorCode = "DATA_INVALID"
	ErrorCodeDataMissing    ErrorCode = "DATA_MISSING"
	ErrorCodeDataCorrupted  ErrorCode = "DATA_CORRUPTED"
	ErrorCodeParseError     ErrorCode = "PARSE_ERROR"
	
	// External service errors
	ErrorCodeJiraError     ErrorCode = "JIRA_ERROR"
	ErrorCodeGeminiError   ErrorCode = "GEMINI_ERROR"
	ErrorCodeGoogleError   ErrorCode = "GOOGLE_ERROR"
	
	// Security errors
	ErrorCodeKeyringError    ErrorCode = "KEYRING_ERROR"
	ErrorCodeEncryptionError ErrorCode = "ENCRYPTION_ERROR"
	ErrorCodeTLSError        ErrorCode = "TLS_ERROR"
	
	// UI errors
	ErrorCodeUIError        ErrorCode = "UI_ERROR"
	ErrorCodeValidationError ErrorCode = "VALIDATION_ERROR"
	
	// General errors
	ErrorCodeUnknown        ErrorCode = "UNKNOWN"
	ErrorCodeInternalError  ErrorCode = "INTERNAL_ERROR"
	ErrorCodeNetworkError   ErrorCode = "NETWORK_ERROR"
	ErrorCodeTimeoutError   ErrorCode = "TIMEOUT_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code      ErrorCode   `json:"code"`
	Message   string      `json:"message"`
	Details   string      `json:"details,omitempty"`
	Cause     error       `json:"cause,omitempty"`
	Retryable bool        `json:"retryable"`
	Context   ErrorContext `json:"context,omitempty"`
}

// ErrorContext provides additional context for errors
type ErrorContext struct {
	Operation string                 `json:"operation,omitempty"`
	Service   string                 `json:"service,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// NewAppError creates a new application error with stack trace
func NewAppError(code ErrorCode, message string, cause error) *AppError {
	_, file, line, _ := runtime.Caller(1)
	
	return &AppError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Retryable: isRetryable(code),
		Context: ErrorContext{
			File: file,
			Line: line,
		},
	}
}

// NewAppErrorWithContext creates a new application error with additional context
func NewAppErrorWithContext(code ErrorCode, message string, cause error, context ErrorContext) *AppError {
	_, file, line, _ := runtime.Caller(1)
	
	if context.File == "" {
		context.File = file
	}
	if context.Line == 0 {
		context.Line = line
	}
	
	return &AppError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Retryable: isRetryable(code),
		Context:   context,
	}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// String returns a detailed string representation
func (e *AppError) String() string {
	return fmt.Sprintf("AppError{Code: %s, Message: %s, Retryable: %t, Context: %+v}",
		e.Code, e.Message, e.Retryable, e.Context)
}

// WithOperation adds operation context to the error
func (e *AppError) WithOperation(operation string) *AppError {
	e.Context.Operation = operation
	return e
}

// WithService adds service context to the error
func (e *AppError) WithService(service string) *AppError {
	e.Context.Service = service
	return e
}

// WithUserID adds user ID context to the error
func (e *AppError) WithUserID(userID string) *AppError {
	e.Context.UserID = userID
	return e
}

// WithRequestID adds request ID context to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.Context.RequestID = requestID
	return e
}

// WithExtra adds extra context information
func (e *AppError) WithExtra(key string, value interface{}) *AppError {
	if e.Context.Extra == nil {
		e.Context.Extra = make(map[string]interface{})
	}
	e.Context.Extra[key] = value
	return e
}

// IsRetryable returns whether the error is retryable
func (e *AppError) IsRetryable() bool {
	return e.Retryable
}

// IsType checks if the error is of a specific type
func (e *AppError) IsType(code ErrorCode) bool {
	return e.Code == code
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// isRetryable determines if an error code represents a retryable error
func isRetryable(code ErrorCode) bool {
	retryableCodes := map[ErrorCode]bool{
		ErrorCodeAPITimeout:     true,
		ErrorCodeAPIRateLimit:   true,
		ErrorCodeAPIServerError: true,
		ErrorCodeNetworkError:   true,
		ErrorCodeTimeoutError:   true,
	}
	
	return retryableCodes[code]
}

// WrapError wraps an existing error with additional context
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, preserve the original context
	if appErr, ok := err.(*AppError); ok {
		return NewAppError(code, message, appErr)
	}
	
	return NewAppError(code, message, err)
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// Handle handles an error with appropriate logging and context
func (h *ErrorHandler) Handle(err error, operation string) {
	if err == nil {
		return
	}
	
	if appErr, ok := err.(*AppError); ok {
		h.handleAppError(appErr, operation)
	} else {
		h.handleGenericError(err, operation)
	}
}

// handleAppError handles application-specific errors
func (h *ErrorHandler) handleAppError(err *AppError, operation string) {
	fields := []Field{
		NewField("error_code", string(err.Code)),
		NewField("operation", operation),
		NewField("retryable", err.Retryable),
	}
	
	if err.Context.Service != "" {
		fields = append(fields, NewField("service", err.Context.Service))
	}
	
	if err.Context.UserID != "" {
		fields = append(fields, NewField("user_id", err.Context.UserID))
	}
	
	if err.Context.RequestID != "" {
		fields = append(fields, NewField("request_id", err.Context.RequestID))
	}
	
	if err.Context.File != "" {
		fields = append(fields, NewField("file", err.Context.File))
	}
	
	if err.Context.Line != 0 {
		fields = append(fields, NewField("line", err.Context.Line))
	}
	
	for key, value := range err.Context.Extra {
		fields = append(fields, NewField(key, value))
	}
	
	h.logger.Error(err.Message, err.Cause, fields...)
}

// handleGenericError handles generic errors
func (h *ErrorHandler) handleGenericError(err error, operation string) {
	h.logger.Error("Unhandled error", err,
		NewField("operation", operation),
		NewField("error_code", string(ErrorCodeUnknown)),
	)
}

// RecoverFromPanic recovers from panics and converts them to errors
func RecoverFromPanic(logger Logger) {
	if r := recover(); r != nil {
		var err error
		if e, ok := r.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", r)
		}
		
		// Log the panic with stack trace
		logger.Error("Panic recovered", err,
			NewField("error_code", string(ErrorCodeInternalError)),
			NewField("panic_value", r),
		)
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	
	if len(v) == 1 {
		return fmt.Sprintf("validation failed: %s", v[0].Message)
	}
	
	return fmt.Sprintf("validation failed with %d errors", len(v))
}

// Add adds a validation error
func (v *ValidationErrors) Add(field, message string, value interface{}) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// GetFieldErrors returns all errors for a specific field
func (v ValidationErrors) GetFieldErrors(field string) []ValidationError {
	var errors []ValidationError
	for _, err := range v {
		if err.Field == field {
			errors = append(errors, err)
		}
	}
	return errors
}

// ToAppError converts validation errors to an AppError
func (v ValidationErrors) ToAppError() *AppError {
	if !v.HasErrors() {
		return nil
	}
	
	return NewAppError(ErrorCodeValidationError, v.Error(), nil).
		WithExtra("validation_errors", v)
}