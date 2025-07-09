package gdocs

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
	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}

	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	client := NewClient(cfg, authManager, logger)

	assert.NotNil(t, client)
	assert.Equal(t, BaseURL, client.baseURL)
	assert.Equal(t, DriveBaseURL, client.driveBaseURL)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.auth)
	assert.NotNil(t, client.rateLimiter)
	assert.NotNil(t, client.retryConfig)
	assert.Equal(t, logger, client.logger)
}

func TestClient_formatMetadata(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	metadata := map[string]interface{}{
		"generated_at":    time.Date(2023, 1, 15, 14, 30, 0, 0, time.UTC),
		"model":          "gemini-pro",
		"tokens_used":    150,
		"activity_count": 25,
		"time_range":     "January 1-7, 2023",
	}

	result := client.formatMetadata(metadata)

	assert.Contains(t, result, "Generated: January 15, 2023 at 2:30 PM")
	assert.Contains(t, result, "AI Model: gemini-pro")
	assert.Contains(t, result, "Tokens Used: 150")
	assert.Contains(t, result, "Activities Analyzed: 25")
	assert.Contains(t, result, "Time Period: January 1-7, 2023")
}

func TestClient_formatMetadata_Empty(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	result := client.formatMetadata(map[string]interface{}{})
	assert.Equal(t, "", result)
}

func TestClient_handleErrorResponse(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
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
			responseBody:  `{"error":{"message":"Invalid credentials"}}`,
			expectedError: utils.ErrorCodeAPIUnauthorized,
			message:       "Invalid credentials",
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
			responseBody:  `{"error":{"message":"Document not found"}}`,
			expectedError: utils.ErrorCodeAPINotFound,
			message:       "Document not found",
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
			expectedError: utils.ErrorCodeGoogleError,
			message:       "I'm a teapot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
			}

			err := client.handleErrorResponse(resp, "Test error")

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
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	// Set up credentials for the test
	err := authManager.GetCredentialStore().SetGoogleCredentials(security.GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
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
	assert.NotEmpty(t, req.Header.Get("Authorization"))

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}

func TestClient_createDriveRequest(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	// Set up credentials for the test
	err := authManager.GetCredentialStore().SetGoogleCredentials(security.GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	})
	require.NoError(t, err)

	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()
	method := "POST"
	endpoint := "/drive/v3/files/test/permissions"
	body := []byte(`{"test": "data"}`)

	req, err := client.createDriveRequest(ctx, method, endpoint, body)
	require.NoError(t, err)

	assert.Equal(t, method, req.Method)
	assert.Equal(t, client.driveBaseURL+endpoint, req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.NotEmpty(t, req.Header.Get("Authorization"))

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}

