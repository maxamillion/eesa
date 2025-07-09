package validation

import (
	"time"
	"github.com/company/eesa/pkg/utils"
)

// ServiceValidationRules contains predefined validation rules for all services
type ServiceValidationRules struct {
	validator *APIResponseValidator
}

// NewServiceValidationRules creates a new service validation rules manager
func NewServiceValidationRules(logger utils.Logger) *ServiceValidationRules {
	validator := NewAPIResponseValidator(logger)
	rules := &ServiceValidationRules{
		validator: validator,
	}
	
	// Register all predefined rules
	rules.registerJiraRules()
	rules.registerGeminiRules()
	rules.registerGoogleDocsRules()
	
	return rules
}

// GetValidator returns the underlying validator
func (r *ServiceValidationRules) GetValidator() *APIResponseValidator {
	return r.validator
}

// registerJiraRules registers validation rules for Jira API responses
func (r *ServiceValidationRules) registerJiraRules() {
	jiraRules := map[string]ValidationRule{
		// Search endpoint
		"search": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"startAt": {
					Field:    "startAt",
					Type:     "integer",
					Required: true,
					MinValue: floatPtr(0),
				},
				"maxResults": {
					Field:    "maxResults",
					Type:     "integer",
					Required: true,
					MinValue: floatPtr(1),
					MaxValue: floatPtr(1000),
				},
				"total": {
					Field:    "total",
					Type:     "integer",
					Required: true,
					MinValue: floatPtr(0),
				},
				"issues": {
					Field:    "issues",
					Type:     "array",
					Required: true,
				},
			},
		},
		
		// Issue endpoint
		"issue": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"id": {
					Field:       "id",
					Type:        "string",
					Required:    true,
					MinLength:   intPtr(1),
					CustomRules: []string{"non_empty"},
				},
				"key": {
					Field:       "key",
					Type:        "string",
					Required:    true,
					Pattern:     stringPtr(`^[A-Z]+-\d+$`),
					CustomRules: []string{"non_empty"},
				},
				"fields": {
					Field:    "fields",
					Type:     "object",
					Required: true,
					Nested: map[string]ValidationRule{
						"summary": {
							Field:     "summary",
							Type:      "string",
							Required:  true,
							MaxLength: intPtr(1000),
						},
						"issuetype": {
							Field:    "issuetype",
							Type:     "object",
							Required: true,
							Nested: map[string]ValidationRule{
								"name": {
									Field:    "name",
									Type:     "string",
									Required: true,
									Enum:     []string{"Bug", "Task", "Story", "Epic", "Sub-task"},
								},
							},
						},
						"status": {
							Field:    "status",
							Type:     "object",
							Required: true,
							Nested: map[string]ValidationRule{
								"name": {
									Field:    "name",
									Type:     "string",
									Required: true,
								},
							},
						},
						"priority": {
							Field:    "priority",
							Type:     "object",
							Required: false,
							Nested: map[string]ValidationRule{
								"name": {
									Field: "name",
									Type:  "string",
									Enum:  []string{"Lowest", "Low", "Medium", "High", "Highest", "Critical"},
								},
							},
						},
						"created": {
							Field:    "created",
							Type:     "timestamp",
							Required: true,
						},
						"updated": {
							Field:    "updated",
							Type:     "timestamp",
							Required: true,
						},
					},
				},
			},
		},
		
		// Worklog endpoint
		"worklog": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"worklogs": {
					Field:    "worklogs",
					Type:     "array",
					Required: true,
				},
			},
		},
		
		// Comments endpoint
		"comments": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"comments": {
					Field:    "comments",
					Type:     "array",
					Required: true,
				},
			},
		},
		
		// Current user endpoint
		"myself": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"accountId": {
					Field:       "accountId",
					Type:        "string",
					Required:    true,
					CustomRules: []string{"non_empty"},
				},
				"displayName": {
					Field:    "displayName",
					Type:     "string",
					Required: true,
				},
				"active": {
					Field:    "active",
					Type:     "boolean",
					Required: true,
				},
			},
		},
	}
	
	r.validator.RegisterRules("jira", jiraRules)
}

// registerGeminiRules registers validation rules for Gemini API responses
func (r *ServiceValidationRules) registerGeminiRules() {
	geminiRules := map[string]ValidationRule{
		// Generate content endpoint
		"generateContent": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"candidates": {
					Field:     "candidates",
					Type:      "array",
					Required:  true,
					MinLength: intPtr(1),
				},
				"usageMetadata": {
					Field:    "usageMetadata",
					Type:     "object",
					Required: false,
					Nested: map[string]ValidationRule{
						"promptTokenCount": {
							Field:    "promptTokenCount",
							Type:     "integer",
							Required: false,
							MinValue: floatPtr(0),
						},
						"candidatesTokenCount": {
							Field:    "candidatesTokenCount",
							Type:     "integer",
							Required: false,
							MinValue: floatPtr(0),
						},
						"totalTokenCount": {
							Field:    "totalTokenCount",
							Type:     "integer",
							Required: false,
							MinValue: floatPtr(0),
						},
					},
				},
			},
		},
		
		// List models endpoint
		"models": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"models": {
					Field:    "models",
					Type:     "array",
					Required: true,
				},
			},
		},
	}
	
	r.validator.RegisterRules("gemini", geminiRules)
}

