package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
)

// SearchRequest represents a Jira search request
type SearchRequest struct {
	JQL        string   `json:"jql"`
	StartAt    int      `json:"startAt"`
	MaxResults int      `json:"maxResults"`
	Fields     []string `json:"fields"`
	Expand     []string `json:"expand"`
}

// SearchResult represents a Jira search result
type SearchResult struct {
	StartAt    int             `json:"startAt"`
	MaxResults int             `json:"maxResults"`
	Total      int             `json:"total"`
	Issues     []IssueResponse `json:"issues"`
}

// IssueResponse represents a Jira issue response
type IssueResponse struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Fields IssueFields `json:"fields"`
}

// IssueFields represents the fields of a Jira issue
type IssueFields struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	IssueType   IssueType   `json:"issuetype"`
	Status      Status      `json:"status"`
	Priority    Priority    `json:"priority"`
	Reporter    UserField   `json:"reporter"`
	Assignee    UserField   `json:"assignee"`
	Created     string      `json:"created"`
	Updated     string      `json:"updated"`
	Project     ProjectField `json:"project"`
	TimeTracking TimeTracking `json:"timetracking"`
}

// IssueType represents an issue type
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Status represents an issue status
type Status struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Priority represents an issue priority
type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserField represents a user field in Jira
type UserField struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
}

// ProjectField represents a project field in Jira
type ProjectField struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Lead        UserField `json:"lead"`
}

// TimeTracking represents time tracking information
type TimeTracking struct {
	TimeSpent         string `json:"timeSpent"`
	TimeSpentSeconds  int64  `json:"timeSpentSeconds"`
	RemainingEstimate string `json:"remainingEstimate"`
	RemainingEstimateSeconds int64 `json:"remainingEstimateSeconds"`
}

// WorklogResponse represents a worklog response
type WorklogResponse struct {
	StartAt    int            `json:"startAt"`
	MaxResults int            `json:"maxResults"`
	Total      int            `json:"total"`
	Worklogs   []WorklogEntry `json:"worklogs"`
}

// WorklogEntry represents a worklog entry
type WorklogEntry struct {
	ID               string    `json:"id"`
	Author           UserField `json:"author"`
	TimeSpent        string    `json:"timeSpent"`
	TimeSpentSeconds int64     `json:"timeSpentSeconds"`
	Comment          string    `json:"comment"`
	Started          string    `json:"started"`
	Created          string    `json:"created"`
	Updated          string    `json:"updated"`
}

// CommentsResponse represents a comments response
type CommentsResponse struct {
	StartAt    int            `json:"startAt"`
	MaxResults int            `json:"maxResults"`
	Total      int            `json:"total"`
	Comments   []CommentEntry `json:"comments"`
}

// CommentEntry represents a comment entry
type CommentEntry struct {
	ID       string    `json:"id"`
	Author   UserField `json:"author"`
	Body     string    `json:"body"`
	Created  string    `json:"created"`
	Updated  string    `json:"updated"`
	Visibility CommentVisibility `json:"visibility"`
}

// CommentVisibility represents comment visibility
type CommentVisibility struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// convertIssueToActivity converts a Jira issue to an activity
func (c *Client) convertIssueToActivity(issue *IssueResponse) (*models.Activity, error) {
	activity := &models.Activity{
		ID:          issue.ID,
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		Type:        issue.Fields.IssueType.Name,
		Status:      issue.Fields.Status.Name,
		Priority:    issue.Fields.Priority.Name,
		TimeSpent:   issue.Fields.TimeTracking.TimeSpentSeconds,
	}
	
	// Convert dates
	var err error
	activity.Created, err = parseJiraTimestamp(issue.Fields.Created)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse created date", err)
	}
	
	activity.Updated, err = parseJiraTimestamp(issue.Fields.Updated)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse updated date", err)
	}
	
	// Convert users
	activity.Reporter = convertUserField(issue.Fields.Reporter)
	activity.Assignee = convertUserField(issue.Fields.Assignee)
	
	// Convert project
	activity.Project = convertProjectField(issue.Fields.Project)
	
	return activity, nil
}

// convertWorklogEntry converts a Jira worklog entry to a worklog model
func (c *Client) convertWorklogEntry(entry *WorklogEntry) (*models.Worklog, error) {
	worklog := &models.Worklog{
		ID:          entry.ID,
		Author:      convertUserField(entry.Author),
		TimeSpent:   entry.TimeSpentSeconds,
		Description: entry.Comment,
	}
	
	// Convert dates
	var err error
	worklog.Started, err = parseJiraTimestamp(entry.Started)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse worklog started date", err)
	}
	
	worklog.Created, err = parseJiraTimestamp(entry.Created)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse worklog created date", err)
	}
	
	worklog.Updated, err = parseJiraTimestamp(entry.Updated)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse worklog updated date", err)
	}
	
	return worklog, nil
}

