package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/company/eesa/pkg/utils"
)

// ValidationRule represents a validation rule for API responses
type ValidationRule struct {
	Field       string                 `json:"field"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	MinLength   *int                   `json:"min_length,omitempty"`
	MaxLength   *int                   `json:"max_length,omitempty"`
	Pattern     *string                `json:"pattern,omitempty"`
	MinValue    *float64               `json:"min_value,omitempty"`
	MaxValue    *float64               `json:"max_value,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	CustomRules []string               `json:"custom_rules,omitempty"`
	Nested      map[string]ValidationRule `json:"nested,omitempty"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid      bool              `json:"valid"`
	Errors     []ValidationError `json:"errors"`
	Warnings   []ValidationError `json:"warnings"`
	FieldCount int               `json:"field_count"`
	ValidatedAt time.Time        `json:"validated_at"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"`
	Message  string      `json:"message"`
	Value    interface{} `json:"value,omitempty"`
	Expected interface{} `json:"expected,omitempty"`
}

// APIResponseValidator provides validation for API responses
type APIResponseValidator struct {
	logger utils.Logger
	rules  map[string]map[string]ValidationRule // service -> endpoint -> rules
}

// NewAPIResponseValidator creates a new API response validator
func NewAPIResponseValidator(logger utils.Logger) *APIResponseValidator {
	return &APIResponseValidator{
		logger: logger,
		rules:  make(map[string]map[string]ValidationRule),
	}
}

// RegisterRule registers a validation rule for a specific service and endpoint
func (v *APIResponseValidator) RegisterRule(service, endpoint string, rule ValidationRule) {
	if v.rules[service] == nil {
		v.rules[service] = make(map[string]ValidationRule)
	}
	v.rules[service][endpoint] = rule
	
	v.logger.Debug("Registered validation rule",
		utils.NewField("service", service),
		utils.NewField("endpoint", endpoint),
		utils.NewField("field", rule.Field),
	)
}

// RegisterRules registers multiple validation rules for a service
func (v *APIResponseValidator) RegisterRules(service string, rules map[string]ValidationRule) {
	if v.rules[service] == nil {
		v.rules[service] = make(map[string]ValidationRule)
	}
	
	for endpoint, rule := range rules {
		v.rules[service][endpoint] = rule
	}
	
	v.logger.Info("Registered validation rules for service",
		utils.NewField("service", service),
		utils.NewField("rule_count", len(rules)),
	)
}

// ValidateResponse validates an API response against registered rules
func (v *APIResponseValidator) ValidateResponse(service, endpoint string, response interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationError{},
		ValidatedAt: time.Now(),
	}
	
	// Get validation rules for the service and endpoint
	serviceRules, exists := v.rules[service]
	if !exists {
		v.logger.Debug("No validation rules found for service", utils.NewField("service", service))
		return result
	}
	
	rule, exists := serviceRules[endpoint]
	if !exists {
		v.logger.Debug("No validation rules found for endpoint",
			utils.NewField("service", service),
			utils.NewField("endpoint", endpoint),
		)
		return result
	}
	
	// Perform validation
	v.validateField("", response, rule, result)
	
	// Log validation results
	if len(result.Errors) > 0 {
		result.Valid = false
		v.logger.Warn("API response validation failed",
			utils.NewField("service", service),
			utils.NewField("endpoint", endpoint),
			utils.NewField("error_count", len(result.Errors)),
			utils.NewField("warning_count", len(result.Warnings)),
		)
	} else {
		v.logger.Debug("API response validation passed",
			utils.NewField("service", service),
			utils.NewField("endpoint", endpoint),
			utils.NewField("field_count", result.FieldCount),
		)
	}
	
	return result
}

