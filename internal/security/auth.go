package security

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/company/eesa/pkg/utils"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	TLSMinVersion string `yaml:"tls_min_version"`
	VerifySSL     bool   `yaml:"verify_ssl"`
	Timeout       time.Duration
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		TLSMinVersion: "1.3",
		VerifySSL:     true,
		Timeout:       30 * time.Second,
	}
}

// AuthenticatedHTTPClient provides a secure HTTP client with authentication
type AuthenticatedHTTPClient struct {
	httpClient *http.Client
	config     *AuthConfig
	logger     utils.Logger
}

// NewAuthenticatedHTTPClient creates a new authenticated HTTP client
func NewAuthenticatedHTTPClient(config *AuthConfig, logger utils.Logger) *AuthenticatedHTTPClient {
	if config == nil {
		config = DefaultAuthConfig()
	}
	
	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion:         parseTLSVersion(config.TLSMinVersion),
		InsecureSkipVerify: !config.VerifySSL,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
	}
	
	// Create transport with security settings
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	
	// Create HTTP client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
	
	return &AuthenticatedHTTPClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// GetClient returns the underlying HTTP client
func (c *AuthenticatedHTTPClient) GetClient() *http.Client {
	return c.httpClient
}

// CreateRequest creates a new HTTP request with security headers
func (c *AuthenticatedHTTPClient) CreateRequest(method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeNetworkError, "Failed to create HTTP request", err)
	}
	
	// Add security headers
	req.Header.Set("User-Agent", "ESA/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	// Add security-specific headers
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")
	
	return req, nil
}

// DoRequest performs an HTTP request with logging and error handling
func (c *AuthenticatedHTTPClient) DoRequest(req *http.Request) (*http.Response, error) {
	start := time.Now()
	
	// Log request
	c.logger.Debug("Making HTTP request",
		utils.NewField("method", req.Method),
		utils.NewField("url", req.URL.String()),
		utils.NewField("host", req.Host),
	)
	
	// Perform request
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		c.logger.Error("HTTP request failed", err,
			utils.NewField("method", req.Method),
			utils.NewField("url", req.URL.String()),
			utils.NewField("duration_ms", duration.Milliseconds()),
		)
		return nil, utils.NewAppError(utils.ErrorCodeNetworkError, "HTTP request failed", err)
	}
	
	// Log response
	c.logger.Debug("HTTP request completed",
		utils.NewField("method", req.Method),
		utils.NewField("url", req.URL.String()),
		utils.NewField("status_code", resp.StatusCode),
		utils.NewField("duration_ms", duration.Milliseconds()),
	)
	
	return resp, nil
}

// parseTLSVersion parses TLS version string to tls constant
func parseTLSVersion(version string) uint16 {
	switch version {
	case "1.3":
		return tls.VersionTLS13
	case "1.2":
		return tls.VersionTLS12
	case "1.1":
		return tls.VersionTLS11
	case "1.0":
		return tls.VersionTLS10
	default:
		return tls.VersionTLS13 // Default to most secure
	}
}

// JiraAuthenticator handles Jira authentication
type JiraAuthenticator struct {
	httpClient *AuthenticatedHTTPClient
	creds      *CredentialStore
	logger     utils.Logger
}

// NewJiraAuthenticator creates a new Jira authenticator
func NewJiraAuthenticator(httpClient *AuthenticatedHTTPClient, creds *CredentialStore, logger utils.Logger) *JiraAuthenticator {
	return &JiraAuthenticator{
		httpClient: httpClient,
		creds:      creds,
		logger:     logger,
	}
}

// AddAuthHeaders adds Jira authentication headers to a request
func (j *JiraAuthenticator) AddAuthHeaders(req *http.Request, username string) error {
	// Get Jira credentials
	creds, err := j.creds.GetJiraCredentials()
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to get Jira credentials")
	}
	
	// Add Basic Auth header
	req.SetBasicAuth(username, creds.Token)
	
	j.logger.Debug("Added Jira authentication headers",
		utils.NewField("username", username),
		utils.NewField("url", req.URL.String()),
	)
	
	return nil
}

// ValidateCredentials validates Jira credentials by making a test request
func (j *JiraAuthenticator) ValidateCredentials(baseURL, username string) error {
	// Create test request to Jira API
	req, err := j.httpClient.CreateRequest("GET", baseURL+"/rest/api/2/myself", nil)
	if err != nil {
		return err
	}
	
	// Add authentication
	err = j.AddAuthHeaders(req, username)
	if err != nil {
		return err
	}
	
	// Make request
	resp, err := j.httpClient.DoRequest(req)
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to validate Jira credentials")
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode == http.StatusUnauthorized {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Invalid Jira credentials", nil)
	}
	
	if resp.StatusCode != http.StatusOK {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Unexpected response from Jira", nil).
			WithExtra("status_code", resp.StatusCode)
	}
	
	j.logger.Info("Jira credentials validated successfully",
		utils.NewField("username", username),
		utils.NewField("base_url", baseURL),
	)
	
	return nil
}

// GeminiAuthenticator handles Google Gemini authentication
type GeminiAuthenticator struct {
	httpClient *AuthenticatedHTTPClient
	creds      *CredentialStore
	logger     utils.Logger
}

// NewGeminiAuthenticator creates a new Gemini authenticator
func NewGeminiAuthenticator(httpClient *AuthenticatedHTTPClient, creds *CredentialStore, logger utils.Logger) *GeminiAuthenticator {
	return &GeminiAuthenticator{
		httpClient: httpClient,
		creds:      creds,
		logger:     logger,
	}
}

