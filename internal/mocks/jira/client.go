package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/eesa/internal/config"
	"github.com/company/eesa/internal/jira"
	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
)

// MockJiraClient is a mock implementation of the Jira client for testing
type MockJiraClient struct {
	// Mock configuration
	baseURL          string
	username         string
	logger           utils.Logger
	
	// Mock behavior controls
	shouldFailAuth   bool
	shouldFailSearch bool
	shouldFailIssue  bool
	shouldFailUser   bool
	simulateTimeout  bool
	simulateRateLimit bool
	
	// Mock data
	searchResults    *jira.SearchResult
	userActivities   []models.Activity
	worklogs         []models.Worklog
	comments         []models.Comment
	
	// Call tracking
	mu               sync.RWMutex
	authCalls        int
	searchCalls      int
	issueCalls       int
	userCalls        int
	worklogCalls     int
	commentCalls     int
	lastSearchQuery  string
	lastIssueKey     string
	lastUsers        []string
}

// NewMockJiraClient creates a new mock Jira client
func NewMockJiraClient(baseURL, username, apiToken string, logger utils.Logger) *MockJiraClient {
	return &MockJiraClient{
		baseURL:  baseURL,
		username: username,
		logger:   logger,
		
		// Default mock data
		searchResults: &jira.SearchResult{
			StartAt:    0,
			MaxResults: 50,
			Total:      2,
		},
		
		userActivities: []models.Activity{
			{
				ID:          "activity1",
				Type:        models.ActivityTypeIssue,
				UserID:      "user123",
				IssueKey:    "TEST-1",
				Title:       "Test issue 1",
				Description: "This is a test issue",
				Status:      "Open",
				Priority:    "High",
				Created:     time.Now().Add(-24 * time.Hour),
				Updated:     time.Now().Add(-1 * time.Hour),
				TimeSpent:   7200, // 2 hours in seconds
			},
			{
				ID:          "activity2",
				Type:        models.ActivityTypeIssue,
				UserID:      "user123",
				IssueKey:    "TEST-2",
				Title:       "Test issue 2",
				Description: "This is another test issue",
				Status:      "In Progress",
				Priority:    "Medium",
				Created:     time.Now().Add(-48 * time.Hour),
				Updated:     time.Now().Add(-2 * time.Hour),
				TimeSpent:   3600, // 1 hour in seconds
			},
		},
		
		worklogs: []models.Worklog{
			{
				ID:          "worklog1",
				IssueKey:    "TEST-1",
				UserID:      "user123",
				TimeSpent:   7200, // 2 hours
				Description: "Development work",
				Created:     time.Now().Add(-24 * time.Hour),
			},
		},
		
		comments: []models.Comment{
			{
				ID:       "comment1",
				IssueKey: "TEST-1",
				UserID:   "user123",
				Body:     "This is a test comment",
				Created:  time.Now().Add(-12 * time.Hour),
			},
		},
	}
}

// Configuration methods
func (m *MockJiraClient) SetFailAuth(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailAuth = fail
}

func (m *MockJiraClient) SetFailSearch(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailSearch = fail
}

func (m *MockJiraClient) SetFailIssue(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailIssue = fail
}

func (m *MockJiraClient) SetSimulateTimeout(timeout bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateTimeout = timeout
}

func (m *MockJiraClient) SetSimulateRateLimit(rateLimit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateRateLimit = rateLimit
}

// Data configuration methods
func (m *MockJiraClient) SetMockUserActivities(activities []models.Activity) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userActivities = activities
}

func (m *MockJiraClient) SetMockWorklogs(worklogs []models.Worklog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.worklogs = worklogs
}

func (m *MockJiraClient) SetMockComments(comments []models.Comment) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.comments = comments
}

// Tracking methods
func (m *MockJiraClient) GetCallCounts() (auth, search, issue, user, worklog, comment int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.authCalls, m.searchCalls, m.issueCalls, m.userCalls, m.worklogCalls, m.commentCalls
}

func (m *MockJiraClient) GetLastSearchQuery() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSearchQuery
}

func (m *MockJiraClient) GetLastIssueKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastIssueKey
}

func (m *MockJiraClient) GetLastUsers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastUsers
}

func (m *MockJiraClient) ResetCallCounts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authCalls = 0
	m.searchCalls = 0
	m.issueCalls = 0
	m.userCalls = 0
	m.worklogCalls = 0
	m.commentCalls = 0
	m.lastSearchQuery = ""
	m.lastIssueKey = ""
	m.lastUsers = nil
}

// Helper method for common behavior simulation
func (m *MockJiraClient) simulateCommonBehavior(ctx context.Context) error {
	if m.simulateTimeout {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	
	if m.shouldFailAuth {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Mock authentication failed", nil)
	}
	
	if m.simulateRateLimit {
		return utils.NewAppError(utils.ErrorCodeAPIRateLimit, "Mock rate limit exceeded", nil)
	}
	
	return nil
}

// Interface implementation
func (m *MockJiraClient) ValidateConnection(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.authCalls++
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return err
	}
	
	m.logger.Info("Mock Jira connection validation successful")
	return nil
}