// convertComment converts a Jira comment to a comment model
func (c *Client) convertComment(comment *CommentEntry) (*models.Comment, error) {
	commentModel := &models.Comment{
		ID:     comment.ID,
		Author: convertUserField(comment.Author),
		Body:   comment.Body,
	}
	
	// Convert visibility
	if comment.Visibility.Type != "" {
		commentModel.Visibility = comment.Visibility.Type + ":" + comment.Visibility.Value
	}
	
	// Convert dates
	var err error
	commentModel.Created, err = parseJiraTimestamp(comment.Created)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse comment created date", err)
	}
	
	commentModel.Updated, err = parseJiraTimestamp(comment.Updated)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeDataInvalid, "Failed to parse comment updated date", err)
	}
	
	return commentModel, nil
}

// convertUserField converts a Jira user field to a user model
func convertUserField(userField UserField) models.User {
	return models.User{
		AccountID:    userField.AccountID,
		DisplayName:  userField.DisplayName,
		EmailAddress: userField.EmailAddress,
		Active:       userField.Active,
	}
}

// convertProjectField converts a Jira project field to a project model
func convertProjectField(projectField ProjectField) models.Project {
	return models.Project{
		ID:          projectField.ID,
		Key:         projectField.Key,
		Name:        projectField.Name,
		Description: projectField.Description,
		Lead:        convertUserField(projectField.Lead),
	}
}

// parseJiraTimestamp parses a Jira timestamp string to time.Time
func parseJiraTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, nil
	}
	
	// Jira uses ISO 8601 format with timezone
	layouts := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, timestamp); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, utils.NewAppError(utils.ErrorCodeDataInvalid, "Unsupported timestamp format", nil).
		WithExtra("timestamp", timestamp)
}

// parseTimeSpent parses time spent string to seconds
func parseTimeSpent(timeSpent string) int64 {
	if timeSpent == "" {
		return 0
	}
	
	// Handle different formats: "1h 30m", "30m", "1h", "90m"
	var totalSeconds int64
	
	// Simple parsing for common formats
	if strings.Contains(timeSpent, "h") {
		parts := strings.Split(timeSpent, "h")
		if len(parts) >= 1 {
			if hours, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err == nil {
				totalSeconds += hours * 3600
			}
		}
		if len(parts) >= 2 {
			minutePart := strings.TrimSpace(parts[1])
			if strings.HasSuffix(minutePart, "m") {
				minutePart = strings.TrimSuffix(minutePart, "m")
				if minutes, err := strconv.ParseInt(strings.TrimSpace(minutePart), 10, 64); err == nil {
					totalSeconds += minutes * 60
				}
			}
		}
	} else if strings.Contains(timeSpent, "m") {
		minutePart := strings.TrimSuffix(timeSpent, "m")
		if minutes, err := strconv.ParseInt(strings.TrimSpace(minutePart), 10, 64); err == nil {
			totalSeconds = minutes * 60
		}
	}
	
	return totalSeconds
}

// ErrorResponse represents a Jira error response
type ErrorResponse struct {
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

// Error returns the error message
func (e *ErrorResponse) Error() string {
	if len(e.ErrorMessages) > 0 {
		return strings.Join(e.ErrorMessages, "; ")
	}
	
	if len(e.Errors) > 0 {
		var messages []string
		for field, message := range e.Errors {
			messages = append(messages, fmt.Sprintf("%s: %s", field, message))
		}
		return strings.Join(messages, "; ")
	}
	
	return "Unknown Jira error"
}

// ValidationResult represents the result of validating Jira connection
type ValidationResult struct {
	Valid    bool   `json:"valid"`
	Message  string `json:"message"`
	UserInfo *UserInfo `json:"user_info,omitempty"`
}

// UserInfo represents current user information
type UserInfo struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
	Permissions  []string `json:"permissions"`
}

// ServerInfo represents Jira server information
type ServerInfo struct {
	BaseURL        string `json:"baseUrl"`
	Version        string `json:"version"`
	BuildNumber    int    `json:"buildNumber"`
	BuildDate      string `json:"buildDate"`
	ServerTitle    string `json:"serverTitle"`
	DeploymentType string `json:"deploymentType"`
}

// GetServerInfo gets server information
func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	var serverInfo *ServerInfo
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		req, err := c.createRequest(ctx, "GET", "/rest/api/2/serverInfo", nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to get server info")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to get server info")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to read server info response", err)
		}
		
		serverInfo = &ServerInfo{}
		if err := json.Unmarshal(body, serverInfo); err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to parse server info response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return serverInfo, nil
}

// GetCurrentUser gets current user information
func (c *Client) GetCurrentUser(ctx context.Context) (*UserInfo, error) {
	var userInfo *UserInfo
	
	err := utils.RetryWithRateLimit(ctx, c.retryConfig, c.rateLimiter, func() error {
		req, err := c.createRequest(ctx, "GET", "/rest/api/2/myself", nil)
		if err != nil {
			return err
		}
		
		resp, err := c.httpClient.DoRequest(req)
		if err != nil {
			return utils.WrapError(err, utils.ErrorCodeJiraError, "Failed to get current user")
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp, "Failed to get current user")
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to read current user response", err)
		}
		
		userInfo = &UserInfo{}
		if err := json.Unmarshal(body, userInfo); err != nil {
			return utils.NewAppError(utils.ErrorCodeJiraError, "Failed to parse current user response", err)
		}
		
		return nil
	}, c.logger)
	
	if err != nil {
		return nil, err
	}
	
	return userInfo, nil
}