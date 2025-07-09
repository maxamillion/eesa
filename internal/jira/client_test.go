package jira

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
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	
	client := NewClient(cfg, authManager, logger)
	
	assert.NotNil(t, client)
	assert.Equal(t, "https://test.atlassian.net", client.baseURL)
	assert.Equal(t, "testuser", client.username)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.auth)
	assert.NotNil(t, client.rateLimiter)
	assert.NotNil(t, client.retryConfig)
	assert.Equal(t, logger, client.logger)
}

func TestClient_buildUserActivitiesJQL(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)
	
	users := []string{"user1", "user2"}
	timeRange := config.TimeRange{
		Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2023, 1, 7, 23, 59, 59, 0, time.UTC),
	}
	
	jql := client.buildUserActivitiesJQL(users, timeRange)
	
	assert.Contains(t, jql, "assignee = 'user1' OR reporter = 'user1'")
	assert.Contains(t, jql, "assignee = 'user2' OR reporter = 'user2'")
	assert.Contains(t, jql, "updated >= '2023-01-01' AND updated <= '2023-01-07'")
	assert.Contains(t, jql, "ORDER BY updated DESC")
}

func TestClient_getDefaultFields(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)
	
	fields := client.getDefaultFields()
	
	expectedFields := []string{
		"id", "key", "summary", "description", "issuetype",
		"status", "priority", "reporter", "assignee", "created",
		"updated", "project", "timetracking", "worklog", "comment",
	}
	
	assert.ElementsMatch(t, expectedFields, fields)
}

func TestClient_handleErrorResponse(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
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
			responseBody:  `{"errorMessages":["Invalid credentials"]}`,
			expectedError: utils.ErrorCodeAPIUnauthorized,
			message:       "Test error",
		},
		{
			name:          "forbidden",
			statusCode:    http.StatusForbidden,
			responseBody:  `{"errorMessages":["Access denied"]}`,
			expectedError: utils.ErrorCodeAPIUnauthorized,
			message:       "Test error",
		},
		{
			name:          "not found",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"errorMessages":["Issue not found"]}`,
			expectedError: utils.ErrorCodeAPINotFound,
			message:       "Test error",
		},
		{
			name:          "rate limit",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"errorMessages":["Rate limit exceeded"]}`,
			expectedError: utils.ErrorCodeAPIRateLimit,
			message:       "Test error",
		},
		{
			name:          "bad request",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"errorMessages":["Invalid request"]}`,
			expectedError: utils.ErrorCodeAPIBadRequest,
			message:       "Test error",
		},
		{
			name:          "server error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"errorMessages":["Internal server error"]}`,
			expectedError: utils.ErrorCodeAPIServerError,
			message:       "Test error",
		},
		{
			name:          "other error",
			statusCode:    http.StatusTeapot,
			responseBody:  `{"errorMessages":["I'm a teapot"]}`,
			expectedError: utils.ErrorCodeJiraError,
			message:       "Test error",
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

func TestClient_convertIssueToActivity(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)
	
	issue := &IssueResponse{
		ID:  "12345",
		Key: "TEST-123",
		Fields: IssueFields{
			Summary:     "Test issue",
			Description: "Test description",
			IssueType:   IssueType{Name: "Bug"},
			Status:      Status{Name: "In Progress"},
			Priority:    Priority{Name: "High"},
			Reporter: UserField{
				AccountID:    "reporter123",
				DisplayName:  "Reporter User",
				EmailAddress: "reporter@example.com",
				Active:       true,
			},
			Assignee: UserField{
				AccountID:    "assignee123",
				DisplayName:  "Assignee User",
				EmailAddress: "assignee@example.com",
				Active:       true,
			},
			Created: "2023-01-01T10:00:00.000Z",
			Updated: "2023-01-02T15:30:00.000Z",
			Project: ProjectField{
				ID:   "proj123",
				Key:  "TEST",
				Name: "Test Project",
			},
			TimeTracking: TimeTracking{
				TimeSpentSeconds: 3600,
			},
		},
	}
	
	activity, err := client.convertIssueToActivity(issue)
	require.NoError(t, err)
	
	assert.Equal(t, "12345", activity.ID)
	assert.Equal(t, "TEST-123", activity.Key)
	assert.Equal(t, "Test issue", activity.Summary)
	assert.Equal(t, "Test description", activity.Description)
	assert.Equal(t, "Bug", activity.Type)
	assert.Equal(t, "In Progress", activity.Status)
	assert.Equal(t, "High", activity.Priority)
	assert.Equal(t, int64(3600), activity.TimeSpent)
	
	assert.Equal(t, "reporter123", activity.Reporter.AccountID)
	assert.Equal(t, "Reporter User", activity.Reporter.DisplayName)
	assert.Equal(t, "assignee123", activity.Assignee.AccountID)
	assert.Equal(t, "Assignee User", activity.Assignee.DisplayName)
	
	assert.Equal(t, "proj123", activity.Project.ID)
	assert.Equal(t, "TEST", activity.Project.Key)
	assert.Equal(t, "Test Project", activity.Project.Name)
	
	assert.Equal(t, time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), activity.Created)
	assert.Equal(t, time.Date(2023, 1, 2, 15, 30, 0, 0, time.UTC), activity.Updated)
}