func (m *MockJiraClient) GetUserActivities(ctx context.Context, users []string, timeRange config.TimeRange) ([]models.Activity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.userCalls++
	m.lastUsers = users
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailUser {
		return nil, utils.NewAppError(utils.ErrorCodeJiraError, "Mock user activities retrieval failed", nil)
	}
	
	// Filter activities by users and time range
	var filteredActivities []models.Activity
	for _, activity := range m.userActivities {
		// Check if activity user is in the requested users list
		userMatch := false
		for _, user := range users {
			if activity.UserID == user {
				userMatch = true
				break
			}
		}
		
		if userMatch && activity.Created.After(timeRange.StartTime) && activity.Created.Before(timeRange.EndTime) {
			filteredActivities = append(filteredActivities, activity)
		}
	}
	
	m.logger.Info("Mock Jira user activities retrieved",
		utils.NewField("user_count", len(users)),
		utils.NewField("activity_count", len(filteredActivities)),
	)
	
	return filteredActivities, nil
}

func (m *MockJiraClient) SearchIssues(ctx context.Context, jql string, fields []string, startAt, maxResults int) (*jira.SearchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.searchCalls++
	m.lastSearchQuery = jql
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailSearch {
		return nil, utils.NewAppError(utils.ErrorCodeJiraError, "Mock search failed", nil)
	}
	
	// Create a copy of search results with pagination
	result := &jira.SearchResult{
		StartAt:    startAt,
		MaxResults: maxResults,
		Total:      m.searchResults.Total,
	}
	
	m.logger.Info("Mock Jira search completed",
		utils.NewField("jql", jql),
		utils.NewField("start_at", startAt),
		utils.NewField("max_results", maxResults),
	)
	
	return result, nil
}

func (m *MockJiraClient) GetIssue(ctx context.Context, issueKey string, fields []string) (*models.Activity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.issueCalls++
	m.lastIssueKey = issueKey
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailIssue {
		return nil, utils.NewAppError(utils.ErrorCodeJiraError, "Mock issue retrieval failed", nil)
	}
	
	// Find activity with matching issue key
	for _, activity := range m.userActivities {
		if activity.IssueKey == issueKey {
			m.logger.Info("Mock Jira issue retrieved",
				utils.NewField("issue_key", issueKey),
			)
			return &activity, nil
		}
	}
	
	// Return a default activity if not found
	defaultActivity := &models.Activity{
		ID:          "default",
		Type:        models.ActivityTypeIssue,
		UserID:      "user123",
		IssueKey:    issueKey,
		Title:       fmt.Sprintf("Mock issue %s", issueKey),
		Description: "This is a mock issue",
		Status:      "Open",
		Priority:    "Medium",
		Created:     time.Now().Add(-24 * time.Hour),
		Updated:     time.Now(),
		TimeSpent:   3600,
	}
	
	m.logger.Info("Mock Jira issue retrieved (default)",
		utils.NewField("issue_key", issueKey),
	)
	
	return defaultActivity, nil
}

func (m *MockJiraClient) GetWorklog(ctx context.Context, issueKey string) ([]models.Worklog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.worklogCalls++
	m.lastIssueKey = issueKey
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailIssue {
		return nil, utils.NewAppError(utils.ErrorCodeJiraError, "Mock worklog retrieval failed", nil)
	}
	
	// Filter worklogs by issue key
	var filteredWorklogs []models.Worklog
	for _, worklog := range m.worklogs {
		if worklog.IssueKey == issueKey {
			filteredWorklogs = append(filteredWorklogs, worklog)
		}
	}
	
	m.logger.Info("Mock Jira worklog retrieved",
		utils.NewField("issue_key", issueKey),
		utils.NewField("worklog_count", len(filteredWorklogs)),
	)
	
	return filteredWorklogs, nil
}

func (m *MockJiraClient) GetComments(ctx context.Context, issueKey string) ([]models.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.commentCalls++
	m.lastIssueKey = issueKey
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailIssue {
		return nil, utils.NewAppError(utils.ErrorCodeJiraError, "Mock comment retrieval failed", nil)
	}
	
	// Filter comments by issue key
	var filteredComments []models.Comment
	for _, comment := range m.comments {
		if comment.IssueKey == issueKey {
			filteredComments = append(filteredComments, comment)
		}
	}
	
	m.logger.Info("Mock Jira comments retrieved",
		utils.NewField("issue_key", issueKey),
		utils.NewField("comment_count", len(filteredComments)),
	)
	
	return filteredComments, nil
}

// Verify that MockJiraClient implements the JiraClientInterface
var _ jira.JiraClientInterface = (*MockJiraClient)(nil)