package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/company/eesa/pkg/utils"
)

// IntegrationExample demonstrates how to use the validation system
// with API clients in a real application scenario
type IntegrationExample struct {
	validator *ServiceValidationRules
	metrics   *ValidationMetrics
	logger    utils.Logger
}

// NewIntegrationExample creates a new integration example
func NewIntegrationExample(logger utils.Logger) *IntegrationExample {
	return &IntegrationExample{
		validator: NewServiceValidationRules(logger),
		metrics:   NewValidationMetrics(100), // Keep last 100 results
		logger:    logger,
	}
}

// ValidateJiraSearchResponse demonstrates validating a Jira search response
func (e *IntegrationExample) ValidateJiraSearchResponse(ctx context.Context, resp *http.Response, body []byte) (*ValidationResult, error) {
	// First validate the HTTP response
	httpResult := e.validator.GetValidator().ValidateHTTPResponse("jira", "search", resp, body)
	e.metrics.RecordValidation("jira", "search", httpResult)
	
	if !httpResult.Valid {
		e.logger.Error("HTTP response validation failed",
			utils.NewField("service", "jira"),
			utils.NewField("endpoint", "search"),
			utils.NewField("errors", len(httpResult.Errors)),
		)
		return httpResult, fmt.Errorf("HTTP response validation failed: %d errors", len(httpResult.Errors))
	}
	
	// Parse and validate the JSON response
	var searchData interface{}
	if err := json.Unmarshal(body, &searchData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	
	// Validate the parsed data
	result := e.validator.ValidateJiraResponse("search", searchData)
	e.metrics.RecordValidation("jira", "search", result)
	
	if !result.Valid {
		e.logger.Warn("Jira search response validation failed",
			utils.NewField("errors", len(result.Errors)),
			utils.NewField("warnings", len(result.Warnings)),
		)
		
		// Log specific validation errors
		for _, err := range result.Errors {
			e.logger.Error("Validation error",
				utils.NewField("field", err.Field),
				utils.NewField("type", err.Type),
				utils.NewField("message", err.Message),
			)
		}
	}
	
	return result, nil
}

// ValidateGeminiGenerateResponse demonstrates validating a Gemini generate response
func (e *IntegrationExample) ValidateGeminiGenerateResponse(ctx context.Context, responseData interface{}) (*ValidationResult, error) {
	result := e.validator.ValidateGeminiResponse("generateContent", responseData)
	e.metrics.RecordValidation("gemini", "generateContent", result)
	
	if !result.Valid {
		e.logger.Error("Gemini generate response validation failed",
			utils.NewField("errors", len(result.Errors)),
			utils.NewField("warnings", len(result.Warnings)),
		)
		return result, fmt.Errorf("validation failed: %d errors", len(result.Errors))
	}
	
	e.logger.Info("Gemini response validated successfully",
		utils.NewField("field_count", result.FieldCount),
		utils.NewField("warnings", len(result.Warnings)),
	)
	
	return result, nil
}

// ValidateGoogleDocsCreateResponse demonstrates validating a Google Docs create response
func (e *IntegrationExample) ValidateGoogleDocsCreateResponse(ctx context.Context, responseData interface{}) (*ValidationResult, error) {
	result := e.validator.ValidateGoogleDocsResponse("create", responseData)
	e.metrics.RecordValidation("google_docs", "create", result)
	
	if !result.Valid {
		e.logger.Error("Google Docs create response validation failed",
			utils.NewField("errors", len(result.Errors)),
		)
		return result, fmt.Errorf("validation failed: %d errors", len(result.Errors))
	}
	
	return result, nil
}

// GetValidationMetrics returns current validation metrics
func (e *IntegrationExample) GetValidationMetrics() *ValidationSummary {
	return e.metrics.GetSummary()
}

// GetRecentValidationResults returns recent validation results
func (e *IntegrationExample) GetRecentValidationResults() []ValidationResult {
	return e.metrics.GetRecentResults()
}

// AddCustomValidationRule adds a custom validation rule
func (e *IntegrationExample) AddCustomValidationRule(service, endpoint string, rule ValidationRule) {
	e.validator.AddCustomValidationRule(service, endpoint, rule)
	e.logger.Info("Added custom validation rule",
		utils.NewField("service", service),
		utils.NewField("endpoint", endpoint),
		utils.NewField("field", rule.Field),
	)
}

// ValidateWithCustomRules demonstrates using custom validation rules
func (e *IntegrationExample) ValidateWithCustomRules(ctx context.Context) error {
	// Add a custom rule for a hypothetical API endpoint
	customRule := ValidationRule{
		Field:    "root",
		Type:     "object",
		Required: true,
		Nested: map[string]ValidationRule{
			"customField": {
				Field:       "customField",
				Type:        "string",
				Required:    true,
				MinLength:   intPtr(1),
				MaxLength:   intPtr(100),
				Pattern:     stringPtr("^[a-zA-Z0-9_-]+$"),
				CustomRules: []string{"non_empty"},
			},
			"numericField": {
				Field:       "numericField",
				Type:        "number",
				Required:    true,
				MinValue:    floatPtr(0),
				MaxValue:    floatPtr(1000),
				CustomRules: []string{"positive"},
			},
			"emailField": {
				Field:       "emailField",
				Type:        "string",
				Required:    false,
				CustomRules: []string{"email"},
			},
		},
	}
	
	e.AddCustomValidationRule("custom_service", "custom_endpoint", customRule)
	
	// Test the custom rule
	testData := map[string]interface{}{
		"customField":  "valid_field_123",
		"numericField": 42,
		"emailField":   "user@example.com",
	}
	
	result := e.validator.GetValidator().ValidateResponse("custom_service", "custom_endpoint", testData)
	e.metrics.RecordValidation("custom_service", "custom_endpoint", result)
	
	if !result.Valid {
		return fmt.Errorf("custom validation failed: %d errors", len(result.Errors))
	}
	
	e.logger.Info("Custom validation successful",
		utils.NewField("field_count", result.FieldCount),
	)
	
	return nil
}

// MonitorValidationHealth provides health monitoring for validation
func (e *IntegrationExample) MonitorValidationHealth() *ValidationHealthStatus {
	summary := e.metrics.GetSummary()
	
	status := &ValidationHealthStatus{
		OverallHealth: "healthy",
		TotalValidations: summary.TotalValidations,
		SuccessRate: 0,
		FailureRate: 0,
		WarningRate: 0,
	}
	
	if summary.TotalValidations > 0 {
		status.SuccessRate = float64(summary.SuccessfulCount) / float64(summary.TotalValidations) * 100
		status.FailureRate = float64(summary.FailedCount) / float64(summary.TotalValidations) * 100
		status.WarningRate = float64(summary.WarningCount) / float64(summary.TotalValidations) * 100
	}
	
	// Determine overall health
	if status.FailureRate > 20 {
		status.OverallHealth = "unhealthy"
	} else if status.FailureRate > 10 || status.WarningRate > 30 {
		status.OverallHealth = "degraded"
	}
	
	// Add service-specific health
	status.ServiceHealth = make(map[string]string)
	for serviceName, serviceResult := range summary.ServiceResults {
		serviceFailureRate := float64(serviceResult.FailedCalls) / float64(serviceResult.TotalCalls) * 100
		
		if serviceFailureRate > 20 {
			status.ServiceHealth[serviceName] = "unhealthy"
		} else if serviceFailureRate > 10 {
			status.ServiceHealth[serviceName] = "degraded"
		} else {
			status.ServiceHealth[serviceName] = "healthy"
		}
	}
	
	return status
}

// ValidationHealthStatus represents the health status of validation
type ValidationHealthStatus struct {
	OverallHealth     string             `json:"overall_health"`
	TotalValidations  int                `json:"total_validations"`
	SuccessRate       float64            `json:"success_rate"`
	FailureRate       float64            `json:"failure_rate"`
	WarningRate       float64            `json:"warning_rate"`
	ServiceHealth     map[string]string  `json:"service_health"`
}

// ResetMetrics resets all validation metrics
func (e *IntegrationExample) ResetMetrics() {
	e.metrics.Reset()
	e.logger.Info("Validation metrics reset")
}

// Example usage patterns for integration with API clients
func ExampleUsagePatterns() {
	logger := utils.NewMockLogger()
	example := NewIntegrationExample(logger)
	
	// Example 1: Validate Jira API response
	jiraResponseBody := `{
		"startAt": 0,
		"maxResults": 50,
		"total": 100,
		"issues": [
			{
				"id": "12345",
				"key": "TEST-123",
				"fields": {
					"summary": "Test issue",
					"issuetype": {"name": "Bug"},
					"status": {"name": "Open"},
					"created": "2023-01-01T10:00:00.000Z",
					"updated": "2023-01-01T10:00:00.000Z"
				}
			}
		]
	}`
	
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")
	
	result, err := example.ValidateJiraSearchResponse(context.Background(), resp, []byte(jiraResponseBody))
	if err != nil {
		logger.Error("Jira validation failed", utils.NewField("error", err.Error()))
	} else if result.Valid {
		logger.Info("Jira validation successful")
	}
	
	// Example 2: Validate Gemini API response
	geminiResponseData := map[string]interface{}{
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
			"totalTokenCount": 150,
		},
	}
	
	geminiResult, err := example.ValidateGeminiGenerateResponse(context.Background(), geminiResponseData)
	if err != nil {
		logger.Error("Gemini validation failed", utils.NewField("error", err.Error()))
	} else if geminiResult.Valid {
		logger.Info("Gemini validation successful")
	}
	
	// Example 3: Custom validation rules
	err = example.ValidateWithCustomRules(context.Background())
	if err != nil {
		logger.Error("Custom validation failed", utils.NewField("error", err.Error()))
	}
	
	// Example 4: Health monitoring
	health := example.MonitorValidationHealth()
	logger.Info("Validation health status",
		utils.NewField("overall_health", health.OverallHealth),
		utils.NewField("success_rate", health.SuccessRate),
		utils.NewField("failure_rate", health.FailureRate),
	)
	
	// Example 5: Get metrics
	metrics := example.GetValidationMetrics()
	logger.Info("Validation metrics",
		utils.NewField("total_validations", metrics.TotalValidations),
		utils.NewField("successful_count", metrics.SuccessfulCount),
		utils.NewField("failed_count", metrics.FailedCount),
	)
}

// Helper functions for pointer types
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}