// ValidateHTTPResponse validates an HTTP response with status code and headers
func (v *APIResponseValidator) ValidateHTTPResponse(service, endpoint string, resp *http.Response, body []byte) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationError{},
		ValidatedAt: time.Now(),
	}
	
	// Validate status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Errors = append(result.Errors, ValidationError{
			Field:    "status_code",
			Type:     "http_status",
			Message:  fmt.Sprintf("Invalid HTTP status code: %d", resp.StatusCode),
			Value:    resp.StatusCode,
			Expected: "2xx",
		})
		result.Valid = false
	}
	
	// Validate content type for JSON responses
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") && len(body) > 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:    "content_type",
			Type:     "http_header",
			Message:  "Expected JSON content type",
			Value:    contentType,
			Expected: "application/json",
		})
	}
	
	// Parse and validate JSON body if present
	if len(body) > 0 && strings.Contains(contentType, "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "json_body",
				Type:    "json_parse",
				Message: fmt.Sprintf("Invalid JSON: %s", err.Error()),
				Value:   string(body),
			})
			result.Valid = false
		} else {
			// Validate the parsed JSON data
			bodyResult := v.ValidateResponse(service, endpoint, jsonData)
			result.Errors = append(result.Errors, bodyResult.Errors...)
			result.Warnings = append(result.Warnings, bodyResult.Warnings...)
			result.FieldCount = bodyResult.FieldCount
			if !bodyResult.Valid {
				result.Valid = false
			}
		}
	}
	
	return result
}

// validateField validates a single field against its validation rule
func (v *APIResponseValidator) validateField(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	result.FieldCount++
	
	// Handle nil values
	if value == nil {
		if rule.Required {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPath,
				Type:    "required",
				Message: "Required field is missing or null",
				Value:   nil,
			})
		}
		return
	}
	
	// Get the actual value and type
	val := reflect.ValueOf(value)
	actualType := val.Type()
	
	// Handle different value types
	switch rule.Type {
	case "string":
		v.validateString(fieldPath, value, rule, result)
	case "number", "integer", "float":
		v.validateNumber(fieldPath, value, rule, result)
	case "boolean":
		v.validateBoolean(fieldPath, value, rule, result)
	case "array":
		v.validateArray(fieldPath, value, rule, result)
	case "object":
		v.validateObject(fieldPath, value, rule, result)
	case "timestamp", "datetime":
		v.validateTimestamp(fieldPath, value, rule, result)
	default:
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   fieldPath,
			Type:    "unknown_type",
			Message: fmt.Sprintf("Unknown validation type: %s", rule.Type),
			Value:   actualType.String(),
		})
	}
	
	// Validate enum values
	if len(rule.Enum) > 0 {
		v.validateEnum(fieldPath, value, rule, result)
	}
	
	// Apply custom validation rules
	for _, customRule := range rule.CustomRules {
		v.applyCustomRule(fieldPath, value, customRule, result)
	}
}

// validateString validates string values
func (v *APIResponseValidator) validateString(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	str, ok := value.(string)
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "type_mismatch",
			Message:  "Expected string type",
			Value:    value,
			Expected: "string",
		})
		return
	}
	
	// Check length constraints
	if rule.MinLength != nil && len(str) < *rule.MinLength {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "min_length",
			Message:  fmt.Sprintf("String length %d is below minimum %d", len(str), *rule.MinLength),
			Value:    len(str),
			Expected: *rule.MinLength,
		})
	}
	
	if rule.MaxLength != nil && len(str) > *rule.MaxLength {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "max_length",
			Message:  fmt.Sprintf("String length %d exceeds maximum %d", len(str), *rule.MaxLength),
			Value:    len(str),
			Expected: *rule.MaxLength,
		})
	}
	
	// Check pattern matching
	if rule.Pattern != nil {
		if matched, err := regexp.MatchString(*rule.Pattern, str); err != nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   fieldPath,
				Type:    "pattern_error",
				Message: fmt.Sprintf("Invalid regex pattern: %s", err.Error()),
				Value:   *rule.Pattern,
			})
		} else if !matched {
			result.Errors = append(result.Errors, ValidationError{
				Field:    fieldPath,
				Type:     "pattern_mismatch",
				Message:  fmt.Sprintf("String does not match pattern: %s", *rule.Pattern),
				Value:    str,
				Expected: *rule.Pattern,
			})
		}
	}
}

