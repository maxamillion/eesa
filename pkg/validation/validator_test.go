package validation

import (
	"net/http"
	"testing"

	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIResponseValidator(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	assert.NotNil(t, validator)
	assert.Equal(t, logger, validator.logger)
	assert.NotNil(t, validator.rules)
	assert.Equal(t, 0, len(validator.rules))
}

func TestAPIResponseValidator_RegisterRule(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test_field",
		Type:     "string",
		Required: true,
	}
	
	validator.RegisterRule("test_service", "test_endpoint", rule)
	
	assert.True(t, validator.HasRules("test_service", "test_endpoint"))
	assert.False(t, validator.HasRules("test_service", "other_endpoint"))
	assert.False(t, validator.HasRules("other_service", "test_endpoint"))
}

func TestAPIResponseValidator_RegisterRules(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rules := map[string]ValidationRule{
		"endpoint1": {
			Field:    "field1",
			Type:     "string",
			Required: true,
		},
		"endpoint2": {
			Field:    "field2",
			Type:     "number",
			Required: false,
		},
	}
	
	validator.RegisterRules("test_service", rules)
	
	assert.True(t, validator.HasRules("test_service", "endpoint1"))
	assert.True(t, validator.HasRules("test_service", "endpoint2"))
	assert.False(t, validator.HasRules("test_service", "endpoint3"))
}