func TestHelperFunctions(t *testing.T) {
	t.Run("boolPtr", func(t *testing.T) {
		val := true
		ptr := boolPtr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("stringPtr", func(t *testing.T) {
		val := "test"
		ptr := stringPtr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("int32Ptr", func(t *testing.T) {
		val := int32(42)
		ptr := int32Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("float64Ptr", func(t *testing.T) {
		val := 3.14
		ptr := float64Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})
}

func TestGoogleAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    GoogleAPIError
		expected string
	}{
		{
			name: "with message",
			error: GoogleAPIError{
				Code:    400,
				Message: "Invalid request",
				Status:  "INVALID_ARGUMENT",
			},
			expected: "Invalid request",
		},
		{
			name: "without message",
			error: GoogleAPIError{
				Code:   500,
				Status: "INTERNAL_ERROR",
			},
			expected: "Unknown Google API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoogleErrorResponse_Error(t *testing.T) {
	errorResponse := GoogleErrorResponse{
		ErrorInfo: GoogleAPIError{
			Code:    400,
			Message: "Invalid request",
			Status:  "INVALID_ARGUMENT",
		},
	}

	result := errorResponse.Error()
	assert.Equal(t, "Invalid request", result)
}

// Mock HTTP server for integration tests
func createMockGoogleDocsServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/documents":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(DocumentResponse{
				DocumentID: "test_document_id",
				Title:      "Test Document",
				Body: &Body{
					Content: []StructuralElement{
						{
							StartIndex: 0,
							EndIndex:   1,
							Paragraph: &Paragraph{
								Elements: []ParagraphElement{
									{
										StartIndex: 0,
										EndIndex:   1,
										TextRun: &TextRun{
											Content: "\n",
										},
									},
								},
							},
						},
					},
				},
			})
		case r.Method == "POST" && strings.Contains(r.URL.Path, "batchUpdate"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(BatchUpdateResponse{
				DocumentID: "test_document_id",
				Replies: []Reply{
					{
						InsertText: &InsertTextReply{},
					},
				},
			})
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/v1/documents/"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(DocumentResponse{
				DocumentID: "test_document_id",
				Title:      "Test Document",
				Body: &Body{
					Content: []StructuralElement{
						{
							StartIndex: 0,
							EndIndex:   25,
							Paragraph: &Paragraph{
								Elements: []ParagraphElement{
									{
										StartIndex: 0,
										EndIndex:   25,
										TextRun: &TextRun{
											Content: "This is test content.\n",
										},
									},
								},
							},
						},
					},
				},
			})
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/drive/v3/files/") && strings.Contains(r.URL.Path, "/permissions"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "permission_id",
				"type": "user",
				"role": "reader",
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
	server := createMockGoogleDocsServer(t)
	defer server.Close()

	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	// Set up mock credentials
	err := authManager.GetCredentialStore().SetGoogleCredentials(security.GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	})
	require.NoError(t, err)

	// Override URLs for testing
	client := NewClient(cfg, authManager, logger)
	client.baseURL = server.URL
	client.driveBaseURL = server.URL

	ctx := context.Background()

	// Test document creation
	doc, err := client.CreateDocument(ctx, "Test Executive Summary", "This is test content for the executive summary.")
	assert.NoError(t, err)
	assert.Equal(t, "test_document_id", doc.DocumentID)
	assert.Equal(t, "Test Document", doc.Title)

	// Test document retrieval
	retrievedDoc, err := client.GetDocument(ctx, "test_document_id")
	assert.NoError(t, err)
	assert.Equal(t, "test_document_id", retrievedDoc.DocumentID)
	assert.Equal(t, "Test Document", retrievedDoc.Title)

	// Test document update
	updateRequests := []Request{
		{
			InsertText: &InsertTextRequest{
				Text:     "Additional content",
				Location: &Location{Index: 1},
			},
		},
	}
	updateResponse, err := client.UpdateDocument(ctx, "test_document_id", updateRequests)
	assert.NoError(t, err)
	assert.Equal(t, "test_document_id", updateResponse.DocumentID)

	// Test document sharing
	err = client.ShareDocument(ctx, "test_document_id", []string{"user@example.com"}, "reader")
	assert.NoError(t, err)

	// Test executive summary document creation
	metadata := map[string]interface{}{
		"generated_at":    time.Now(),
		"model":          "gemini-pro",
		"tokens_used":    150,
		"activity_count": 25,
	}
	execDoc, err := client.CreateExecutiveSummaryDocument(ctx, "Weekly Executive Summary", "This is a comprehensive summary of activities.", metadata)
	assert.NoError(t, err)
	assert.Equal(t, "test_document_id", execDoc.DocumentID)

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}

func TestClient_CreateDocument_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()

	// Test with empty title
	_, err := client.CreateDocument(ctx, "", "content")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "title is required")
}

func TestClient_UpdateDocument_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()

	// Test with empty document ID
	_, err := client.UpdateDocument(ctx, "", []Request{})
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "Document ID is required")

	// Test with empty requests
	_, err = client.UpdateDocument(ctx, "doc_id", []Request{})
	require.Error(t, err)
	appErr, ok = err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "At least one request is required")
}

func TestClient_GetDocument_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()

	// Test with empty document ID
	_, err := client.GetDocument(ctx, "")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "Document ID is required")
}

func TestClient_ShareDocument_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()

	// Test with empty document ID
	err := client.ShareDocument(ctx, "", []string{"user@example.com"}, "reader")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "Document ID is required")

	// Test with empty emails
	err = client.ShareDocument(ctx, "doc_id", []string{}, "reader")
	require.Error(t, err)
	appErr, ok = err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "At least one email is required")
}

func TestClient_CreateExecutiveSummaryDocument_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)

	ctx := context.Background()

	// Test with empty title
	_, err := client.CreateExecutiveSummaryDocument(ctx, "", "summary", nil)
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "title is required")

	// Test with empty summary
	_, err = client.CreateExecutiveSummaryDocument(ctx, "title", "", nil)
	require.Error(t, err)
	appErr, ok = err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeDataInvalid, appErr.Code)
	assert.Contains(t, appErr.Message, "Summary content is required")
}

func TestClient_DocumentNotFound(t *testing.T) {
	// Create mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/v1/documents/") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(GoogleErrorResponse{
				ErrorInfo: GoogleAPIError{
					Code:    404,
					Message: "Document not found",
					Status:  "NOT_FOUND",
				},
			})
		}
	}))
	defer server.Close()

	logger := utils.NewMockLogger()
	cfg := &config.Config{}
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)

	err := authManager.GetCredentialStore().SetGoogleCredentials(security.GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	})
	require.NoError(t, err)

	client := NewClient(cfg, authManager, logger)
	client.baseURL = server.URL

	ctx := context.Background()

	_, err = client.GetDocument(ctx, "nonexistent_doc_id")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeAPINotFound, appErr.Code)
	assert.Equal(t, "nonexistent_doc_id", appErr.Context.Extra["document_id"])

	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}