// registerGoogleDocsRules registers validation rules for Google Docs API responses
func (r *ServiceValidationRules) registerGoogleDocsRules() {
	googleDocsRules := map[string]ValidationRule{
		// Create document endpoint
		"create": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"documentId": {
					Field:       "documentId",
					Type:        "string",
					Required:    true,
					CustomRules: []string{"non_empty"},
				},
				"title": {
					Field:    "title",
					Type:     "string",
					Required: true,
				},
			},
		},
		
		// Get document endpoint
		"get": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"documentId": {
					Field:       "documentId",
					Type:        "string",
					Required:    true,
					CustomRules: []string{"non_empty"},
				},
				"title": {
					Field:    "title",
					Type:     "string",
					Required: true,
				},
				"body": {
					Field:    "body",
					Type:     "object",
					Required: false,
					Nested: map[string]ValidationRule{
						"content": {
							Field:    "content",
							Type:     "array",
							Required: true,
						},
					},
				},
			},
		},
		
		// Batch update endpoint
		"batchUpdate": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"documentId": {
					Field:       "documentId",
					Type:        "string",
					Required:    true,
					CustomRules: []string{"non_empty"},
				},
				"replies": {
					Field:    "replies",
					Type:     "array",
					Required: true,
				},
			},
		},
		
		// Share permission endpoint
		"permission": {
			Field:    "root",
			Type:     "object",
			Required: true,
			Nested: map[string]ValidationRule{
				"id": {
					Field:       "id",
					Type:        "string",
					Required:    true,
					CustomRules: []string{"non_empty"},
				},
				"type": {
					Field:    "type",
					Type:     "string",
					Required: true,
					Enum:     []string{"user", "group", "domain", "anyone"},
				},
				"role": {
					Field:    "role",
					Type:     "string",
					Required: true,
					Enum:     []string{"owner", "organizer", "fileOrganizer", "writer", "commenter", "reader"},
				},
			},
		},
	}
	
	r.validator.RegisterRules("google_docs", googleDocsRules)
}

// GetJiraValidationRule returns a specific Jira validation rule
func (r *ServiceValidationRules) GetJiraValidationRule(endpoint string) (ValidationRule, bool) {
	if rules, exists := r.validator.rules["jira"]; exists {
		if rule, exists := rules[endpoint]; exists {
			return rule, true
		}
	}
	return ValidationRule{}, false
}

// GetGeminiValidationRule returns a specific Gemini validation rule
func (r *ServiceValidationRules) GetGeminiValidationRule(endpoint string) (ValidationRule, bool) {
	if rules, exists := r.validator.rules["gemini"]; exists {
		if rule, exists := rules[endpoint]; exists {
			return rule, true
		}
	}
	return ValidationRule{}, false
}

// GetGoogleDocsValidationRule returns a specific Google Docs validation rule
func (r *ServiceValidationRules) GetGoogleDocsValidationRule(endpoint string) (ValidationRule, bool) {
	if rules, exists := r.validator.rules["google_docs"]; exists {
		if rule, exists := rules[endpoint]; exists {
			return rule, true
		}
	}
	return ValidationRule{}, false
}

// ValidateJiraResponse validates a Jira API response
func (r *ServiceValidationRules) ValidateJiraResponse(endpoint string, response interface{}) *ValidationResult {
	return r.validator.ValidateResponse("jira", endpoint, response)
}

// ValidateGeminiResponse validates a Gemini API response
func (r *ServiceValidationRules) ValidateGeminiResponse(endpoint string, response interface{}) *ValidationResult {
	return r.validator.ValidateResponse("gemini", endpoint, response)
}

// ValidateGoogleDocsResponse validates a Google Docs API response
func (r *ServiceValidationRules) ValidateGoogleDocsResponse(endpoint string, response interface{}) *ValidationResult {
	return r.validator.ValidateResponse("google_docs", endpoint, response)
}

// CreateCustomRule creates a custom validation rule
func (r *ServiceValidationRules) CreateCustomRule(field, fieldType string, required bool) ValidationRule {
	return ValidationRule{
		Field:    field,
		Type:     fieldType,
		Required: required,
	}
}

// CreateStringRule creates a string validation rule with common constraints
func (r *ServiceValidationRules) CreateStringRule(field string, required bool, minLen, maxLen *int, pattern *string) ValidationRule {
	return ValidationRule{
		Field:     field,
		Type:      "string",
		Required:  required,
		MinLength: minLen,
		MaxLength: maxLen,
		Pattern:   pattern,
	}
}

// CreateNumberRule creates a number validation rule with common constraints
func (r *ServiceValidationRules) CreateNumberRule(field string, required bool, minVal, maxVal *float64) ValidationRule {
	return ValidationRule{
		Field:    field,
		Type:     "number",
		Required: required,
		MinValue: minVal,
		MaxValue: maxVal,
	}
}

