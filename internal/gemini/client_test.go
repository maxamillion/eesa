package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/company/eesa/internal/config"
	"github.com/company/eesa/internal/security"
	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	client := NewClient(cfg, authManager, logger)

	assert.NotNil(t, client)
	assert.Equal(t, BaseURL, client.baseURL)
	assert.Equal(t, "gemini-pro", client.model)
	assert.Equal(t, float32(0.7), client.temperature)
	assert.Equal(t, 2048, client.maxTokens)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.auth)
	assert.NotNil(t, client.rateLimiter)
	assert.NotNil(t, client.retryConfig)
	assert.Equal(t, logger, client.logger)
}

func TestClient_buildSummaryPrompt(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	activities := []models.Activity{
		{
			ID:      "1",
			Key:     "TEST-123",
			Summary: "Test issue",
			Type:    "Bug",
			Status:  "Done",
			Priority: "High",
			Project: models.Project{
				Key:  "TEST",
				Name: "Test Project",
			},
			Assignee: models.User{
				DisplayName: "John Doe",
			},
			Created:   time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Updated:   time.Date(2023, 1, 2, 15, 30, 0, 0, time.UTC),
			TimeSpent: 3600,
		},
	}

	prompt := client.buildSummaryPrompt(activities, "")

	assert.Contains(t, prompt, "executive assistant")
	assert.Contains(t, prompt, "TEST-123")
	assert.Contains(t, prompt, "Test issue")
	assert.Contains(t, prompt, "Bug")
	assert.Contains(t, prompt, "Done")
	assert.Contains(t, prompt, "High")
	assert.Contains(t, prompt, "John Doe")
	assert.Contains(t, prompt, "PROJECT: TEST")
	assert.Contains(t, prompt, "Total Issues: 1")
	assert.Contains(t, prompt, "Completed Issues: 1")
	assert.Contains(t, prompt, "Time Spent: 1h")
}

func TestClient_buildSummaryPrompt_WithCustomPrompt(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	activities := []models.Activity{
		{
			ID:      "1",
			Key:     "TEST-123",
			Summary: "Test issue",
		},
	}

	customPrompt := "Focus on security aspects"
	prompt := client.buildSummaryPrompt(activities, customPrompt)

	assert.Contains(t, prompt, "ADDITIONAL INSTRUCTIONS:")
	assert.Contains(t, prompt, "Focus on security aspects")
}