// validateNumber validates numeric values
func (v *APIResponseValidator) validateNumber(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	var num float64
	var ok bool
	
	switch v := value.(type) {
	case float64:
		num = v
		ok = true
	case float32:
		num = float64(v)
		ok = true
	case int:
		num = float64(v)
		ok = true
	case int32:
		num = float64(v)
		ok = true
	case int64:
		num = float64(v)
		ok = true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			num = parsed
			ok = true
		}
	}
	
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "type_mismatch",
			Message:  "Expected numeric type",
			Value:    value,
			Expected: "number",
		})
		return
	}
	
	// Check value constraints
	if rule.MinValue != nil && num < *rule.MinValue {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "min_value",
			Message:  fmt.Sprintf("Value %g is below minimum %g", num, *rule.MinValue),
			Value:    num,
			Expected: *rule.MinValue,
		})
	}
	
	if rule.MaxValue != nil && num > *rule.MaxValue {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "max_value",
			Message:  fmt.Sprintf("Value %g exceeds maximum %g", num, *rule.MaxValue),
			Value:    num,
			Expected: *rule.MaxValue,
		})
	}
	
	// Validate integer type specifically
	if rule.Type == "integer" {
		if num != float64(int64(num)) {
			result.Errors = append(result.Errors, ValidationError{
				Field:    fieldPath,
				Type:     "type_mismatch",
				Message:  "Expected integer value",
				Value:    num,
				Expected: "integer",
			})
		}
	}
}

// validateBoolean validates boolean values
func (v *APIResponseValidator) validateBoolean(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	if _, ok := value.(bool); !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "type_mismatch",
			Message:  "Expected boolean type",
			Value:    value,
			Expected: "boolean",
		})
	}
}

// validateArray validates array values
func (v *APIResponseValidator) validateArray(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "type_mismatch",
			Message:  "Expected array type",
			Value:    value,
			Expected: "array",
		})
		return
	}
	
	// Check length constraints
	length := val.Len()
	if rule.MinLength != nil && length < *rule.MinLength {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "min_length",
			Message:  fmt.Sprintf("Array length %d is below minimum %d", length, *rule.MinLength),
			Value:    length,
			Expected: *rule.MinLength,
		})
	}
	
	if rule.MaxLength != nil && length > *rule.MaxLength {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "max_length",
			Message:  fmt.Sprintf("Array length %d exceeds maximum %d", length, *rule.MaxLength),
			Value:    length,
			Expected: *rule.MaxLength,
		})
	}
}

// validateObject validates object values
func (v *APIResponseValidator) validateObject(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Map && val.Kind() != reflect.Struct {
		// Try to handle interface{} containing a map
		if mapValue, ok := value.(map[string]interface{}); ok {
			v.validateObjectFields(fieldPath, mapValue, rule, result)
			return
		}
		
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "type_mismatch",
			Message:  "Expected object type",
			Value:    value,
			Expected: "object",
		})
		return
	}
	
	// Validate nested fields if rules are provided
	if mapValue, ok := value.(map[string]interface{}); ok {
		v.validateObjectFields(fieldPath, mapValue, rule, result)
	}
}

// validateObjectFields validates fields within an object
func (v *APIResponseValidator) validateObjectFields(fieldPath string, obj map[string]interface{}, rule ValidationRule, result *ValidationResult) {
	for nestedField, nestedRule := range rule.Nested {
		nestedFieldPath := fieldPath
		if fieldPath != "" {
			nestedFieldPath += "."
		}
		nestedFieldPath += nestedField
		
		if nestedValue, exists := obj[nestedField]; exists {
			v.validateField(nestedFieldPath, nestedValue, nestedRule, result)
		} else if nestedRule.Required {
			result.Errors = append(result.Errors, ValidationError{
				Field:   nestedFieldPath,
				Type:    "required",
				Message: "Required nested field is missing",
				Value:   nil,
			})
		}
	}
}