// AddAuthHeaders adds Gemini authentication headers to a request
func (g *GeminiAuthenticator) AddAuthHeaders(req *http.Request) error {
	// Get Gemini credentials
	creds, err := g.creds.GetGeminiCredentials()
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to get Gemini credentials")
	}
	
	// Add API key header - Gemini uses x-goog-api-key header
	req.Header.Set("x-goog-api-key", creds.APIKey)
	
	g.logger.Debug("Added Gemini authentication headers",
		utils.NewField("url", req.URL.String()),
	)
	
	return nil
}

// ValidateCredentials validates Gemini API credentials
func (g *GeminiAuthenticator) ValidateCredentials() error {
	// Create test request to Gemini API
	req, err := g.httpClient.CreateRequest("GET", "https://generativelanguage.googleapis.com/v1/models", nil)
	if err != nil {
		return err
	}
	
	// Add authentication
	err = g.AddAuthHeaders(req)
	if err != nil {
		return err
	}
	
	// Make request
	resp, err := g.httpClient.DoRequest(req)
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to validate Gemini credentials")
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode == http.StatusUnauthorized {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Invalid Gemini API key", nil)
	}
	
	if resp.StatusCode != http.StatusOK {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Unexpected response from Gemini", nil).
			WithExtra("status_code", resp.StatusCode)
	}
	
	g.logger.Info("Gemini credentials validated successfully")
	
	return nil
}

// GoogleAuthenticator handles Google API authentication
type GoogleAuthenticator struct {
	httpClient *AuthenticatedHTTPClient
	creds      *CredentialStore
	logger     utils.Logger
}

// NewGoogleAuthenticator creates a new Google authenticator
func NewGoogleAuthenticator(httpClient *AuthenticatedHTTPClient, creds *CredentialStore, logger utils.Logger) *GoogleAuthenticator {
	return &GoogleAuthenticator{
		httpClient: httpClient,
		creds:      creds,
		logger:     logger,
	}
}

// AddAuthHeaders adds Google authentication headers to a request
func (g *GoogleAuthenticator) AddAuthHeaders(req *http.Request) error {
	// Get Google credentials
	creds, err := g.creds.GetGoogleCredentials()
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to get Google credentials")
	}
	
	// Add access token header
	if creds.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+creds.AccessToken)
	}
	
	g.logger.Debug("Added Google authentication headers",
		utils.NewField("url", req.URL.String()),
	)
	
	return nil
}

// ValidateCredentials validates Google API credentials
func (g *GoogleAuthenticator) ValidateCredentials() error {
	// Get Google credentials
	creds, err := g.creds.GetGoogleCredentials()
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to get Google credentials")
	}
	
	if creds.AccessToken == "" {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Google access token is required", nil)
	}
	
	// Create test request to Google API
	req, err := g.httpClient.CreateRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		return err
	}
	
	// Add authentication
	err = g.AddAuthHeaders(req)
	if err != nil {
		return err
	}
	
	// Make request
	resp, err := g.httpClient.DoRequest(req)
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Failed to validate Google credentials")
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode == http.StatusUnauthorized {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Invalid Google access token", nil)
	}
	
	if resp.StatusCode != http.StatusOK {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Unexpected response from Google", nil).
			WithExtra("status_code", resp.StatusCode)
	}
	
	g.logger.Info("Google credentials validated successfully")
	
	return nil
}

// AuthManager manages all authentication providers
type AuthManager struct {
	httpClient        *AuthenticatedHTTPClient
	credentialStore   *CredentialStore
	jiraAuth          *JiraAuthenticator
	geminiAuth        *GeminiAuthenticator
	googleAuth        *GoogleAuthenticator
	logger            utils.Logger
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig, logger utils.Logger) *AuthManager {
	httpClient := NewAuthenticatedHTTPClient(config, logger)
	credentialStore := NewCredentialStore(logger)
	
	return &AuthManager{
		httpClient:      httpClient,
		credentialStore: credentialStore,
		jiraAuth:        NewJiraAuthenticator(httpClient, credentialStore, logger),
		geminiAuth:      NewGeminiAuthenticator(httpClient, credentialStore, logger),
		googleAuth:      NewGoogleAuthenticator(httpClient, credentialStore, logger),
		logger:          logger,
	}
}

// GetHTTPClient returns the authenticated HTTP client
func (m *AuthManager) GetHTTPClient() *AuthenticatedHTTPClient {
	return m.httpClient
}

// GetCredentialStore returns the credential store
func (m *AuthManager) GetCredentialStore() *CredentialStore {
	return m.credentialStore
}

// GetJiraAuthenticator returns the Jira authenticator
func (m *AuthManager) GetJiraAuthenticator() *JiraAuthenticator {
	return m.jiraAuth
}

// GetGeminiAuthenticator returns the Gemini authenticator
func (m *AuthManager) GetGeminiAuthenticator() *GeminiAuthenticator {
	return m.geminiAuth
}

// GetGoogleAuthenticator returns the Google authenticator
func (m *AuthManager) GetGoogleAuthenticator() *GoogleAuthenticator {
	return m.googleAuth
}

// ValidateAllCredentials validates all stored credentials
func (m *AuthManager) ValidateAllCredentials(jiraURL, jiraUsername string) error {
	// Validate that credentials exist
	err := m.credentialStore.ValidateAllCredentials()
	if err != nil {
		return err
	}
	
	// Validate Jira credentials
	err = m.jiraAuth.ValidateCredentials(jiraURL, jiraUsername)
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Jira credential validation failed")
	}
	
	// Validate Gemini credentials
	err = m.geminiAuth.ValidateCredentials()
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Gemini credential validation failed")
	}
	
	// Validate Google credentials
	err = m.googleAuth.ValidateCredentials()
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeAuthFailed, "Google credential validation failed")
	}
	
	m.logger.Info("All credentials validated successfully")
	
	return nil
}