func TestClient_handleErrorResponse(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  utils.ErrorCode
		message        string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error":{"message":"Invalid API key"}}`,
			expectedError: utils.ErrorCodeAPIUnauthorized,
			message:       "Invalid API key",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			responseBody:  `{"error":{"message":"Access denied"}}`,
			expectedError: utils.ErrorCodeAPIUnauthorized,
			message:       "Access denied",
		},
		{
			name:          "not found",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error":{"message":"Model not found"}}`,
			expectedError: utils.ErrorCodeAPINotFound,
			message:       "Model not found",
		},
		{
			name:          "rate limit",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"error":{"message":"Rate limit exceeded"}}`,
			expectedError: utils.ErrorCodeAPIRateLimit,
			message:       "Rate limit exceeded",
		},
		{
			name:          "bad request",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"error":{"message":"Invalid request"}}`,
			expectedError: utils.ErrorCodeAPIBadRequest,
			message:       "Invalid request",
		},
		{
			name:          "server error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error":{"message":"Internal server error"}}`,
			expectedError: utils.ErrorCodeAPIServerError,
			message:       "Internal server error",
		},
		{
			name:          "other error",
			statusCode:    http.StatusTeapot,
			responseBody:  `{"error":{"message":"I'm a teapot"}}`,
			expectedError: utils.ErrorCodeGeminiError,
			message:       "I'm a teapot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
			}

			err := client.handleErrorResponse(resp, tt.message)

			require.Error(t, err)
			appErr, ok := err.(*utils.AppError)
			require.True(t, ok)
			assert.Equal(t, tt.expectedError, appErr.Code)
			assert.Equal(t, tt.message, appErr.Message)
			assert.Equal(t, tt.statusCode, appErr.Context.Extra["status_code"])
		})
	}
}

func TestClient_createRequest(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	// Set up credentials for the test
	err := authManager.GetCredentialStore().SetGeminiCredentials(security.GeminiCredentials{
		APIKey: "test_api_key",
	})
	require.NoError(t, err)

	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()
	method := "POST"
	endpoint := "/test"
	body := []byte(`{"test": "data"}`)

	req, err := client.createRequest(ctx, method, endpoint, body)
	require.NoError(t, err)

	assert.Equal(t, method, req.Method)
	assert.Equal(t, client.baseURL+endpoint, req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.NotEmpty(t, req.Header.Get("x-goog-api-key"))

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}

func TestHelperFunctions(t *testing.T) {
	t.Run("float32Ptr", func(t *testing.T) {
		val := float32(0.7)
		ptr := float32Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("int32Ptr", func(t *testing.T) {
		val := int32(40)
		ptr := int32Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})
}

func TestGeminiError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    GeminiError
		expected string
	}{
		{
			name: "with message",
			error: GeminiError{
				Code:    400,
				Message: "Invalid request",
				Status:  "INVALID_ARGUMENT",
			},
			expected: "Invalid request",
		},
		{
			name: "without message",
			error: GeminiError{
				Code:   500,
				Status: "INTERNAL_ERROR",
			},
			expected: "Unknown Gemini API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	errorResponse := ErrorResponse{
		ErrorInfo: GeminiError{
			Code:    400,
			Message: "Invalid request",
			Status:  "INVALID_ARGUMENT",
		},
	}

	result := errorResponse.Error()
	assert.Equal(t, "Invalid request", result)
}

// Mock HTTP server for integration tests
func createMockGeminiServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ModelsResponse{
				Models: []Model{
					{
						Name:             "models/gemini-pro",
						Version:          "001",
						DisplayName:      "Gemini Pro",
						Description:      "The Gemini Pro model",
						InputTokenLimit:  30720,
						OutputTokenLimit: 2048,
						SupportedGeneration: []string{"generateContent"},
					},
				},
			})
		case "/v1/models/gemini-pro:generateContent":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GenerateResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{
									Text: "## Executive Summary\n\nThis week, the team completed 5 high-priority issues including critical bug fixes and feature enhancements. The completion rate was 83%, showing strong progress toward our quarterly goals.",
								},
							},
						},
						FinishReason: FinishReasonStop,
						Index:        0,
						SafetyRatings: []SafetyRating{
							{
								Category:    SafetyCategoryHarassment,
								Probability: "NEGLIGIBLE",
							},
						},
					},
				},
				UsageMetadata: &UsageMetadata{
					PromptTokenCount:     150,
					CandidatesTokenCount: 45,
					TotalTokenCount:      195,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock server
	server := createMockGeminiServer(t)
	defer server.Close()

	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	// Set up mock credentials
	err := authManager.GetCredentialStore().SetGeminiCredentials(security.GeminiCredentials{
		APIKey: "test_api_key",
	})
	require.NoError(t, err)

	// Override base URL for testing
	client := NewClient(cfg, authManager, logger)
	client.baseURL = server.URL

	ctx := context.Background()

	// Test API key validation
	err = client.ValidateAPIKey(ctx)
	assert.NoError(t, err)

	// Test list models
	modelsResponse, err := client.ListModels(ctx)
	assert.NoError(t, err)
	assert.Len(t, modelsResponse.Models, 1)
	assert.Equal(t, "models/gemini-pro", modelsResponse.Models[0].Name)

	// Test summary generation
	activities := []models.Activity{
		{
			ID:       "1",
			Key:      "TEST-123",
			Summary:  "Fix critical authentication bug",
			Type:     "Bug",
			Status:   "Done",
			Priority: "Critical",
			Project: models.Project{
				Key:  "AUTH",
				Name: "Authentication Service",
			},
			Assignee: models.User{
				DisplayName: "Alice Johnson",
			},
			Created:   time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Updated:   time.Date(2023, 1, 2, 15, 30, 0, 0, time.UTC),
			TimeSpent: 7200, // 2 hours
		},
		{
			ID:       "2",
			Key:      "TEST-124",
			Summary:  "Implement user profile feature",
			Type:     "Story",
			Status:   "In Progress",
			Priority: "High",
			Project: models.Project{
				Key:  "USER",
				Name: "User Management",
			},
			Assignee: models.User{
				DisplayName: "Bob Smith",
			},
			Created:   time.Date(2023, 1, 2, 9, 0, 0, 0, time.UTC),
			Updated:   time.Date(2023, 1, 3, 14, 45, 0, 0, time.UTC),
			TimeSpent: 10800, // 3 hours
		},
	}

	summaryResponse, err := client.GenerateSummary(ctx, activities, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, summaryResponse.Summary)
	assert.Contains(t, summaryResponse.Summary, "Executive Summary")
	assert.Equal(t, "gemini-pro", summaryResponse.Model)
	assert.Equal(t, float32(0.7), summaryResponse.Temperature)
	assert.Equal(t, 195, summaryResponse.TokensUsed)
	assert.Len(t, summaryResponse.Activities, 2)

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}

func TestClient_GenerateSummary_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()

	// Test with empty activities
	_, err := client.GenerateSummary(ctx, []models.Activity{}, "")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "No activities provided")
}

func TestClient_GenerateContent_EmptyResponse(t *testing.T) {
	// Create mock server that returns empty candidates
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models/gemini-pro:generateContent" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GenerateResponse{
				Candidates: []Candidate{},
			})
		}
	}))
	defer server.Close()

	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	err := authManager.GetCredentialStore().SetGeminiCredentials(security.GeminiCredentials{
		APIKey: "test_api_key",
	})
	require.NoError(t, err)

	client := NewClient(cfg, authManager, logger)
	client.baseURL = server.URL

	ctx := context.Background()
	activities := []models.Activity{
		{
			ID:      "1",
			Key:     "TEST-123",
			Summary: "Test issue",
		},
	}

	_, err = client.GenerateSummary(ctx, activities, "")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeGeminiError, appErr.Code)
	assert.Contains(t, appErr.Message, "No content generated")

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}