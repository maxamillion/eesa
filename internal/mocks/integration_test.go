package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/company/eesa/internal/config"
	"github.com/company/eesa/internal/gemini"
	"github.com/company/eesa/internal/mocks/gdocs"
	"github.com/company/eesa/internal/mocks/gemini"
	"github.com/company/eesa/internal/mocks/jira"
	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockServicesIntegration tests the integration of all mock services
func TestMockServicesIntegration(t *testing.T) {
	logger := utils.NewMockLogger()
	ctx := context.Background()
	
	// Initialize mock clients
	jiraClient := jira.NewMockJiraClient("https://test.atlassian.net", "user", "token", logger)
	geminiClient := gemini.NewMockGeminiClient("test-api-key", logger)
	gdocsClient := gdocs.NewMockGoogleDocsClient(logger)
	
	t.Run("Jira Mock Service", func(t *testing.T) {
		// Test successful operations
		err := jiraClient.ValidateConnection(ctx)
		assert.NoError(t, err)
		
		// Test user activities
		timeRange := config.TimeRange{
			StartTime: time.Now().Add(-7 * 24 * time.Hour),
			EndTime:   time.Now(),
		}
		
		activities, err := jiraClient.GetUserActivities(ctx, []string{"user123"}, timeRange)
		require.NoError(t, err)
		assert.NotNil(t, activities)
		assert.Equal(t, 2, len(activities))
		assert.Equal(t, "TEST-1", activities[0].IssueKey)
		
		// Test search issues
		searchResult, err := jiraClient.SearchIssues(ctx, "project = TEST", []string{"summary", "status"}, 0, 10)
		require.NoError(t, err)
		assert.NotNil(t, searchResult)
		assert.Equal(t, 0, searchResult.StartAt)
		assert.Equal(t, 10, searchResult.MaxResults)
		
		// Test get issue
		issue, err := jiraClient.GetIssue(ctx, "TEST-1", []string{"summary", "status"})
		require.NoError(t, err)
		assert.NotNil(t, issue)
		assert.Equal(t, "TEST-1", issue.IssueKey)
		
		// Test worklogs
		worklogs, err := jiraClient.GetWorklog(ctx, "TEST-1")
		require.NoError(t, err)
		assert.NotNil(t, worklogs)
		assert.Equal(t, 1, len(worklogs))
		
		// Test comments
		comments, err := jiraClient.GetComments(ctx, "TEST-1")
		require.NoError(t, err)
		assert.NotNil(t, comments)
		assert.Equal(t, 1, len(comments))
		
		// Test call counting
		auth, search, issue, user, worklog, comment := jiraClient.GetCallCounts()
		assert.Equal(t, 1, auth)
		assert.Equal(t, 1, search)
		assert.Equal(t, 1, issue)
		assert.Equal(t, 1, user)
		assert.Equal(t, 1, worklog)
		assert.Equal(t, 1, comment)
		
		// Test tracking
		assert.Equal(t, "project = TEST", jiraClient.GetLastSearchQuery())
		assert.Equal(t, "TEST-1", jiraClient.GetLastIssueKey())
		assert.Equal(t, []string{"user123"}, jiraClient.GetLastUsers())
		
		// Test failure scenarios
		jiraClient.SetFailSearch(true)
		_, err = jiraClient.SearchIssues(ctx, "project = TEST", []string{}, 0, 10)
		assert.Error(t, err)
		
		jiraClient.SetFailSearch(false)
		jiraClient.SetSimulateRateLimit(true)
		_, err = jiraClient.SearchIssues(ctx, "project = TEST", []string{}, 0, 10)
		assert.Error(t, err)
		
		// Reset for next tests
		jiraClient.SetSimulateRateLimit(false)
		jiraClient.ResetCallCounts()
	})
	
	t.Run("Gemini Mock Service", func(t *testing.T) {
		// Test API key validation
		err := geminiClient.ValidateAPIKey(ctx)
		assert.NoError(t, err)
		
		// Test list models
		modelsResponse, err := geminiClient.ListModels(ctx)
		require.NoError(t, err)
		assert.NotNil(t, modelsResponse)
		assert.Equal(t, 2, len(modelsResponse.Models))
		assert.Equal(t, "gemini-pro", modelsResponse.Models[0].Name)
		
		// Test generate summary
		testActivities := []models.Activity{
			{
				ID:          "test1",
				Type:        models.ActivityTypeIssue,
				UserID:      "user123",
				IssueKey:    "TEST-1",
				Title:       "Test issue 1",
				Status:      "Done",
				Priority:    "High",
				TimeSpent:   7200,
			},
		}
		
		summaryResponse, err := geminiClient.GenerateSummary(ctx, testActivities, "Generate executive summary")
		require.NoError(t, err)
		assert.NotNil(t, summaryResponse)
		assert.NotEmpty(t, summaryResponse.Summary)
		assert.True(t, len(summaryResponse.KeyPoints) > 0)
		assert.True(t, len(summaryResponse.Recommendations) > 0)
		
		// Test generate content
		generateRequest := &gemini.GenerateRequest{
			Contents: []gemini.Content{
				{
					Parts: []gemini.Part{
						{
							Text: "Generate a test response",
						},
					},
				},
			},
		}
		
		generateResponse, err := geminiClient.GenerateContent(ctx, generateRequest)
		require.NoError(t, err)
		assert.NotNil(t, generateResponse)
		assert.Equal(t, 1, len(generateResponse.Candidates))
		assert.NotEmpty(t, generateResponse.Candidates[0].Content.Parts[0].Text)
		assert.NotNil(t, generateResponse.UsageMetadata)
		
		// Test custom response
		geminiClient.SetCustomResponse("custom test", "This is a custom response")
		customRequest := &gemini.GenerateRequest{
			Contents: []gemini.Content{
				{
					Parts: []gemini.Part{
						{
							Text: "custom test",
						},
					},
				},
			},
		}
		customResponse, err := geminiClient.GenerateContent(ctx, customRequest)
		require.NoError(t, err)
		assert.Equal(t, "This is a custom response", customResponse.Candidates[0].Content.Parts[0].Text)
		
		// Test call counting
		summary, generate, models, validate := geminiClient.GetCallCounts()
		assert.Equal(t, 1, summary)
		assert.Equal(t, 2, generate) // GenerateContent called twice
		assert.Equal(t, 1, models)
		assert.Equal(t, 1, validate)
		
		// Test tracking
		assert.Equal(t, "custom test", geminiClient.GetLastPrompt())
		assert.Equal(t, testActivities, geminiClient.GetLastActivities())
		
		// Test failure scenarios
		geminiClient.SetFailGenerate(true)
		_, err = geminiClient.GenerateContent(ctx, generateRequest)
		assert.Error(t, err)
		
		geminiClient.SetFailGenerate(false)
		geminiClient.SetSimulateRateLimit(true)
		_, err = geminiClient.GenerateContent(ctx, generateRequest)
		assert.Error(t, err)
		
		// Reset for next tests
		geminiClient.SetSimulateRateLimit(false)
		geminiClient.ResetCallCounts()
	})
	
	t.Run("Google Docs Mock Service", func(t *testing.T) {
		// Test credential validation
		err := gdocsClient.ValidateCredentials(ctx)
		assert.NoError(t, err)
		
		// Test document creation
		createResponse, err := gdocsClient.CreateDocument(ctx, "Test Document", "This is test content")
		require.NoError(t, err)
		assert.NotNil(t, createResponse)
		assert.NotEmpty(t, createResponse.DocumentID)
		assert.Equal(t, "Test Document", createResponse.Title)
		
		documentID := createResponse.DocumentID
		
		// Test document retrieval
		document, err := gdocsClient.GetDocument(ctx, documentID)
		require.NoError(t, err)
		assert.NotNil(t, document)
		assert.Equal(t, documentID, document.DocumentID)
		assert.Equal(t, "Test Document", document.Title)
		
		// Test document sharing
		err = gdocsClient.ShareDocument(ctx, documentID, []string{"test@example.com"}, "reader")
		require.NoError(t, err)
		
		// Test executive summary creation
		metadata := map[string]interface{}{
			"author": "test user",
			"date":   time.Now().Format("2006-01-02"),
		}
		execSummaryResponse, err := gdocsClient.CreateExecutiveSummaryDocument(ctx, "Executive Summary", "This is test summary content", metadata)
		require.NoError(t, err)
		assert.NotNil(t, execSummaryResponse)
		assert.NotEmpty(t, execSummaryResponse.DocumentID)
		assert.Equal(t, "Executive Summary", execSummaryResponse.Title)
		
		// Test call counting
		create, update, share, get, validate, execSummary := gdocsClient.GetCallCounts()
		assert.Equal(t, 1, create) // CreateDocument
		assert.Equal(t, 0, update)
		assert.Equal(t, 1, share)
		assert.Equal(t, 1, get)
		assert.Equal(t, 1, validate)
		assert.Equal(t, 1, execSummary) // CreateExecutiveSummaryDocument
		
		// Test tracking
		assert.Equal(t, "Executive Summary", gdocsClient.GetLastTitle())
		assert.Equal(t, []string{"test@example.com"}, gdocsClient.GetLastEmails())
		
		// Test stored data
		storedDocs := gdocsClient.GetStoredDocuments()
		assert.Equal(t, 2, len(storedDocs))
		
		sharedWith := gdocsClient.GetSharedWith()
		assert.Equal(t, 1, len(sharedWith))
		assert.Equal(t, 1, len(sharedWith[documentID]))
		
		// Test failure scenarios
		gdocsClient.SetFailCreate(true)
		_, err = gdocsClient.CreateDocument(ctx, "Should Fail", "content")
		assert.Error(t, err)
		
		gdocsClient.SetFailCreate(false)
		gdocsClient.SetSimulateRateLimit(true)
		_, err = gdocsClient.CreateDocument(ctx, "Should Rate Limit", "content")
		assert.Error(t, err)
		
		// Reset for next tests
		gdocsClient.SetSimulateRateLimit(false)
		gdocsClient.ResetCallCounts()
		gdocsClient.ClearStoredData()
	})
	
	t.Run("Cross-Service Integration", func(t *testing.T) {
		// Simulate a full workflow: Jira -> Gemini -> Google Docs
		
		// 1. Get data from Jira
		timeRange := config.TimeRange{
			StartTime: time.Now().Add(-7 * 24 * time.Hour),
			EndTime:   time.Now(),
		}
		
		activities, err := jiraClient.GetUserActivities(ctx, []string{"user123"}, timeRange)
		require.NoError(t, err)
		assert.NotNil(t, activities)
		
		// 2. Generate summary with Gemini
		summaryResponse, err := geminiClient.GenerateSummary(ctx, activities, "Generate weekly executive summary")
		require.NoError(t, err)
		assert.NotNil(t, summaryResponse)
		
		summaryContent := summaryResponse.Summary
		assert.NotEmpty(t, summaryContent)
		
		// 3. Create Google Docs document
		metadata := map[string]interface{}{
			"period": "weekly",
			"user_count": len([]string{"user123"}),
			"activity_count": len(activities),
		}
		docResponse, err := gdocsClient.CreateExecutiveSummaryDocument(ctx, "Weekly Executive Summary", summaryContent, metadata)
		require.NoError(t, err)
		assert.NotNil(t, docResponse)
		
		// 4. Share the document
		err = gdocsClient.ShareDocument(ctx, docResponse.DocumentID, []string{"manager@example.com"}, "reader")
		require.NoError(t, err)
		
		// Verify the workflow worked
		storedDocs := gdocsClient.GetStoredDocuments()
		assert.Equal(t, 1, len(storedDocs))
		
		sharedWith := gdocsClient.GetSharedWith()
		assert.Equal(t, 1, len(sharedWith))
		
		// Verify URL generation
		documentURL := gdocsClient.GetDocumentURL(docResponse.DocumentID)
		assert.Contains(t, documentURL, "docs.google.com")
		assert.Contains(t, documentURL, docResponse.DocumentID)
		
		// Test plain text extraction
		document, err := gdocsClient.GetDocument(ctx, docResponse.DocumentID)
		require.NoError(t, err)
		
		plainText := gdocsClient.ExtractPlainText(document)
		assert.NotEmpty(t, plainText)
		assert.Contains(t, plainText, "Weekly Executive Summary")
	})
	
	t.Run("Timeout and Context Handling", func(t *testing.T) {
		// Test timeout handling with context
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		
		// Enable timeout simulation
		jiraClient.SetSimulateTimeout(true)
		geminiClient.SetSimulateTimeout(true)
		gdocsClient.SetSimulateTimeout(true)
		
		// Test Jira timeout
		_, err := jiraClient.SearchIssues(timeoutCtx, "project = TEST", []string{}, 0, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
		
		// Test Gemini timeout
		generateRequest := &gemini.GenerateRequest{
			Contents: []gemini.Content{
				{
					Parts: []gemini.Part{
						{
							Text: "test",
						},
					},
				},
			},
		}
		_, err = geminiClient.GenerateContent(timeoutCtx, generateRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
		
		// Test Google Docs timeout
		_, err = gdocsClient.CreateDocument(timeoutCtx, "Test", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
		
		// Disable timeout simulation
		jiraClient.SetSimulateTimeout(false)
		geminiClient.SetSimulateTimeout(false)
		gdocsClient.SetSimulateTimeout(false)
	})
	
	t.Run("Response Delay Testing", func(t *testing.T) {
		// Test response delay functionality
		delay := 10 * time.Millisecond
		geminiClient.SetResponseDelay(delay)
		gdocsClient.SetResponseDelay(delay)
		
		generateRequest := &gemini.GenerateRequest{
			Contents: []gemini.Content{
				{
					Parts: []gemini.Part{
						{
							Text: "test",
						},
					},
				},
			},
		}
		
		start := time.Now()
		_, err := geminiClient.GenerateContent(ctx, generateRequest)
		elapsed := time.Since(start)
		
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, elapsed, delay)
		
		start = time.Now()
		_, err = gdocsClient.CreateDocument(ctx, "Test", "content")
		elapsed = time.Since(start)
		
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, elapsed, delay)
		
		// Reset delays
		geminiClient.SetResponseDelay(0)
		gdocsClient.SetResponseDelay(0)
	})
}