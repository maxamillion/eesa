package gemini

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
	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
)

const (
	// Gemini API endpoints
	BaseURL         = "https://generativelanguage.googleapis.com"
	GenerateEndpoint = "/v1/models/gemini-pro:generateContent"
	ModelsEndpoint   = "/v1/models"
)

// GeminiClientInterface defines the interface for Gemini AI client
type GeminiClientInterface interface {
	GenerateSummary(ctx context.Context, activities []models.Activity, prompt string) (*SummaryResponse, error)
	ValidateAPIKey(ctx context.Context) error
	ListModels(ctx context.Context) (*ModelsResponse, error)
	GenerateContent(ctx context.Context, request *GenerateRequest) (*GenerateResponse, error)
}

// Client represents a Gemini AI client
type Client struct {
	baseURL     string
	model       string
	temperature float32
	maxTokens   int
	httpClient  *security.AuthenticatedHTTPClient
	auth        *security.GeminiAuthenticator
	rateLimiter *utils.RateLimiter
	retryConfig *utils.RetryConfig
	logger      utils.Logger
}

// NewClient creates a new Gemini AI client
func NewClient(cfg *config.Config, authManager *security.AuthManager, logger utils.Logger) *Client {
	// Create rate limiter (60 requests per minute for Gemini API)
	rateLimiter := utils.NewRateLimiter(60, time.Minute, logger)
	
	// Create retry configuration
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors, utils.ErrorCodeGeminiError)
	
	return &Client{
		baseURL:     BaseURL,
		model:       cfg.Gemini.Model,
		temperature: cfg.Gemini.Temperature,
		maxTokens:   cfg.Gemini.MaxTokens,
		httpClient:  authManager.GetHTTPClient(),
		auth:        authManager.GetGeminiAuthenticator(),
		rateLimiter: rateLimiter,
		retryConfig: retryConfig,
		logger:      logger,
	}
}

// GenerateSummary generates an executive summary from activities
func (c *Client) GenerateSummary(ctx context.Context, activities []models.Activity, customPrompt string) (*SummaryResponse, error) {
	if len(activities) == 0 {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "No activities provided for summary generation", nil)
	}
	
	// Build the prompt with activities data
	prompt := c.buildSummaryPrompt(activities, customPrompt)
	
	// Create generate request
	request := &GenerateRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:  &c.temperature,
			MaxTokens:    &c.maxTokens,
			TopP:         float32Ptr(0.8),
			TopK:         int32Ptr(40),
			StopSequences: []string{"---END---"},
		},
		SafetySettings: []SafetySetting{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
		},
	}
	
	// Generate content
	response, err := c.GenerateContent(ctx, request)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorCodeGeminiError, "Failed to generate summary")
	}
	
	// Extract summary from response
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, utils.NewAppError(utils.ErrorCodeGeminiError, "No content generated", nil)
	}
	
	summaryText := response.Candidates[0].Content.Parts[0].Text
	
	// Create summary response
	summaryResponse := &SummaryResponse{
		Summary:     summaryText,
		TokensUsed:  response.UsageMetadata.TotalTokenCount,
		Model:       c.model,
		Temperature: c.temperature,
		GeneratedAt: time.Now(),
		Activities:  activities,
	}
	
	c.logger.Info("Generated executive summary",
		utils.NewField("activities_count", len(activities)),
		utils.NewField("summary_length", len(summaryText)),
		utils.NewField("tokens_used", response.UsageMetadata.TotalTokenCount),
		utils.NewField("model", c.model),
	)
	
	return summaryResponse, nil
}

// GenerateContent generates content using Gemini API
func (c *Client) GenerateContent(ctx context.Context, request *GenerateRequest) (*GenerateResponse, error) {
	var response *GenerateResponse
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		// Marshal request
		reqBody, err := json.Marshal(request)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGeminiError, "Failed to marshal request", err)
		}
		
		// Create HTTP request
		req, err := c.createRequest(ctx, "POST", GenerateEndpoint, reqBody)
		if err != nil {
			return err
		}
		
		// Make request
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeGeminiError, "Failed to make request")
		}
		defer resp.Body.Close()
		
		// Handle error responses
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Content generation failed")
		}
		
		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGeminiError, "Failed to read response", err)
		}
		
		// Parse response
		response = &GenerateResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return utils.NewAppError(utils.ErrorCodeGeminiError, "Failed to parse response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// ValidateAPIKey validates the Gemini API key
func (c *Client) ValidateAPIKey(ctx context.Context) error {
	_, err := c.ListModels(ctx)
	return err
}

// ListModels lists available models
func (c *Client) ListModels(ctx context.Context) (*ModelsResponse, error) {
	var response *ModelsResponse
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		req, err := c.createRequest(ctx, "GET", ModelsEndpoint, nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeGeminiError, "Failed to list models")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to list models")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeGeminiError, "Failed to read response", err)
		}
		
		response = &ModelsResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return utils.NewAppError(utils.ErrorCodeGeminiError, "Failed to parse response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// createRequest creates an authenticated HTTP request
func (c *Client) createRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Request, error) {
	fullURL := c.baseURL + endpoint
	
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeGeminiError, "Failed to create request", err)
	}
	
	// Add authentication headers
	err = c.auth.AddAuthHeaders(req)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorCodeGeminiError, "Failed to add auth headers")
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
			errorCode = utils.ErrorCodeGeminiError
		}
	}
	
	// Try to parse error response
	var geminiError ErrorResponse
	if err := json.Unmarshal(body, &geminiError); err == nil && geminiError.ErrorInfo.Message != "" {
		message = geminiError.ErrorInfo.Message
	}
	
	return utils.NewAppError(errorCode, message, nil).
		WithService("gemini").
		WithExtra("status_code", resp.StatusCode).
		WithExtra("response_body", string(body))
}