// CreateEnumRule creates an enum validation rule
func (r *ServiceValidationRules) CreateEnumRule(field string, required bool, enumValues []string) ValidationRule {
	return ValidationRule{
		Field:    field,
		Type:     "string",
		Required: required,
		Enum:     enumValues,
	}
}

// CreateArrayRule creates an array validation rule
func (r *ServiceValidationRules) CreateArrayRule(field string, required bool, minLen, maxLen *int) ValidationRule {
	return ValidationRule{
		Field:     field,
		Type:      "array",
		Required:  required,
		MinLength: minLen,
		MaxLength: maxLen,
	}
}

// CreateObjectRule creates an object validation rule with nested rules
func (r *ServiceValidationRules) CreateObjectRule(field string, required bool, nested map[string]ValidationRule) ValidationRule {
	return ValidationRule{
		Field:    field,
		Type:     "object",
		Required: required,
		Nested:   nested,
	}
}

// AddCustomValidationRule adds a custom validation rule for a specific service and endpoint
func (r *ServiceValidationRules) AddCustomValidationRule(service, endpoint string, rule ValidationRule) {
	r.validator.RegisterRule(service, endpoint, rule)
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

// ValidationSummary provides a summary of validation results across services
type ValidationSummary struct {
	TotalValidations int                            `json:"total_validations"`
	SuccessfulCount  int                            `json:"successful_count"`
	FailedCount      int                            `json:"failed_count"`
	WarningCount     int                            `json:"warning_count"`
	ServiceResults   map[string]ServiceSummary      `json:"service_results"`
	MostCommonErrors []ValidationError              `json:"most_common_errors"`
	GeneratedAt      string                         `json:"generated_at"`
}

// ServiceSummary provides validation summary for a specific service
type ServiceSummary struct {
	ServiceName     string             `json:"service_name"`
	TotalCalls      int                `json:"total_calls"`
	SuccessfulCalls int                `json:"successful_calls"`
	FailedCalls     int                `json:"failed_calls"`
	WarningCalls    int                `json:"warning_calls"`
	EndpointResults map[string]int     `json:"endpoint_results"`
	CommonErrors    []ValidationError  `json:"common_errors"`
}

// ValidationMetrics tracks validation metrics over time
type ValidationMetrics struct {
	summary         *ValidationSummary
	serviceResults  map[string]*ServiceSummary
	recentResults   []ValidationResult
	maxResults      int
}

// NewValidationMetrics creates a new validation metrics tracker
func NewValidationMetrics(maxResults int) *ValidationMetrics {
	return &ValidationMetrics{
		summary: &ValidationSummary{
			ServiceResults: make(map[string]ServiceSummary),
		},
		serviceResults: make(map[string]*ServiceSummary),
		recentResults:  make([]ValidationResult, 0, maxResults),
		maxResults:     maxResults,
	}
}

// RecordValidation records a validation result
func (m *ValidationMetrics) RecordValidation(service, endpoint string, result *ValidationResult) {
	// Add to recent results
	m.recentResults = append(m.recentResults, *result)
	if len(m.recentResults) > m.maxResults {
		m.recentResults = m.recentResults[1:]
	}
	
	// Update summary
	m.summary.TotalValidations++
	if result.Valid {
		m.summary.SuccessfulCount++
	} else {
		m.summary.FailedCount++
	}
	if len(result.Warnings) > 0 {
		m.summary.WarningCount++
	}
	
	// Update service summary
	if _, exists := m.serviceResults[service]; !exists {
		m.serviceResults[service] = &ServiceSummary{
			ServiceName:     service,
			EndpointResults: make(map[string]int),
		}
	}
	
	serviceSummary := m.serviceResults[service]
	serviceSummary.TotalCalls++
	if result.Valid {
		serviceSummary.SuccessfulCalls++
	} else {
		serviceSummary.FailedCalls++
	}
	if len(result.Warnings) > 0 {
		serviceSummary.WarningCalls++
	}
	serviceSummary.EndpointResults[endpoint]++
}

// GetSummary returns the current validation summary
func (m *ValidationMetrics) GetSummary() *ValidationSummary {
	// Update service results in summary
	for _, serviceSummary := range m.serviceResults {
		m.summary.ServiceResults[serviceSummary.ServiceName] = *serviceSummary
	}
	
	m.summary.GeneratedAt = time.Now().Format(time.RFC3339)
	return m.summary
}

// GetRecentResults returns recent validation results
func (m *ValidationMetrics) GetRecentResults() []ValidationResult {
	return m.recentResults
}

// Reset resets all validation metrics
func (m *ValidationMetrics) Reset() {
	m.summary = &ValidationSummary{
		ServiceResults: make(map[string]ServiceSummary),
	}
	m.serviceResults = make(map[string]*ServiceSummary)
	m.recentResults = make([]ValidationResult, 0, m.maxResults)
}