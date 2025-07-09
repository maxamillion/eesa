package gdocs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/company/eesa/internal/config"
	"github.com/company/eesa/internal/security"
	"github.com/company/eesa/pkg/utils"
)

const (
	// Google Docs API endpoints
	BaseURL           = "https://docs.googleapis.com"
	CreateEndpoint    = "/v1/documents"
	BatchUpdateEndpoint = "/v1/documents/%s:batchUpdate"
	GetEndpoint       = "/v1/documents/%s"
	
	// Google Drive API endpoints for sharing
	DriveBaseURL      = "https://www.googleapis.com"
	ShareEndpoint     = "/drive/v3/files/%s/permissions"
)

// GoogleDocsClientInterface defines the interface for Google Docs client
type GoogleDocsClientInterface interface {
	CreateDocument(ctx context.Context, title string, content string) (*DocumentResponse, error)
	UpdateDocument(ctx context.Context, documentID string, requests []Request) (*BatchUpdateResponse, error)
	GetDocument(ctx context.Context, documentID string) (*DocumentResponse, error)
	ShareDocument(ctx context.Context, documentID string, emails []string, role string) error
	ValidateCredentials(ctx context.Context) error
	CreateExecutiveSummaryDocument(ctx context.Context, title, summary string, metadata map[string]interface{}) (*DocumentResponse, error)
}

// Client represents a Google Docs API client
type Client struct {
	baseURL     string
	driveBaseURL string
	httpClient  *security.AuthenticatedHTTPClient
	auth        *security.GoogleAuthenticator
	rateLimiter *utils.RateLimiter
	retryConfig *utils.RetryConfig
	logger      utils.Logger
}

// NewClient creates a new Google Docs client
func NewClient(cfg *config.Config, authManager *security.AuthManager, logger utils.Logger) *Client {
	// Create rate limiter (100 requests per 100 seconds for Google Docs API)
	rateLimiter := utils.NewRateLimiter(100, 100*time.Second, logger)
	
	// Create retry configuration
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors, utils.ErrorCodeGoogleError)
	
	return &Client{
		baseURL:     BaseURL,
		driveBaseURL: DriveBaseURL,
		httpClient:  authManager.GetHTTPClient(),
		auth:        authManager.GetGoogleAuthenticator(),
		rateLimiter: rateLimiter,
		retryConfig: retryConfig,
		logger:      logger,
	}
}

// CreateDocument creates a new Google Docs document
func (c *Client) CreateDocument(ctx context.Context, title string, content string) (*DocumentResponse, error) {
	if title == "" {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Document title is required", nil)
	}

	createRequest := CreateDocumentRequest{
		Title: title,
	}

	var response *DocumentResponse
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		// Marshal request
		reqBody, err := json.Marshal(createRequest)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to marshal create request", err)
		}
		
		// Create HTTP request
		req, err := c.createRequest(ctx, "POST", CreateEndpoint, reqBody)
		if err != nil {
			return err
		}
		
		// Make request
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to create document")
		}
		defer resp.Body.Close()
		
		// Handle error responses
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Document creation failed")
		}
		
		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to read response", err)
		}
		
		// Parse response
		response = &DocumentResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to parse response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}

	// If content is provided, add it to the document
	if content != "" {
		_, err = c.UpdateDocument(ctx, response.DocumentID, []Request{
			{
				InsertText: &InsertTextRequest{
					Text: content,
					Location: &Location{Index: 1},
				},
			},
		})
		if err != nil {
			c.logger.Warn("Failed to add content to document",
				utils.NewField("document_id", response.DocumentID),
				utils.NewField("error", err.Error()),
			)
		}
	}

	c.logger.Info("Created Google Docs document",
		utils.NewField("document_id", response.DocumentID),
		utils.NewField("title", response.Title),
	)

	return response, nil
}

// UpdateDocument updates a Google Docs document with batch requests
func (c *Client) UpdateDocument(ctx context.Context, documentID string, requests []Request) (*BatchUpdateResponse, error) {
	if documentID == "" {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Document ID is required", nil)
	}
	if len(requests) == 0 {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "At least one request is required", nil)
	}

	updateRequest := BatchUpdateDocumentRequest{
		Requests: requests,
	}

	var response *BatchUpdateResponse
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		// Marshal request
		reqBody, err := json.Marshal(updateRequest)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to marshal update request", err)
		}
		
		// Create HTTP request
		endpoint := fmt.Sprintf(BatchUpdateEndpoint, documentID)
		req, err := c.createRequest(ctx, "POST", endpoint, reqBody)
		if err != nil {
			return err
		}
		
		// Make request
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to update document")
		}
		defer resp.Body.Close()
		
		// Handle error responses
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Document update failed")
		}
		
		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to read response", err)
		}
		
		// Parse response
		response = &BatchUpdateResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to parse response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}

	c.logger.Info("Updated Google Docs document",
		utils.NewField("document_id", documentID),
		utils.NewField("requests_count", len(requests)),
	)

	return response, nil
}