func TestAPIResponseValidator_ValidateString(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	tests := []struct {
		name     string
		rule     ValidationRule
		value    interface{}
		expected bool
	}{
		{
			name: "valid string",
			rule: ValidationRule{
				Field:    "test",
				Type:     "string",
				Required: true,
			},
			value:    "test string",
			expected: true,
		},
		{
			name: "invalid type",
			rule: ValidationRule{
				Field:    "test",
				Type:     "string",
				Required: true,
			},
			value:    123,
			expected: false,
		},
		{
			name: "string too short",
			rule: ValidationRule{
				Field:     "test",
				Type:      "string",
				Required:  true,
				MinLength: testIntPtr(5),
			},
			value:    "abc",
			expected: false,
		},
		{
			name: "string too long",
			rule: ValidationRule{
				Field:     "test",
				Type:      "string",
				Required:  true,
				MaxLength: testIntPtr(5),
			},
			value:    "abcdefg",
			expected: false,
		},
		{
			name: "pattern match success",
			rule: ValidationRule{
				Field:    "test",
				Type:     "string",
				Required: true,
				Pattern:  testStringPtr("^[A-Z]+-\\d+$"),
			},
			value:    "ABC-123",
			expected: true,
		},
		{
			name: "pattern match failure",
			rule: ValidationRule{
				Field:    "test",
				Type:     "string",
				Required: true,
				Pattern:  testStringPtr("^[A-Z]+-\\d+$"),
			},
			value:    "invalid-pattern",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.RegisterRule("test", "endpoint", tt.rule)
			result := validator.ValidateResponse("test", "endpoint", tt.value)
			
			assert.Equal(t, tt.expected, result.Valid)
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestAPIResponseValidator_ValidateNumber(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	tests := []struct {
		name     string
		rule     ValidationRule
		value    interface{}
		expected bool
	}{
		{
			name: "valid int",
			rule: ValidationRule{
				Field:    "test",
				Type:     "number",
				Required: true,
			},
			value:    42,
			expected: true,
		},
		{
			name: "valid float",
			rule: ValidationRule{
				Field:    "test",
				Type:     "number",
				Required: true,
			},
			value:    3.14,
			expected: true,
		},
		{
			name: "invalid type",
			rule: ValidationRule{
				Field:    "test",
				Type:     "number",
				Required: true,
			},
			value:    "not a number",
			expected: false,
		},
		{
			name: "number too small",
			rule: ValidationRule{
				Field:    "test",
				Type:     "number",
				Required: true,
				MinValue: testFloatPtr(10),
			},
			value:    5,
			expected: false,
		},
		{
			name: "number too large",
			rule: ValidationRule{
				Field:    "test",
				Type:     "number",
				Required: true,
				MaxValue: testFloatPtr(10),
			},
			value:    15,
			expected: false,
		},
		{
			name: "integer validation success",
			rule: ValidationRule{
				Field:    "test",
				Type:     "integer",
				Required: true,
			},
			value:    42,
			expected: true,
		},
		{
			name: "integer validation failure",
			rule: ValidationRule{
				Field:    "test",
				Type:     "integer",
				Required: true,
			},
			value:    3.14,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.RegisterRule("test", "endpoint", tt.rule)
			result := validator.ValidateResponse("test", "endpoint", tt.value)
			
			assert.Equal(t, tt.expected, result.Valid)
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestAPIResponseValidator_ValidateBoolean(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "boolean",
		Required: true,
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	// Valid boolean
	result := validator.ValidateResponse("test", "endpoint", true)
	assert.True(t, result.Valid)
	
	// Invalid type
	result = validator.ValidateResponse("test", "endpoint", "not a boolean")
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
}

func TestAPIResponseValidator_ValidateArray(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	tests := []struct {
		name     string
		rule     ValidationRule
		value    interface{}
		expected bool
	}{
		{
			name: "valid array",
			rule: ValidationRule{
				Field:    "test",
				Type:     "array",
				Required: true,
			},
			value:    []string{"a", "b", "c"},
			expected: true,
		},
		{
			name: "invalid type",
			rule: ValidationRule{
				Field:    "test",
				Type:     "array",
				Required: true,
			},
			value:    "not an array",
			expected: false,
		},
		{
			name: "array too short",
			rule: ValidationRule{
				Field:     "test",
				Type:      "array",
				Required:  true,
				MinLength: testIntPtr(3),
			},
			value:    []string{"a", "b"},
			expected: false,
		},
		{
			name: "array too long",
			rule: ValidationRule{
				Field:     "test",
				Type:      "array",
				Required:  true,
				MaxLength: testIntPtr(2),
			},
			value:    []string{"a", "b", "c"},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.RegisterRule("test", "endpoint", tt.rule)
			result := validator.ValidateResponse("test", "endpoint", tt.value)
			
			assert.Equal(t, tt.expected, result.Valid)
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestAPIResponseValidator_ValidateObject(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "object",
		Required: true,
		Nested: map[string]ValidationRule{
			"name": {
				Field:    "name",
				Type:     "string",
				Required: true,
			},
			"age": {
				Field:    "age",
				Type:     "integer",
				Required: true,
				MinValue: testFloatPtr(0),
			},
		},
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	// Valid object
	validObj := map[string]interface{}{
		"name": "John",
		"age":  30,
	}
	result := validator.ValidateResponse("test", "endpoint", validObj)
	assert.True(t, result.Valid)
	
	// Missing required field
	invalidObj := map[string]interface{}{
		"name": "John",
	}
	result = validator.ValidateResponse("test", "endpoint", invalidObj)
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	
	// Invalid nested field
	invalidObj2 := map[string]interface{}{
		"name": "John",
		"age":  -5,
	}
	result = validator.ValidateResponse("test", "endpoint", invalidObj2)
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
}

func TestAPIResponseValidator_ValidateTimestamp(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "timestamp",
		Required: true,
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	// Valid timestamps
	validTimestamps := []string{
		"2023-01-01T10:00:00Z",
		"2023-01-01T10:00:00.000Z",
		"2023-01-01T10:00:00+00:00",
		"2023-01-01 10:00:00",
	}
	
	for _, ts := range validTimestamps {
		result := validator.ValidateResponse("test", "endpoint", ts)
		assert.True(t, result.Valid, "Timestamp should be valid: %s", ts)
	}
	
	// Invalid timestamps
	invalidTimestamps := []string{
		"invalid-timestamp",
		"2023-13-01T10:00:00Z",
		"not-a-date",
	}
	
	for _, ts := range invalidTimestamps {
		result := validator.ValidateResponse("test", "endpoint", ts)
		assert.False(t, result.Valid, "Timestamp should be invalid: %s", ts)
	}
}

func TestAPIResponseValidator_ValidateEnum(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "string",
		Required: true,
		Enum:     []string{"red", "green", "blue"},
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	// Valid enum value
	result := validator.ValidateResponse("test", "endpoint", "red")
	assert.True(t, result.Valid)
	
	// Invalid enum value
	result = validator.ValidateResponse("test", "endpoint", "yellow")
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
}

func TestAPIResponseValidator_CustomRules(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	tests := []struct {
		name       string
		customRule string
		value      interface{}
		expected   bool
	}{
		{
			name:       "non_empty success",
			customRule: "non_empty",
			value:      "not empty",
			expected:   true,
		},
		{
			name:       "non_empty failure",
			customRule: "non_empty",
			value:      "   ",
			expected:   false,
		},
		{
			name:       "email success",
			customRule: "email",
			value:      "user@example.com",
			expected:   true,
		},
		{
			name:       "email failure",
			customRule: "email",
			value:      "invalid-email",
			expected:   false,
		},
		{
			name:       "url success",
			customRule: "url",
			value:      "https://example.com",
			expected:   true,
		},
		{
			name:       "url failure",
			customRule: "url",
			value:      "not-a-url",
			expected:   false,
		},
		{
			name:       "uuid success",
			customRule: "uuid",
			value:      "123e4567-e89b-12d3-a456-426614174000",
			expected:   true,
		},
		{
			name:       "uuid failure",
			customRule: "uuid",
			value:      "not-a-uuid",
			expected:   false,
		},
		{
			name:       "positive success",
			customRule: "positive",
			value:      42,
			expected:   true,
		},
		{
			name:       "positive failure",
			customRule: "positive",
			value:      -5,
			expected:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use appropriate type for the custom rule
			ruleType := "string"
			if tt.customRule == "positive" {
				ruleType = "number"
			}
			
			rule := ValidationRule{
				Field:       "test",
				Type:        ruleType,
				Required:    true,
				CustomRules: []string{tt.customRule},
			}
			
			validator.RegisterRule("test", "endpoint", rule)
			result := validator.ValidateResponse("test", "endpoint", tt.value)
			
			assert.Equal(t, tt.expected, result.Valid)
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestAPIResponseValidator_ValidateHTTPResponse(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	// Valid HTTP response
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")
	
	body := []byte(`{"message": "success"}`)
	result := validator.ValidateHTTPResponse("test", "endpoint", resp, body)
	
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	
	// Invalid status code
	resp.StatusCode = 500
	result = validator.ValidateHTTPResponse("test", "endpoint", resp, body)
	
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	
	// Invalid content type (warning)
	resp.StatusCode = 200
	resp.Header.Set("Content-Type", "text/plain")
	result = validator.ValidateHTTPResponse("test", "endpoint", resp, body)
	
	assert.True(t, result.Valid) // Still valid, just a warning
	assert.True(t, len(result.Warnings) > 0)
	
	// Invalid JSON
	invalidBody := []byte(`{"invalid": json}`)
	resp.Header.Set("Content-Type", "application/json")
	result = validator.ValidateHTTPResponse("test", "endpoint", resp, invalidBody)
	
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
}

func TestAPIResponseValidator_RequiredFields(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "string",
		Required: true,
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	// Missing required field (nil)
	result := validator.ValidateResponse("test", "endpoint", nil)
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	
	// Present required field
	result = validator.ValidateResponse("test", "endpoint", "present")
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestAPIResponseValidator_ClearRules(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "string",
		Required: true,
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	assert.True(t, validator.HasRules("test", "endpoint"))
	
	validator.ClearRules("test")
	assert.False(t, validator.HasRules("test", "endpoint"))
}

func TestAPIResponseValidator_GetRules(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "string",
		Required: true,
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	rules := validator.GetRules()
	assert.NotNil(t, rules)
	assert.Contains(t, rules, "test")
	assert.Contains(t, rules["test"], "endpoint")
	assert.Equal(t, rule, rules["test"]["endpoint"])
}

func TestAPIResponseValidator_NoRules(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	// Should return valid result when no rules exist
	result := validator.ValidateResponse("nonexistent", "endpoint", "any value")
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
}

func TestValidationResult_Structure(t *testing.T) {
	logger := utils.NewMockLogger()
	validator := NewAPIResponseValidator(logger)
	
	rule := ValidationRule{
		Field:    "test",
		Type:     "string",
		Required: true,
		Pattern:  testStringPtr("^[A-Z]+$"),
	}
	
	validator.RegisterRule("test", "endpoint", rule)
	
	result := validator.ValidateResponse("test", "endpoint", "invalid")
	
	assert.False(t, result.Valid)
	assert.True(t, len(result.Errors) > 0)
	assert.True(t, result.FieldCount > 0)
	assert.False(t, result.ValidatedAt.IsZero())
	
	// Check error structure
	err := result.Errors[0]
	assert.NotEmpty(t, err.Type)
	assert.NotEmpty(t, err.Message)
	assert.NotNil(t, err.Value)
	assert.NotNil(t, err.Expected)
}

// Helper functions for tests are defined in rules_test.go