// validateTimestamp validates timestamp values
func (v *APIResponseValidator) validateTimestamp(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	str, ok := value.(string)
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "type_mismatch",
			Message:  "Expected timestamp string",
			Value:    value,
			Expected: "timestamp string",
		})
		return
	}
	
	// Try to parse common timestamp formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-0700",
		"2006-01-02 15:04:05",
	}
	
	var parsed bool
	for _, format := range formats {
		if _, err := time.Parse(format, str); err == nil {
			parsed = true
			break
		}
	}
	
	if !parsed {
		result.Errors = append(result.Errors, ValidationError{
			Field:    fieldPath,
			Type:     "timestamp_format",
			Message:  "Invalid timestamp format",
			Value:    str,
			Expected: "RFC3339 or similar format",
		})
	}
}

// validateEnum validates enum values
func (v *APIResponseValidator) validateEnum(fieldPath string, value interface{}, rule ValidationRule, result *ValidationResult) {
	valueStr := fmt.Sprintf("%v", value)
	
	for _, enumValue := range rule.Enum {
		if valueStr == enumValue {
			return
		}
	}
	
	result.Errors = append(result.Errors, ValidationError{
		Field:    fieldPath,
		Type:     "enum_mismatch",
		Message:  fmt.Sprintf("Value is not in allowed enum values"),
		Value:    valueStr,
		Expected: rule.Enum,
	})
}

// applyCustomRule applies custom validation rules
func (v *APIResponseValidator) applyCustomRule(fieldPath string, value interface{}, customRule string, result *ValidationResult) {
	switch customRule {
	case "non_empty":
		if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPath,
				Type:    "custom_non_empty",
				Message: "String cannot be empty or whitespace only",
				Value:   str,
			})
		}
	case "email":
		if str, ok := value.(string); ok {
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			if matched, _ := regexp.MatchString(emailPattern, str); !matched {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldPath,
					Type:    "custom_email",
					Message: "Invalid email format",
					Value:   str,
				})
			}
		}
	case "url":
		if str, ok := value.(string); ok {
			urlPattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
			if matched, _ := regexp.MatchString(urlPattern, str); !matched {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldPath,
					Type:    "custom_url",
					Message: "Invalid URL format",
					Value:   str,
				})
			}
		}
	case "uuid":
		if str, ok := value.(string); ok {
			uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
			if matched, _ := regexp.MatchString(uuidPattern, str); !matched {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldPath,
					Type:    "custom_uuid",
					Message: "Invalid UUID format",
					Value:   str,
				})
			}
		}
	case "positive":
		if num, ok := v.convertToFloat64(value); ok && num <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPath,
				Type:    "custom_positive",
				Message: "Value must be positive",
				Value:   num,
			})
		}
	default:
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   fieldPath,
			Type:    "unknown_custom_rule",
			Message: fmt.Sprintf("Unknown custom rule: %s", customRule),
			Value:   customRule,
		})
	}
}

// convertToFloat64 attempts to convert a value to float64
func (v *APIResponseValidator) convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

// GetRules returns all registered validation rules
func (v *APIResponseValidator) GetRules() map[string]map[string]ValidationRule {
	return v.rules
}

// HasRules checks if validation rules exist for a service and endpoint
func (v *APIResponseValidator) HasRules(service, endpoint string) bool {
	serviceRules, exists := v.rules[service]
	if !exists {
		return false
	}
	_, exists = serviceRules[endpoint]
	return exists
}

// ClearRules clears all validation rules for a service
func (v *APIResponseValidator) ClearRules(service string) {
	delete(v.rules, service)
	v.logger.Info("Cleared validation rules for service", utils.NewField("service", service))
}