// buildSummaryPrompt builds the prompt for executive summary generation
func (c *Client) buildSummaryPrompt(activities []models.Activity, customPrompt string) string {
	var prompt strings.Builder
	
	// Add system prompt
	prompt.WriteString(`You are an executive assistant creating a comprehensive executive summary for a technology organization. 
Your task is to analyze the provided Jira activity data and create a professional, concise summary suitable for executive leadership.

INSTRUCTIONS:
1. Create a structured executive summary with the following sections:
   - Executive Overview (2-3 sentences)
   - Key Accomplishments
   - Progress by Project/Team
   - Metrics and Performance
   - Issues and Risks
   - Next Steps/Recommendations

2. Focus on business impact and strategic insights, not technical details
3. Use clear, professional language appropriate for C-level executives
4. Highlight trends, patterns, and key metrics
5. Include specific numbers and timeframes where relevant
6. Keep the summary concise but comprehensive (500-1000 words)

`)
	
	// Add custom prompt if provided
	if customPrompt != "" {
		prompt.WriteString("ADDITIONAL INSTRUCTIONS:\n")
		prompt.WriteString(customPrompt)
		prompt.WriteString("\n\n")
	}
	
	// Add activities data
	prompt.WriteString("JIRA ACTIVITY DATA:\n")
	prompt.WriteString("===================\n\n")
	
	// Group activities by project for better organization
	projectGroups := make(map[string][]models.Activity)
	for _, activity := range activities {
		projectKey := activity.Project.Key
		if projectKey == "" {
			projectKey = "UNKNOWN"
		}
		projectGroups[projectKey] = append(projectGroups[projectKey], activity)
	}
	
	// Add project summaries
	for projectKey, projectActivities := range projectGroups {
		prompt.WriteString(fmt.Sprintf("PROJECT: %s\n", projectKey))
		prompt.WriteString("=================\n")
		
		for _, activity := range projectActivities {
			prompt.WriteString(fmt.Sprintf("- %s [%s]: %s\n", 
				activity.Key, activity.Status, activity.Summary))
			prompt.WriteString(fmt.Sprintf("  Priority: %s | Type: %s | Assignee: %s\n", 
				activity.Priority, activity.Type, activity.Assignee.DisplayName))
			prompt.WriteString(fmt.Sprintf("  Created: %s | Updated: %s\n", 
				activity.Created.Format("2006-01-02"), activity.Updated.Format("2006-01-02")))
			
			if activity.TimeSpent > 0 {
				prompt.WriteString(fmt.Sprintf("  Time Spent: %s\n", activity.GetFormattedTimeSpent()))
			}
			
			if len(activity.Comments) > 0 {
				prompt.WriteString(fmt.Sprintf("  Comments: %d\n", len(activity.Comments)))
			}
			
			prompt.WriteString("\n")
		}
		prompt.WriteString("\n")
	}
	
	// Add summary statistics
	prompt.WriteString("SUMMARY STATISTICS:\n")
	prompt.WriteString("==================\n")
	
	totalIssues := len(activities)
	completedIssues := 0
	inProgressIssues := 0
	totalTimeSpent := int64(0)
	
	for _, activity := range activities {
		if activity.IsCompleted() {
			completedIssues++
		} else if activity.IsInProgress() {
			inProgressIssues++
		}
		totalTimeSpent += activity.TimeSpent
	}
	
	prompt.WriteString(fmt.Sprintf("Total Issues: %d\n", totalIssues))
	prompt.WriteString(fmt.Sprintf("Completed Issues: %d\n", completedIssues))
	prompt.WriteString(fmt.Sprintf("In Progress Issues: %d\n", inProgressIssues))
	prompt.WriteString(fmt.Sprintf("Total Time Spent: %s\n", models.FormatTimeSpent(totalTimeSpent)))
	
	if totalIssues > 0 {
		completionRate := float64(completedIssues) / float64(totalIssues) * 100
		prompt.WriteString(fmt.Sprintf("Completion Rate: %.1f%%\n", completionRate))
	}
	
	prompt.WriteString("\n")
	prompt.WriteString("Please generate a comprehensive executive summary based on this data.")
	
	return prompt.String()
}

// Helper functions for pointer types
func float32Ptr(v float32) *float32 {
	return &v
}

func int32Ptr(v int32) *int32 {
	return &v
}