// GetDocument retrieves a Google Docs document
func (c *Client) GetDocument(ctx context.Context, documentID string) (*DocumentResponse, error) {
	if documentID == "" {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Document ID is required", nil)
	}

	var response *DocumentResponse
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		// Create HTTP request
		endpoint := fmt.Sprintf(GetEndpoint, documentID)
		req, err := c.createRequest(ctx, "GET", endpoint, nil)
		if err != nil {
			return err
		}
		
		// Make request
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to get document")
		}
		defer resp.Body.Close()
		
		// Handle error responses
		if resp.StatusCode == http.StatusNotFound {
			return utils.NewAppError(utils.ErrorCodeAPINotFound, "Document not found", nil).
				WithExtra("document_id", documentID)
		}
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to get document")
		}
		
		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to read response", err)
		}
		
		// Parse response
		response = &DocumentResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to parse response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ShareDocument shares a Google Docs document with specified users
func (c *Client) ShareDocument(ctx context.Context, documentID string, emails []string, role string) error {
	if documentID == "" {
		return utils.NewAppError(utils.ErrorCodeDataInvalid, "Document ID is required", nil)
	}
	if len(emails) == 0 {
		return utils.NewAppError(utils.ErrorCodeDataInvalid, "At least one email is required", nil)
	}
	if role == "" {
		role = "reader" // Default to reader permission
	}

	// Share with each email
	for _, email := range emails {
		if email == "" {
			continue
		}

		permission := Permission{
			Type:         "user",
			Role:         role,
			EmailAddress: email,
		}

		err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
			// Marshal request
			reqBody, err := json.Marshal(permission)
			if err != nil {
				return utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to marshal permission request", err)
			}
			
			// Create HTTP request
			endpoint := fmt.Sprintf(ShareEndpoint, documentID)
			req, err := c.createDriveRequest(ctx, "POST", endpoint, reqBody)
			if err != nil {
				return err
			}
			
			// Make request
			resp, err := c.httpClient.DoRequest(req)
			if err != nil {
				return utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to share document")
			}
			defer resp.Body.Close()
			
			// Handle error responses
			if resp.StatusCode != http.StatusOK {
				return c.handleErrorResponse(resp, "Document sharing failed")
			}
			
			return nil
		}, c.logger)

		if err != nil {
			c.logger.Warn("Failed to share document with user",
				utils.NewField("document_id", documentID),
				utils.NewField("email", email),
				utils.NewField("error", err.Error()),
			)
			continue
		}

		c.logger.Info("Shared document with user",
			utils.NewField("document_id", documentID),
			utils.NewField("email", email),
			utils.NewField("role", role),
		)
	}

	return nil
}

// ValidateCredentials validates Google API credentials
func (c *Client) ValidateCredentials(ctx context.Context) error {
	// Try to create a simple test document and then delete it
	testDoc, err := c.CreateDocument(ctx, "API Test Document - Safe to Delete", "This is a test document to validate API credentials.")
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to validate Google Docs credentials")
	}

	c.logger.Info("Google Docs credentials validated successfully",
		utils.NewField("test_document_id", testDoc.DocumentID),
	)

	return nil
}

