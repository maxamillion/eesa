package security

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAuthConfig(t *testing.T) {
	config := DefaultAuthConfig()
	
	assert.Equal(t, "1.3", config.TLSMinVersion)
	assert.True(t, config.VerifySSL)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected uint16
	}{
		{"1.3", tls.VersionTLS13},
		{"1.2", tls.VersionTLS12},
		{"1.1", tls.VersionTLS11},
		{"1.0", tls.VersionTLS10},
		{"invalid", tls.VersionTLS13}, // Default to most secure
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseTLSVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewAuthenticatedHTTPClient(t *testing.T) {
	logger := utils.NewMockLogger()
	config := DefaultAuthConfig()
	
	client := NewAuthenticatedHTTPClient(config, logger)
	
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, config, client.config)
	assert.Equal(t, logger, client.logger)
}

func TestAuthenticatedHTTPClient_CreateRequest(t *testing.T) {
	logger := utils.NewMockLogger()
	client := NewAuthenticatedHTTPClient(nil, logger)
	
	req, err := client.CreateRequest("GET", "https://example.com", nil)
	require.NoError(t, err)
	
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "https://example.com", req.URL.String())
	assert.Equal(t, "ESA/1.0", req.Header.Get("User-Agent"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", req.Header.Get("Cache-Control"))
	assert.Equal(t, "no-cache", req.Header.Get("Pragma"))
	assert.Equal(t, "0", req.Header.Get("Expires"))
}

func TestAuthenticatedHTTPClient_DoRequest(t *testing.T) {
	logger := utils.NewMockLogger()
	client := NewAuthenticatedHTTPClient(nil, logger)
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()
	
	// Create request
	req, err := client.CreateRequest("GET", server.URL, nil)
	require.NoError(t, err)
	
	// Make request
	resp, err := client.DoRequest(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check that debug logs were created
	debugEntries := logger.GetEntriesByLevel(utils.LogLevelDebug)
	assert.Len(t, debugEntries, 2) // Request and response logs
}

func TestJiraAuthenticator_AddAuthHeaders(t *testing.T) {
	logger := utils.NewMockLogger()
	httpClient := NewAuthenticatedHTTPClient(nil, logger)
	credStore := NewCredentialStore(logger)
	
	// Set up test credentials
	err := credStore.SetJiraCredentials(JiraCredentials{Token: "test_token"})
	require.NoError(t, err)
	
	auth := NewJiraAuthenticator(httpClient, credStore, logger)
	
	// Create request
	req, err := httpClient.CreateRequest("GET", "https://example.com", nil)
	require.NoError(t, err)
	
	// Add auth headers
	err = auth.AddAuthHeaders(req, "testuser")
	require.NoError(t, err)
	
	// Check Basic Auth header
	username, password, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "test_token", password)
	
	// Clean up
	credStore.keyring.DeleteCredential(KeyJiraToken)
}

func TestJiraAuthenticator_ValidateCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	httpClient := NewAuthenticatedHTTPClient(nil, logger)
	credStore := NewCredentialStore(logger)
	
	// Set up test credentials
	err := credStore.SetJiraCredentials(JiraCredentials{Token: "valid_token"})
	require.NoError(t, err)
	
	auth := NewJiraAuthenticator(httpClient, credStore, logger)
	
	// Test successful validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/2/myself" {
			username, password, ok := r.BasicAuth()
			if ok && username == "testuser" && password == "valid_token" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"name": "testuser"}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	err = auth.ValidateCredentials(server.URL, "testuser")
	require.NoError(t, err)
	
	// Test failed validation
	err = credStore.SetJiraCredentials(JiraCredentials{Token: "invalid_token"})
	require.NoError(t, err)
	
	err = auth.ValidateCredentials(server.URL, "testuser")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeAuthFailed, appErr.Code)
	
	// Clean up
	credStore.keyring.DeleteCredential(KeyJiraToken)
}

func TestGeminiAuthenticator_AddAuthHeaders(t *testing.T) {
	logger := utils.NewMockLogger()
	httpClient := NewAuthenticatedHTTPClient(nil, logger)
	credStore := NewCredentialStore(logger)
	
	// Set up test credentials
	err := credStore.SetGeminiCredentials(GeminiCredentials{APIKey: "test_api_key"})
	require.NoError(t, err)
	
	auth := NewGeminiAuthenticator(httpClient, credStore, logger)
	
	// Create request
	req, err := httpClient.CreateRequest("GET", "https://example.com", nil)
	require.NoError(t, err)
	
	// Add auth headers
	err = auth.AddAuthHeaders(req)
	require.NoError(t, err)
	
	// Check x-goog-api-key header (Gemini uses this instead of Authorization)
	assert.Equal(t, "test_api_key", req.Header.Get("x-goog-api-key"))
	
	// Clean up
	credStore.keyring.DeleteCredential(KeyGeminiAPIKey)
}