func TestClient_convertWorklogEntry(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)
	
	entry := &WorklogEntry{
		ID: "worklog123",
		Author: UserField{
			AccountID:    "user123",
			DisplayName:  "Test User",
			EmailAddress: "user@example.com",
			Active:       true,
		},
		TimeSpentSeconds: 1800,
		Comment:          "Worked on bug fix",
		Started:          "2023-01-01T09:00:00.000Z",
		Created:          "2023-01-01T09:00:00.000Z",
		Updated:          "2023-01-01T09:00:00.000Z",
	}
	
	worklog, err := client.convertWorklogEntry(entry)
	require.NoError(t, err)
	
	assert.Equal(t, "worklog123", worklog.ID)
	assert.Equal(t, "user123", worklog.Author.AccountID)
	assert.Equal(t, "Test User", worklog.Author.DisplayName)
	assert.Equal(t, int64(1800), worklog.TimeSpent)
	assert.Equal(t, "Worked on bug fix", worklog.Description)
	assert.Equal(t, time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), worklog.Started)
	assert.Equal(t, time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), worklog.Created)
	assert.Equal(t, time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), worklog.Updated)
}

func TestClient_convertComment(t *testing.T) {
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      "https://test.atlassian.net",
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	client := NewClient(cfg, authManager, logger)
	
	comment := &CommentEntry{
		ID: "comment123",
		Author: UserField{
			AccountID:    "user123",
			DisplayName:  "Test User",
			EmailAddress: "user@example.com",
			Active:       true,
		},
		Body:    "This is a test comment",
		Created: "2023-01-01T10:00:00.000Z",
		Updated: "2023-01-01T10:00:00.000Z",
		Visibility: CommentVisibility{
			Type:  "group",
			Value: "developers",
		},
	}
	
	commentModel, err := client.convertComment(comment)
	require.NoError(t, err)
	
	assert.Equal(t, "comment123", commentModel.ID)
	assert.Equal(t, "user123", commentModel.Author.AccountID)
	assert.Equal(t, "Test User", commentModel.Author.DisplayName)
	assert.Equal(t, "This is a test comment", commentModel.Body)
	assert.Equal(t, "group:developers", commentModel.Visibility)
	assert.Equal(t, time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), commentModel.Created)
	assert.Equal(t, time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), commentModel.Updated)
}

func TestParseJiraTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Time
		shouldErr bool
	}{
		{
			name:     "ISO with timezone",
			input:    "2023-01-01T10:00:00.000+0000",
			expected: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "ISO with Z",
			input:    "2023-01-01T10:00:00.000Z",
			expected: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "RFC3339",
			input:    "2023-01-01T10:00:00Z",
			expected: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
		},
		{
			name:      "invalid format",
			input:     "invalid-date",
			shouldErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJiraTimestamp(tt.input)
			
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Convert to UTC for comparison to avoid timezone issues
				assert.True(t, tt.expected.Equal(result), "Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvertUserField(t *testing.T) {
	userField := UserField{
		AccountID:    "user123",
		DisplayName:  "Test User",
		EmailAddress: "user@example.com",
		Active:       true,
	}
	
	user := convertUserField(userField)
	
	assert.Equal(t, "user123", user.AccountID)
	assert.Equal(t, "Test User", user.DisplayName)
	assert.Equal(t, "user@example.com", user.EmailAddress)
	assert.True(t, user.Active)
}

func TestConvertProjectField(t *testing.T) {
	projectField := ProjectField{
		ID:          "proj123",
		Key:         "TEST",
		Name:        "Test Project",
		Description: "Test project description",
		Lead: UserField{
			AccountID:   "lead123",
			DisplayName: "Project Lead",
		},
	}
	
	project := convertProjectField(projectField)
	
	assert.Equal(t, "proj123", project.ID)
	assert.Equal(t, "TEST", project.Key)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "Test project description", project.Description)
	assert.Equal(t, "lead123", project.Lead.AccountID)
	assert.Equal(t, "Project Lead", project.Lead.DisplayName)
}

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name     string
		response ErrorResponse
		expected string
	}{
		{
			name: "error messages",
			response: ErrorResponse{
				ErrorMessages: []string{"Error 1", "Error 2"},
			},
			expected: "Error 1; Error 2",
		},
		{
			name: "field errors",
			response: ErrorResponse{
				Errors: map[string]string{
					"field1": "Field 1 error",
					"field2": "Field 2 error",
				},
			},
			expected: "field1: Field 1 error; field2: Field 2 error",
		},
		{
			name:     "no errors",
			response: ErrorResponse{},
			expected: "Unknown Jira error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.Error()
			
			if tt.name == "field errors" {
				// For field errors, the order might vary, so check both possible orders
				assert.True(t, 
					result == "field1: Field 1 error; field2: Field 2 error" || 
					result == "field2: Field 2 error; field1: Field 1 error")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Mock HTTP server for integration tests
func createMockJiraServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/myself":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"accountId":    "current123",
				"displayName":  "Current User",
				"emailAddress": "current@example.com",
				"active":       true,
			})
		case "/rest/api/2/search":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SearchResult{
				StartAt:    0,
				MaxResults: 50,
				Total:      1,
				Issues: []IssueResponse{
					{
						ID:  "12345",
						Key: "TEST-123",
						Fields: IssueFields{
							Summary:     "Test issue",
							Description: "Test description",
							IssueType:   IssueType{Name: "Bug"},
							Status:      Status{Name: "In Progress"},
							Priority:    Priority{Name: "High"},
							Created:     "2023-01-01T10:00:00.000Z",
							Updated:     "2023-01-02T15:30:00.000Z",
							Reporter: UserField{
								AccountID:   "reporter123",
								DisplayName: "Reporter User",
							},
							Assignee: UserField{
								AccountID:   "assignee123",
								DisplayName: "Assignee User",
							},
							Project: ProjectField{
								ID:   "proj123",
								Key:  "TEST",
								Name: "Test Project",
							},
							TimeTracking: TimeTracking{
								TimeSpentSeconds: 3600,
							},
						},
					},
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
	server := createMockJiraServer(t)
	defer server.Close()
	
	logger := utils.NewMockLogger()
	cfg := &config.Config{
		Jira: struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
		}{
			URL:      server.URL,
			Username: "testuser",
		},
	}
	
	authConfig := security.DefaultAuthConfig()
	authManager := security.NewAuthManager(authConfig, logger)
	
	// Set up mock credentials
	err := authManager.GetCredentialStore().SetJiraCredentials(security.JiraCredentials{
		Token: "test_token",
	})
	require.NoError(t, err)
	
	client := NewClient(cfg, authManager, logger)
	ctx := context.Background()
	
	// Test connection validation
	err = client.ValidateConnection(ctx)
	assert.NoError(t, err)
	
	// Test search
	_ = config.TimeRange{
		Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2023, 1, 7, 23, 59, 59, 0, time.UTC),
	}
	
	searchResult, err := client.SearchIssues(ctx, 
		"updated >= '2023-01-01' AND updated <= '2023-01-07'", 
		client.getDefaultFields(), 0, 50)
	assert.NoError(t, err)
	assert.Equal(t, 1, searchResult.Total)
	assert.Len(t, searchResult.Issues, 1)
	
	// Clean up
	authManager.GetCredentialStore().ClearAllCredentials()
}