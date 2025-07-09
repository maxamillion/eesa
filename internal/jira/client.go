package jira

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

// JiraClientInterface defines the interface for Jira client
type JiraClientInterface interface {
	GetUserActivities(ctx context.Context, users []string, timeRange config.TimeRange) ([]models.Activity, error)
	ValidateConnection(ctx context.Context) error
	SearchIssues(ctx context.Context, jql string, fields []string, startAt, maxResults int) (*SearchResult, error)
	GetIssue(ctx context.Context, issueKey string, fields []string) (*models.Activity, error)
	GetWorklog(ctx context.Context, issueKey string) ([]models.Worklog, error)
	GetComments(ctx context.Context, issueKey string) ([]models.Comment, error)
}

// Client represents a Jira API client
type Client struct {
	baseURL      string
	username     string
	httpClient   *security.AuthenticatedHTTPClient
	auth         *security.JiraAuthenticator
	rateLimiter  *utils.RateLimiter
	retryConfig  *utils.RetryConfig
	logger       utils.Logger
}

// NewClient creates a new Jira client
func NewClient(cfg *config.Config, authManager *security.AuthManager, logger utils.Logger) *Client {
	// Create rate limiter (600 requests per minute for Jira Cloud)
	rateLimiter := utils.NewRateLimiter(600, time.Minute, logger)
	
	// Create retry configuration
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors, utils.ErrorCodeJiraError)
	
	return &Client{
		baseURL:     strings.TrimSuffix(cfg.Jira.URL, "/"),
		username:    cfg.Jira.Username,
		httpClient:  authManager.GetHTTPClient(),
		auth:        authManager.GetJiraAuthenticator(),
		rateLimiter: rateLimiter,
		retryConfig: retryConfig,
		logger:      logger,
	}
}

// ValidateConnection validates the connection to Jira
func (c *Client) ValidateConnection(ctx context.Context) error {
	return utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		req, err := c.createRequest(ctx, "GET", "/rest/api/2/myself", nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to validate Jira connection")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Connection validation failed")
		}
		
		c.logger.Info("Jira connection validated successfully")
		return nil
	}, c.logger)
}

// GetUserActivities retrieves activities for specified users within a time range
func (c *Client) GetUserActivities(ctx context.Context, users []string, timeRange config.TimeRange) ([]models.Activity, error) {
	var allActivities []models.Activity
	
	// Build JQL query
	jql := c.buildUserActivitiesJQL(users, timeRange)
	c.logger.Debug("Built JQL query", utils.NewField("jql", jql))
	
	// Search for issues with pagination
	startAt := 0
	maxResults := 100
	
	for {
		searchResult, err := c.SearchIssues(ctx, jql, c.getDefaultFields(), startAt, maxResults)
		if err != nil {
			return nil, utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to search for user activities")
		}
		
		// Convert search results to activities
		activities, err := c.convertSearchResultToActivities(ctx, searchResult)
		if err != nil {
			return nil, utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to convert search results")
		}
		
		allActivities = append(allActivities, activities...)
		
		// Check if we've retrieved all results
		if startAt+len(searchResult.Issues) >= searchResult.Total {
			break
		}
		
		startAt += maxResults
	}
	
	c.logger.Info("Retrieved user activities",
		utils.NewField("total_activities", len(allActivities)),
		utils.NewField("users", users),
		utils.NewField("time_range", fmt.Sprintf("%v to %v", timeRange.Start, timeRange.End)),
	)
	
	return allActivities, nil
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, fields []string, startAt, maxResults int) (*SearchResult, error) {
	var result *SearchResult
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		searchRequest := SearchRequest{
			JQL:        jql,
			StartAt:    startAt,
			MaxResults: maxResults,
			Fields:     fields,
			Expand:     []string{"changelog"},
		}
		
		reqBody, err := json.Marshal(searchRequest)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to marshal search request", err)
		}
		
		req, err := c.createRequest(ctx, "POST", "/rest/api/2/search", reqBody)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to search issues")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Issue search failed")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to read search response", err)
		}
		
		result = &SearchResult{}
		if err := json.Unmarshal(body, result); err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to parse search response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetIssue retrieves a single issue by key
func (c *Client) GetIssue(ctx context.Context, issueKey string, fields []string) (*models.Activity, error) {
	var activity *models.Activity
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		fieldsParam := strings.Join(fields, ",")
		endpoint := fmt.Sprintf("/rest/api/2/issue/%s?fields=%s&expand=changelog", issueKey, fieldsParam)
		
		req, err := c.createRequest(ctx, "GET", endpoint, nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to get issue")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusNotFound {
			return utils.NewAppError(utils.ErrorCodeAPINotFound, "Issue not found", nil).
				WithExtra("issue_key", issueKey)
		}
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to get issue")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to read issue response", err)
		}
		
		var issue IssueResponse
		if err := json.Unmarshal(body, &issue); err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to parse issue response", err)
		}
		
		// Convert to activity
		activity, err = c.convertIssueToActivity(&issue)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to convert issue to activity")
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return activity, nil
}