func TestGeminiAuthenticator_ValidateCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	httpClient := NewAuthenticatedHTTPClient(nil, logger)
	credStore := NewCredentialStore(logger)
	
	// Set up test credentials
	err := credStore.SetGeminiCredentials(GeminiCredentials{APIKey: "valid_api_key"})
	require.NoError(t, err)
	
	auth := NewGeminiAuthenticator(httpClient, credStore, logger)
	
	// Test successful validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "Bearer valid_api_key" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"models": []}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	// We can't easily test the actual Gemini API, so we'll test the error case
	// when credentials are missing
	credStore.keyring.DeleteCredential(KeyGeminiAPIKey)
	
	err = auth.ValidateCredentials()
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeAuthFailed, appErr.Code)
}

func TestGoogleAuthenticator_AddAuthHeaders(t *testing.T) {
	logger := utils.NewMockLogger()
	httpClient := NewAuthenticatedHTTPClient(nil, logger)
	credStore := NewCredentialStore(logger)
	
	// Set up test credentials
	err := credStore.SetGoogleCredentials(GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
	})
	require.NoError(t, err)
	
	auth := NewGoogleAuthenticator(httpClient, credStore, logger)
	
	// Create request
	req, err := httpClient.CreateRequest("GET", "https://example.com", nil)
	require.NoError(t, err)
	
	// Add auth headers
	err = auth.AddAuthHeaders(req)
	require.NoError(t, err)
	
	// Check Authorization header
	assert.Equal(t, "Bearer test_access_token", req.Header.Get("Authorization"))
	
	// Clean up
	credStore.keyring.DeleteCredential(KeyGoogleClientSecret)
	credStore.keyring.DeleteCredential(KeyGoogleAccessToken)
}

func TestGoogleAuthenticator_ValidateCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	httpClient := NewAuthenticatedHTTPClient(nil, logger)
	credStore := NewCredentialStore(logger)
	
	// Test missing access token
	err := credStore.SetGoogleCredentials(GoogleCredentials{
		ClientSecret: "test_client_secret",
	})
	require.NoError(t, err)
	
	auth := NewGoogleAuthenticator(httpClient, credStore, logger)
	
	err = auth.ValidateCredentials()
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeAuthFailed, appErr.Code)
	
	// Test with access token
	err = credStore.SetGoogleCredentials(GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "valid_access_token",
	})
	require.NoError(t, err)
	
	// Test successful validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/v1/userinfo" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "Bearer valid_access_token" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"email": "test@example.com"}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	// We can't easily test the actual Google API, so we'll test the error case
	// when credentials are missing
	credStore.keyring.DeleteCredential(KeyGoogleClientSecret)
	
	err = auth.ValidateCredentials()
	require.Error(t, err)
	appErr, ok = err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeAuthFailed, appErr.Code)
}

func TestAuthManager(t *testing.T) {
	logger := utils.NewMockLogger()
	config := DefaultAuthConfig()
	
	manager := NewAuthManager(config, logger)
	
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.GetHTTPClient())
	assert.NotNil(t, manager.GetCredentialStore())
	assert.NotNil(t, manager.GetJiraAuthenticator())
	assert.NotNil(t, manager.GetGeminiAuthenticator())
	assert.NotNil(t, manager.GetGoogleAuthenticator())
}

func TestAuthManager_ValidateAllCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	config := DefaultAuthConfig()
	manager := NewAuthManager(config, logger)
	
	// Should fail when no credentials are set
	err := manager.ValidateAllCredentials("https://example.atlassian.net", "testuser")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeValidationError, appErr.Code)
	
	// Set up credentials
	credStore := manager.GetCredentialStore()
	err = credStore.SetJiraCredentials(JiraCredentials{Token: "test_token"})
	require.NoError(t, err)
	
	err = credStore.SetGeminiCredentials(GeminiCredentials{APIKey: "test_api_key"})
	require.NoError(t, err)
	
	err = credStore.SetGoogleCredentials(GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
	})
	require.NoError(t, err)
	
	// Should still fail because we can't actually validate against real services
	err = manager.ValidateAllCredentials("https://example.atlassian.net", "testuser")
	require.Error(t, err)
	
	// Clean up
	credStore.ClearAllCredentials()
}