// CreateExecutiveSummaryDocument creates a properly formatted executive summary document
func (c *Client) CreateExecutiveSummaryDocument(ctx context.Context, title, summary string, metadata map[string]interface{}) (*DocumentResponse, error) {
	if title == "" {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Document title is required", nil)
	}
	if summary == "" {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Summary content is required", nil)
	}

	// Create the document first
	doc, err := c.CreateDocument(ctx, title, "")
	if err != nil {
		return nil, err
	}

	// Build formatting requests for executive summary
	requests := []Request{
		// Insert title
		{
			InsertText: &InsertTextRequest{
				Text:     title + "\n\n",
				Location: &Location{Index: 1},
			},
		},
		// Format title as heading
		{
			UpdateTextStyle: &UpdateTextStyleRequest{
				Range: &Range{StartIndex: 1, EndIndex: int32(len(title) + 1)},
				TextStyle: &TextStyle{
					Bold:         boolPtr(true),
					FontSize:     &Dimension{Magnitude: 18, Unit: "PT"},
				},
				Fields: "bold,fontSize",
			},
		},
		// Insert metadata if provided
	}

	// Add metadata section if provided
	if len(metadata) > 0 {
		metadataText := c.formatMetadata(metadata)
		currentIndex := int32(len(title) + 3) // After title and newlines
		
		requests = append(requests, Request{
			InsertText: &InsertTextRequest{
				Text:     metadataText + "\n\n",
				Location: &Location{Index: currentIndex},
			},
		})
		currentIndex += int32(len(metadataText) + 2)
	}

	// Insert summary content
	summaryIndex := int32(len(title) + 3)
	if len(metadata) > 0 {
		metadataText := c.formatMetadata(metadata)
		summaryIndex += int32(len(metadataText) + 2)
	}

	requests = append(requests, Request{
		InsertText: &InsertTextRequest{
			Text:     summary,
			Location: &Location{Index: summaryIndex},
		},
	})

	// Apply formatting
	_, err = c.UpdateDocument(ctx, doc.DocumentID, requests)
	if err != nil {
		c.logger.Warn("Failed to format executive summary document",
			utils.NewField("document_id", doc.DocumentID),
			utils.NewField("error", err.Error()),
		)
	}

	c.logger.Info("Created executive summary document",
		utils.NewField("document_id", doc.DocumentID),
		utils.NewField("title", title),
		utils.NewField("summary_length", len(summary)),
	)

	return doc, nil
}

// createRequest creates an authenticated HTTP request for Google Docs API
func (c *Client) createRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Request, error) {
	fullURL := c.baseURL + endpoint
	
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to create request", err)
	}
	
	// Add authentication headers
	err = c.auth.AddAuthHeaders(req)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to add auth headers")
	}
	
	// Set content type for POST requests
	if method == "POST" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return req, nil
}

// createDriveRequest creates an authenticated HTTP request for Google Drive API
func (c *Client) createDriveRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Request, error) {
	fullURL := c.driveBaseURL + endpoint
	
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Failed to create drive request", err)
	}
	
	// Add authentication headers
	err = c.auth.AddAuthHeaders(req)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorCodeGoogleError, "Failed to add auth headers")
	}
	
	// Set content type for POST requests
	if method == "POST" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return req, nil
}

// handleErrorResponse handles HTTP error responses
func (c *Client) handleErrorResponse(resp *http.Response, message string) error {
	body, _ := io.ReadAll(resp.Body)
	
	var errorCode utils.ErrorCode
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		errorCode = utils.ErrorCodeAPIUnauthorized
	case http.StatusForbidden:
		errorCode = utils.ErrorCodeAPIUnauthorized
	case http.StatusNotFound:
		errorCode = utils.ErrorCodeAPINotFound
	case http.StatusTooManyRequests:
		errorCode = utils.ErrorCodeAPIRateLimit
	case http.StatusBadRequest:
		errorCode = utils.ErrorCodeAPIBadRequest
	default:
		if resp.StatusCode >= 500 {
			errorCode = utils.ErrorCodeAPIServerError
		} else {
			errorCode = utils.ErrorCodeGoogleError
		}
	}
	
	// Try to parse error response
	var googleError GoogleErrorResponse
	if err := json.Unmarshal(body, &googleError); err == nil && googleError.ErrorInfo.Message != "" {
		message = googleError.ErrorInfo.Message
	}
	
	return utils.NewAppError(errorCode, message, nil).
		WithService("google_docs").
		WithExtra("status_code", resp.StatusCode).
		WithExtra("response_body", string(body))
}

// formatMetadata formats metadata as a readable string
func (c *Client) formatMetadata(metadata map[string]interface{}) string {
	var parts []string
	
	if generatedAt, ok := metadata["generated_at"].(time.Time); ok {
		parts = append(parts, fmt.Sprintf("Generated: %s", generatedAt.Format("January 2, 2006 at 3:04 PM")))
	}
	
	if model, ok := metadata["model"].(string); ok {
		parts = append(parts, fmt.Sprintf("AI Model: %s", model))
	}
	
	if tokensUsed, ok := metadata["tokens_used"].(int); ok {
		parts = append(parts, fmt.Sprintf("Tokens Used: %d", tokensUsed))
	}
	
	if activityCount, ok := metadata["activity_count"].(int); ok {
		parts = append(parts, fmt.Sprintf("Activities Analyzed: %d", activityCount))
	}
	
	if timeRange, ok := metadata["time_range"].(string); ok {
		parts = append(parts, fmt.Sprintf("Time Period: %s", timeRange))
	}
	
	if len(parts) == 0 {
		return ""
	}
	
	return strings.Join(parts, " | ")
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}