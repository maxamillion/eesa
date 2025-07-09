package validation

import (
	"testing"
	"time"

	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceValidationRules(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	assert.NotNil(t, rules)
	assert.NotNil(t, rules.validator)
	
	// Check that predefined rules are registered
	assert.True(t, rules.validator.HasRules("jira", "search"))
	assert.True(t, rules.validator.HasRules("gemini", "generateContent"))
	assert.True(t, rules.validator.HasRules("google_docs", "create"))
}

func TestServiceValidationRules_JiraValidation(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	tests := []struct {
		name     string
		endpoint string
		data     interface{}
		expected bool
	}{
		{
			name:     "valid search response",
			endpoint: "search",
			data: map[string]interface{}{
				"startAt":    0,
				"maxResults": 50,
				"total":      100,
				"issues":     []interface{}{},
			},
			expected: true,
		},
		{
			name:     "invalid search response - missing total",
			endpoint: "search",
			data: map[string]interface{}{
				"startAt":    0,
				"maxResults": 50,
				"issues":     []interface{}{},
			},
			expected: false,
		},
		{
			name:     "invalid search response - negative startAt",
			endpoint: "search",
			data: map[string]interface{}{
				"startAt":    -1,
				"maxResults": 50,
				"total":      100,
				"issues":     []interface{}{},
			},
			expected: false,
		},
		{
			name:     "valid issue response",
			endpoint: "issue",
			data: map[string]interface{}{
				"id":  "12345",
				"key": "TEST-123",
				"fields": map[string]interface{}{
					"summary": "Test issue",
					"issuetype": map[string]interface{}{
						"name": "Bug",
					},
					"status": map[string]interface{}{
						"name": "Open",
					},
					"priority": map[string]interface{}{
						"name": "High",
					},
					"created": "2023-01-01T10:00:00.000Z",
					"updated": "2023-01-01T10:00:00.000Z",
				},
			},
			expected: true,
		},
		{
			name:     "invalid issue response - invalid key format",
			endpoint: "issue",
			data: map[string]interface{}{
				"id":  "12345",
				"key": "invalid-key",
				"fields": map[string]interface{}{
					"summary": "Test issue",
					"issuetype": map[string]interface{}{
						"name": "Bug",
					},
					"status": map[string]interface{}{
						"name": "Open",
					},
					"created": "2023-01-01T10:00:00.000Z",
					"updated": "2023-01-01T10:00:00.000Z",
				},
			},
			expected: false,
		},
		{
			name:     "invalid issue response - invalid issue type",
			endpoint: "issue",
			data: map[string]interface{}{
				"id":  "12345",
				"key": "TEST-123",
				"fields": map[string]interface{}{
					"summary": "Test issue",
					"issuetype": map[string]interface{}{
						"name": "InvalidType",
					},
					"status": map[string]interface{}{
						"name": "Open",
					},
					"created": "2023-01-01T10:00:00.000Z",
					"updated": "2023-01-01T10:00:00.000Z",
				},
			},
			expected: false,
		},
		{
			name:     "valid myself response",
			endpoint: "myself",
			data: map[string]interface{}{
				"accountId":   "user123",
				"displayName": "John Doe",
				"active":      true,
			},
			expected: true,
		},
		{
			name:     "invalid myself response - missing accountId",
			endpoint: "myself",
			data: map[string]interface{}{
				"displayName": "John Doe",
				"active":      true,
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ValidateJiraResponse(tt.endpoint, tt.data)
			assert.Equal(t, tt.expected, result.Valid)
			
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestServiceValidationRules_GeminiValidation(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	tests := []struct {
		name     string
		endpoint string
		data     interface{}
		expected bool
	}{
		{
			name:     "valid generate content response",
			endpoint: "generateContent",
			data: map[string]interface{}{
				"candidates": []interface{}{
					map[string]interface{}{
						"content": map[string]interface{}{
							"parts": []interface{}{
								map[string]interface{}{
									"text": "Generated content",
								},
							},
						},
					},
				},
				"usageMetadata": map[string]interface{}{
					"promptTokenCount":     100,
					"candidatesTokenCount": 50,
					"totalTokenCount":      150,
				},
			},
			expected: true,
		},
		{
			name:     "invalid generate content response - empty candidates",
			endpoint: "generateContent",
			data: map[string]interface{}{
				"candidates": []interface{}{},
			},
			expected: false,
		},
		{
			name:     "invalid generate content response - negative token count",
			endpoint: "generateContent",
			data: map[string]interface{}{
				"candidates": []interface{}{
					map[string]interface{}{
						"content": "some content",
					},
				},
				"usageMetadata": map[string]interface{}{
					"promptTokenCount":     -10,
					"candidatesTokenCount": 50,
					"totalTokenCount":      150,
				},
			},
			expected: false,
		},
		{
			name:     "valid models response",
			endpoint: "models",
			data: map[string]interface{}{
				"models": []interface{}{
					map[string]interface{}{
						"name": "gemini-pro",
					},
				},
			},
			expected: true,
		},
		{
			name:     "invalid models response - missing models",
			endpoint: "models",
			data:     map[string]interface{}{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ValidateGeminiResponse(tt.endpoint, tt.data)
			assert.Equal(t, tt.expected, result.Valid)
			
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestServiceValidationRules_GoogleDocsValidation(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	tests := []struct {
		name     string
		endpoint string
		data     interface{}
		expected bool
	}{
		{
			name:     "valid create response",
			endpoint: "create",
			data: map[string]interface{}{
				"documentId": "doc123",
				"title":      "Test Document",
			},
			expected: true,
		},
		{
			name:     "invalid create response - empty documentId",
			endpoint: "create",
			data: map[string]interface{}{
				"documentId": "",
				"title":      "Test Document",
			},
			expected: false,
		},
		{
			name:     "valid get response",
			endpoint: "get",
			data: map[string]interface{}{
				"documentId": "doc123",
				"title":      "Test Document",
				"body": map[string]interface{}{
					"content": []interface{}{
						map[string]interface{}{
							"paragraph": map[string]interface{}{
								"elements": []interface{}{},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name:     "invalid get response - missing documentId",
			endpoint: "get",
			data: map[string]interface{}{
				"title": "Test Document",
			},
			expected: false,
		},
		{
			name:     "valid batch update response",
			endpoint: "batchUpdate",
			data: map[string]interface{}{
				"documentId": "doc123",
				"replies": []interface{}{
					map[string]interface{}{
						"insertText": map[string]interface{}{},
					},
				},
			},
			expected: true,
		},
		{
			name:     "invalid batch update response - empty documentId",
			endpoint: "batchUpdate",
			data: map[string]interface{}{
				"documentId": "",
				"replies":    []interface{}{},
			},
			expected: false,
		},
		{
			name:     "valid permission response",
			endpoint: "permission",
			data: map[string]interface{}{
				"id":   "perm123",
				"type": "user",
				"role": "reader",
			},
			expected: true,
		},
		{
			name:     "invalid permission response - invalid type",
			endpoint: "permission",
			data: map[string]interface{}{
				"id":   "perm123",
				"type": "invalid_type",
				"role": "reader",
			},
			expected: false,
		},
		{
			name:     "invalid permission response - invalid role",
			endpoint: "permission",
			data: map[string]interface{}{
				"id":   "perm123",
				"type": "user",
				"role": "invalid_role",
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ValidateGoogleDocsResponse(tt.endpoint, tt.data)
			assert.Equal(t, tt.expected, result.Valid)
			
			if !tt.expected {
				assert.True(t, len(result.Errors) > 0)
			}
		})
	}
}

func TestServiceValidationRules_GetValidationRules(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	// Test Jira rules
	jiraRule, exists := rules.GetJiraValidationRule("search")
	assert.True(t, exists)
	assert.Equal(t, "root", jiraRule.Field)
	assert.Equal(t, "object", jiraRule.Type)
	
	_, exists = rules.GetJiraValidationRule("nonexistent")
	assert.False(t, exists)
	
	// Test Gemini rules
	geminiRule, exists := rules.GetGeminiValidationRule("generateContent")
	assert.True(t, exists)
	assert.Equal(t, "root", geminiRule.Field)
	assert.Equal(t, "object", geminiRule.Type)
	
	_, exists = rules.GetGeminiValidationRule("nonexistent")
	assert.False(t, exists)
	
	// Test Google Docs rules
	gdocsRule, exists := rules.GetGoogleDocsValidationRule("create")
	assert.True(t, exists)
	assert.Equal(t, "root", gdocsRule.Field)
	assert.Equal(t, "object", gdocsRule.Type)
	
	_, exists = rules.GetGoogleDocsValidationRule("nonexistent")
	assert.False(t, exists)
}

func TestServiceValidationRules_CreateRules(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	// Test custom rule creation
	customRule := rules.CreateCustomRule("test_field", "string", true)
	assert.Equal(t, "test_field", customRule.Field)
	assert.Equal(t, "string", customRule.Type)
	assert.True(t, customRule.Required)
	
	// Test string rule creation
	stringRule := rules.CreateStringRule("name", true, testIntPtr(1), testIntPtr(100), testStringPtr("^[A-Za-z]+$"))
	assert.Equal(t, "name", stringRule.Field)
	assert.Equal(t, "string", stringRule.Type)
	assert.True(t, stringRule.Required)
	assert.Equal(t, 1, *stringRule.MinLength)
	assert.Equal(t, 100, *stringRule.MaxLength)
	assert.Equal(t, "^[A-Za-z]+$", *stringRule.Pattern)
	
	// Test number rule creation
	numberRule := rules.CreateNumberRule("age", true, testFloatPtr(0), testFloatPtr(150))
	assert.Equal(t, "age", numberRule.Field)
	assert.Equal(t, "number", numberRule.Type)
	assert.True(t, numberRule.Required)
	assert.Equal(t, 0.0, *numberRule.MinValue)
	assert.Equal(t, 150.0, *numberRule.MaxValue)
	
	// Test enum rule creation
	enumRule := rules.CreateEnumRule("status", true, []string{"active", "inactive"})
	assert.Equal(t, "status", enumRule.Field)
	assert.Equal(t, "string", enumRule.Type)
	assert.True(t, enumRule.Required)
	assert.Equal(t, []string{"active", "inactive"}, enumRule.Enum)
	
	// Test array rule creation
	arrayRule := rules.CreateArrayRule("items", true, testIntPtr(1), testIntPtr(10))
	assert.Equal(t, "items", arrayRule.Field)
	assert.Equal(t, "array", arrayRule.Type)
	assert.True(t, arrayRule.Required)
	assert.Equal(t, 1, *arrayRule.MinLength)
	assert.Equal(t, 10, *arrayRule.MaxLength)
	
	// Test object rule creation
	nested := map[string]ValidationRule{
		"nested_field": {
			Field:    "nested_field",
			Type:     "string",
			Required: true,
		},
	}
	objectRule := rules.CreateObjectRule("config", true, nested)
	assert.Equal(t, "config", objectRule.Field)
	assert.Equal(t, "object", objectRule.Type)
	assert.True(t, objectRule.Required)
	assert.Equal(t, nested, objectRule.Nested)
}

func TestServiceValidationRules_AddCustomValidationRule(t *testing.T) {
	logger := utils.NewMockLogger()
	rules := NewServiceValidationRules(logger)
	
	customRule := ValidationRule{
		Field:    "custom_field",
		Type:     "string",
		Required: true,
	}
	
	rules.AddCustomValidationRule("custom_service", "custom_endpoint", customRule)
	
	assert.True(t, rules.validator.HasRules("custom_service", "custom_endpoint"))
	
	result := rules.validator.ValidateResponse("custom_service", "custom_endpoint", "test_value")
	assert.True(t, result.Valid)
}

func TestValidationMetrics(t *testing.T) {
	metrics := NewValidationMetrics(10)
	
	// Test initial state
	summary := metrics.GetSummary()
	assert.Equal(t, 0, summary.TotalValidations)
	assert.Equal(t, 0, summary.SuccessfulCount)
	assert.Equal(t, 0, summary.FailedCount)
	assert.Equal(t, 0, summary.WarningCount)
	
	// Record successful validation
	successResult := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationError{},
		FieldCount:  3,
		ValidatedAt: time.Now(),
	}
	
	metrics.RecordValidation("test_service", "test_endpoint", successResult)
	
	summary = metrics.GetSummary()
	assert.Equal(t, 1, summary.TotalValidations)
	assert.Equal(t, 1, summary.SuccessfulCount)
	assert.Equal(t, 0, summary.FailedCount)
	assert.Equal(t, 0, summary.WarningCount)
	assert.Contains(t, summary.ServiceResults, "test_service")
	
	// Record failed validation
	failedResult := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				Field:   "test_field",
				Type:    "required",
				Message: "Field is required",
			},
		},
		Warnings:    []ValidationError{},
		FieldCount:  2,
		ValidatedAt: time.Now(),
	}
	
	metrics.RecordValidation("test_service", "test_endpoint", failedResult)
	
	summary = metrics.GetSummary()
	assert.Equal(t, 2, summary.TotalValidations)
	assert.Equal(t, 1, summary.SuccessfulCount)
	assert.Equal(t, 1, summary.FailedCount)
	assert.Equal(t, 0, summary.WarningCount)
	
	// Record validation with warnings
	warningResult := &ValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Warnings: []ValidationError{
			{
				Field:   "test_field",
				Type:    "warning",
				Message: "Field has warning",
			},
		},
		FieldCount:  1,
		ValidatedAt: time.Now(),
	}
	
	metrics.RecordValidation("test_service", "test_endpoint", warningResult)
	
	summary = metrics.GetSummary()
	assert.Equal(t, 3, summary.TotalValidations)
	assert.Equal(t, 2, summary.SuccessfulCount)
	assert.Equal(t, 1, summary.FailedCount)
	assert.Equal(t, 1, summary.WarningCount)
	
	// Test recent results
	recentResults := metrics.GetRecentResults()
	assert.Equal(t, 3, len(recentResults))
	
	// Test service summary
	serviceSummary := summary.ServiceResults["test_service"]
	assert.Equal(t, 3, serviceSummary.TotalCalls)
	assert.Equal(t, 2, serviceSummary.SuccessfulCalls)
	assert.Equal(t, 1, serviceSummary.FailedCalls)
	assert.Equal(t, 1, serviceSummary.WarningCalls)
	assert.Equal(t, 3, serviceSummary.EndpointResults["test_endpoint"])
	
	// Test reset
	metrics.Reset()
	summary = metrics.GetSummary()
	assert.Equal(t, 0, summary.TotalValidations)
	assert.Equal(t, 0, len(metrics.GetRecentResults()))
}

func TestValidationMetrics_MaxResults(t *testing.T) {
	metrics := NewValidationMetrics(2) // Max 2 results
	
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationError{},
		FieldCount:  1,
		ValidatedAt: time.Now(),
	}
	
	// Add 3 results
	for i := 0; i < 3; i++ {
		metrics.RecordValidation("test_service", "test_endpoint", result)
	}
	
	// Should only keep the last 2 results
	recentResults := metrics.GetRecentResults()
	assert.Equal(t, 2, len(recentResults))
	
	// But total count should be 3
	summary := metrics.GetSummary()
	assert.Equal(t, 3, summary.TotalValidations)
}

// Helper functions for tests
func testIntPtr(i int) *int {
	return &i
}

func testFloatPtr(f float64) *float64 {
	return &f
}

func testStringPtr(s string) *string {
	return &s
}