// GetWorklog retrieves worklog entries for an issue
func (c *Client) GetWorklog(ctx context.Context, issueKey string) ([]models.Worklog, error) {
	var worklog []models.Worklog
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issueKey)
		
		req, err := c.createRequest(ctx, "GET", endpoint, nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to get worklog")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to get worklog")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to read worklog response", err)
		}
		
		var worklogResponse WorklogResponse
		if err := json.Unmarshal(body, &worklogResponse); err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to parse worklog response", err)
		}
		
		// Convert to worklog entries
		for _, entry := range worklogResponse.Worklogs {
			worklogEntry, err := c.convertWorklogEntry(&entry)
			if err != nil {
				c.logger.Warn("Failed to convert worklog entry", utils.NewField("entry_id", entry.ID))
				continue
			}
			worklog = append(worklog, *worklogEntry)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return worklog, nil
}

// GetComments retrieves comments for an issue
func (c *Client) GetComments(ctx context.Context, issueKey string) ([]models.Comment, error) {
	var comments []models.Comment
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		endpoint := fmt.Sprintf("/rest/api/2/issue/%s/comment", issueKey)
		
		req, err := c.createRequest(ctx, "GET", endpoint, nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to get comments")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to get comments")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to read comments response", err)
		}
		
		var commentsResponse CommentsResponse
		if err := json.Unmarshal(body, &commentsResponse); err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to parse comments response", err)
		}
		
		// Convert to comment entries
		for _, comment := range commentsResponse.Comments {
			commentEntry, err := c.convertComment(&comment)
			if err != nil {
				c.logger.Warn("Failed to convert comment", utils.NewField("comment_id", comment.ID))
				continue
			}
			comments = append(comments, *commentEntry)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return comments, nil
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
		return nil, utils.NewAppError(utils.ErrorCodeJiraError, "Failed to create request", err)
	}
	
	// Add authentication headers
	err = c.auth.AddAuthHeaders(req, c.username)
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to add auth headers")
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
			errorCode = utils.ErrorCodeJiraError
		}
	}
	
	return utils.NewAppError(errorCode, message, nil).
		WithService("jira").
		WithExtra("status_code", resp.StatusCode).
		WithExtra("response_body", string(body))
}

// buildUserActivitiesJQL builds a JQL query for user activities
func (c *Client) buildUserActivitiesJQL(users []string, timeRange config.TimeRange) string {
	var conditions []string
	
	// Add user conditions
	if len(users) > 0 {
		userConditions := make([]string, len(users))
		for i, user := range users {
			userConditions[i] = fmt.Sprintf("assignee = '%s' OR reporter = '%s'", user, user)
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(userConditions, " OR ")))
	}
	
	// Add time range conditions
	startDate := timeRange.Start.Format("2006-01-02")
	endDate := timeRange.End.Format("2006-01-02")
	conditions = append(conditions, fmt.Sprintf("updated >= '%s' AND updated <= '%s'", startDate, endDate))
	
	// Combine conditions
	jql := strings.Join(conditions, " AND ")
	
	// Add ordering
	jql += " ORDER BY updated DESC"
	
	return jql
}

// getDefaultFields returns the default fields to retrieve
func (c *Client) getDefaultFields() []string {
	return []string{
		"id",
		"key",
		"summary",
		"description",
		"issuetype",
		"status",
		"priority",
		"reporter",
		"assignee",
		"created",
		"updated",
		"project",
		"timetracking",
		"worklog",
		"comment",
	}
}

// convertSearchResultToActivities converts search results to activities
func (c *Client) convertSearchResultToActivities(ctx context.Context, searchResult *SearchResult) ([]models.Activity, error) {
	activities := make([]models.Activity, 0, len(searchResult.Issues))
	
	for _, issue := range searchResult.Issues {
		activity, err := c.convertIssueToActivity(&issue)
		if err != nil {
			c.logger.Warn("Failed to convert issue to activity",
				utils.NewField("issue_key", issue.Key),
				utils.NewField("error", err.Error()),
			)
			continue
		}
		
		// Get additional data (worklog and comments)
		worklog, err := c.GetWorklog(ctx, issue.Key)
		if err != nil {
			c.logger.Warn("Failed to get worklog for issue",
				utils.NewField("issue_key", issue.Key),
				utils.NewField("error", err.Error()),
			)
		} else {
			activity.Worklog = worklog
			// Calculate total time spent
			for _, entry := range worklog {
				activity.TimeSpent += entry.TimeSpent
			}
		}
		
		comments, err := c.GetComments(ctx, issue.Key)
		if err != nil {
			c.logger.Warn("Failed to get comments for issue",
				utils.NewField("issue_key", issue.Key),
				utils.NewField("error", err.Error()),
			)
		} else {
			activity.Comments = comments
		}
		
		activities = append(activities, *activity)
	}
	
	return activities, nil
}