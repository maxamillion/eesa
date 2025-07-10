package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/company/eesa/internal/gemini"
	"github.com/company/eesa/pkg/models"
	"github.com/company/eesa/pkg/utils"
)

// MockGeminiClient is a mock implementation of the Gemini client for testing
type MockGeminiClient struct {
	// Mock configuration
	apiKey string
	logger utils.Logger
	
	// Mock behavior controls
	shouldFailAuth         bool
	shouldFailGenerate     bool
	shouldFailModels       bool
	simulateTimeout        bool
	simulateRateLimit      bool
	simulateQuotaExceeded  bool
	
	// Mock data
	summaryResponse   *gemini.SummaryResponse
	generateResponse  *gemini.GenerateResponse
	modelsResponse    *gemini.ModelsResponse
	
	// Call tracking
	mu               sync.RWMutex
	generateCalls    int
	summaryCalls     int
	modelsCalls      int
	validateCalls    int
	lastPrompt       string
	lastActivities   []models.Activity
	
	// Response customization
	customResponses  map[string]string
	responseDelay    time.Duration
}

// NewMockGeminiClient creates a new mock Gemini client
func NewMockGeminiClient(apiKey string, logger utils.Logger) *MockGeminiClient {
	return &MockGeminiClient{
		apiKey: apiKey,
		logger: logger,
		
		// Default mock data
		summaryResponse: &gemini.SummaryResponse{
			Summary: "## Executive Summary\n\nBased on the analyzed activities, the team has been highly productive with 2 key issues addressed. The focus has been on critical bug fixes and feature development, showing strong progress across all priorities.\n\n### Key Achievements\n- Resolved high-priority issues efficiently\n- Maintained good development velocity\n- Strong collaboration and communication\n\n### Recommendations\n- Continue current development pace\n- Monitor upcoming deadlines\n- Ensure proper testing coverage",
			KeyPoints: []string{
				"2 issues completed successfully",
				"Strong focus on high-priority items",
				"Good time allocation across tasks",
				"Effective team collaboration",
			},
			Recommendations: []string{
				"Maintain current development velocity",
				"Ensure comprehensive testing",
				"Monitor upcoming milestone deadlines",
			},
			TokensUsed: 150,
		},
		
		generateResponse: &gemini.GenerateResponse{
			Candidates: []gemini.Candidate{
				{
					Content: gemini.Content{
						Parts: []gemini.Part{
							{
								Text: "This is a mock response from Gemini AI. The content has been generated for testing purposes and demonstrates the AI's ability to create comprehensive, contextual responses based on the provided input.",
							},
						},
					},
				},
			},
			UsageMetadata: &gemini.UsageMetadata{
				PromptTokenCount:     100,
				CandidatesTokenCount: 50,
				TotalTokenCount:      150,
			},
		},
		
		modelsResponse: &gemini.ModelsResponse{
			Models: []gemini.Model{
				{
					Name:        "gemini-pro",
					DisplayName: "Gemini Pro",
					Description: "The best model for scaling across a wide range of tasks",
				},
				{
					Name:        "gemini-pro-vision",
					DisplayName: "Gemini Pro Vision", 
					Description: "The best image understanding model to handle a broad range of applications",
				},
			},
		},
		
		customResponses: make(map[string]string),
		responseDelay:   0,
	}
}

// Configuration methods
func (m *MockGeminiClient) SetFailAuth(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailAuth = fail
}

func (m *MockGeminiClient) SetFailGenerate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailGenerate = fail
}

func (m *MockGeminiClient) SetFailModels(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailModels = fail
}

func (m *MockGeminiClient) SetSimulateTimeout(timeout bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateTimeout = timeout
}

func (m *MockGeminiClient) SetSimulateRateLimit(rateLimit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateRateLimit = rateLimit
}

func (m *MockGeminiClient) SetSimulateQuotaExceeded(quotaExceeded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateQuotaExceeded = quotaExceeded
}

// Data configuration methods
func (m *MockGeminiClient) SetMockSummaryResponse(response *gemini.SummaryResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.summaryResponse = response
}

func (m *MockGeminiClient) SetMockGenerateResponse(response *gemini.GenerateResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.generateResponse = response
}

func (m *MockGeminiClient) SetMockModelsResponse(response *gemini.ModelsResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelsResponse = response
}

func (m *MockGeminiClient) SetCustomResponse(prompt, response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customResponses[prompt] = response
}

func (m *MockGeminiClient) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// Tracking methods
func (m *MockGeminiClient) GetCallCounts() (summary, generate, models, validate int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.summaryCalls, m.generateCalls, m.modelsCalls, m.validateCalls
}

func (m *MockGeminiClient) GetLastPrompt() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastPrompt
}

func (m *MockGeminiClient) GetLastActivities() []models.Activity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastActivities
}

func (m *MockGeminiClient) ResetCallCounts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.generateCalls = 0
	m.summaryCalls = 0
	m.modelsCalls = 0
	m.validateCalls = 0
	m.lastPrompt = ""
	m.lastActivities = nil
}

// Helper method for common behavior simulation
func (m *MockGeminiClient) simulateCommonBehavior(ctx context.Context) error {
	// Simulate response delay
	if m.responseDelay > 0 {
		time.Sleep(m.responseDelay)
	}
	
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
	
	if m.simulateQuotaExceeded {
		return utils.NewAppError(utils.ErrorCodeAPIRateLimit, "Mock quota exceeded", nil)
	}
	
	return nil
}

// Interface implementation
func (m *MockGeminiClient) GenerateSummary(ctx context.Context, activities []models.Activity, prompt string) (*gemini.SummaryResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.summaryCalls++
	m.lastActivities = activities
	m.lastPrompt = prompt
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailGenerate {
		return nil, utils.NewAppError(utils.ErrorCodeGeminiError, "Mock summary generation failed", nil)
	}
	
	// Check for custom response based on prompt
	if customResponse, exists := m.customResponses[prompt]; exists {
		response := &gemini.SummaryResponse{
			Summary: customResponse,
			KeyPoints: []string{
				"Custom response generated",
				"Based on provided prompt",
			},
			Recommendations: []string{
				"Review custom response",
			},
			TokensUsed: len(strings.Split(customResponse, " ")),
		}
		
		m.logger.Info("Mock Gemini summary generated (custom)",
			utils.NewField("activity_count", len(activities)),
			utils.NewField("prompt_length", len(prompt)),
			utils.NewField("summary_length", len(customResponse)),
		)
		
		return response, nil
	}
	
	// Generate contextual summary based on activities
	response := m.generateContextualSummary(activities, prompt)
	
	m.logger.Info("Mock Gemini summary generated",
		utils.NewField("activity_count", len(activities)),
		utils.NewField("prompt_length", len(prompt)),
		utils.NewField("summary_length", len(response.Summary)),
	)
	
	return response, nil
}

func (m *MockGeminiClient) ValidateAPIKey(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.validateCalls++
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return err
	}
	
	m.logger.Info("Mock Gemini API key validation successful")
	return nil
}

func (m *MockGeminiClient) ListModels(ctx context.Context) (*gemini.ModelsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.modelsCalls++
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailModels {
		return nil, utils.NewAppError(utils.ErrorCodeGeminiError, "Mock models listing failed", nil)
	}
	
	m.logger.Info("Mock Gemini models listed",
		utils.NewField("model_count", len(m.modelsResponse.Models)),
	)
	
	return m.modelsResponse, nil
}

func (m *MockGeminiClient) GenerateContent(ctx context.Context, request *gemini.GenerateRequest) (*gemini.GenerateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.generateCalls++
	
	// Extract prompt from request
	var prompt string
	if len(request.Contents) > 0 && len(request.Contents[0].Parts) > 0 {
		prompt = request.Contents[0].Parts[0].Text
		m.lastPrompt = prompt
	}
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailGenerate {
		return nil, utils.NewAppError(utils.ErrorCodeGeminiError, "Mock content generation failed", nil)
	}
	
	// Check for custom response
	if customResponse, exists := m.customResponses[prompt]; exists {
		response := &gemini.GenerateResponse{
			Candidates: []gemini.Candidate{
				{
					Content: gemini.Content{
						Parts: []gemini.Part{
							{
								Text: customResponse,
							},
						},
					},
				},
			},
			UsageMetadata: &gemini.UsageMetadata{
				PromptTokenCount:     len(strings.Split(prompt, " ")),
				CandidatesTokenCount: len(strings.Split(customResponse, " ")),
				TotalTokenCount:      len(strings.Split(prompt, " ")) + len(strings.Split(customResponse, " ")),
			},
		}
		
		m.logger.Info("Mock Gemini content generated (custom response)",
			utils.NewField("prompt_length", len(prompt)),
			utils.NewField("response_length", len(customResponse)),
		)
		
		return response, nil
	}
	
	// Generate contextual response based on prompt
	response := m.generateContextualResponse(prompt)
	response.UsageMetadata = &gemini.UsageMetadata{
		PromptTokenCount:     len(strings.Split(prompt, " ")),
		CandidatesTokenCount: len(strings.Split(response.Candidates[0].Content.Parts[0].Text, " ")),
		TotalTokenCount:      len(strings.Split(prompt, " ")) + len(strings.Split(response.Candidates[0].Content.Parts[0].Text, " ")),
	}
	
	m.logger.Info("Mock Gemini content generated",
		utils.NewField("prompt_length", len(prompt)),
		utils.NewField("response_length", len(response.Candidates[0].Content.Parts[0].Text)),
	)
	
	return response, nil
}

// Helper methods for response generation
func (m *MockGeminiClient) generateContextualSummary(activities []models.Activity, prompt string) *gemini.SummaryResponse {
	// Analyze activities to generate contextual summary
	totalActivities := len(activities)
	var highPriorityCount, completedCount int
	var totalTimeSpent int64
	
	for _, activity := range activities {
		if activity.Priority == "High" || activity.Priority == "Critical" {
			highPriorityCount++
		}
		if activity.Status == "Done" || activity.Status == "Closed" || activity.Status == "Resolved" {
			completedCount++
		}
		totalTimeSpent += activity.TimeSpent
	}
	
	summary := fmt.Sprintf("## Executive Summary\n\nBased on the analysis of %d activities, the team has shown strong productivity and focus. ", totalActivities)
	
	if highPriorityCount > 0 {
		summary += fmt.Sprintf("%d high-priority items were addressed, demonstrating good prioritization. ", highPriorityCount)
	}
	
	if completedCount > 0 {
		summary += fmt.Sprintf("%d activities have been completed successfully. ", completedCount)
	}
	
	summary += fmt.Sprintf("Total time spent across all activities: %.1f hours.\n\n", float64(totalTimeSpent)/3600)
	
	summary += "### Key Achievements\n"
	summary += "- Effective task prioritization and execution\n"
	summary += "- Consistent progress across multiple work streams\n"
	summary += "- Strong time management and allocation\n\n"
	
	summary += "### Recommendations\n"
	summary += "- Continue current development velocity\n"
	summary += "- Monitor upcoming deadlines and dependencies\n"
	summary += "- Ensure adequate testing and code review processes"
	
	keyPoints := []string{
		fmt.Sprintf("%d total activities analyzed", totalActivities),
		fmt.Sprintf("%d high-priority items addressed", highPriorityCount),
		fmt.Sprintf("%d activities completed", completedCount),
		fmt.Sprintf("%.1f hours total time investment", float64(totalTimeSpent)/3600),
	}
	
	recommendations := []string{
		"Maintain current productivity levels",
		"Focus on completing in-progress items",
		"Ensure proper documentation and testing",
	}
	
	return &gemini.SummaryResponse{
		Summary:         summary,
		KeyPoints:       keyPoints,
		Recommendations: recommendations,
		TokensUsed:      len(strings.Split(summary, " ")),
	}
}

func (m *MockGeminiClient) generateContextualResponse(prompt string) *gemini.GenerateResponse {
	lowerPrompt := strings.ToLower(prompt)
	
	var responseText string
	
	// Generate different responses based on prompt content
	if strings.Contains(lowerPrompt, "executive summary") || strings.Contains(lowerPrompt, "summary") {
		responseText = "## Executive Summary\n\nBased on the provided information, here is a comprehensive executive summary:\n\n### Key Highlights\n- Critical issues have been identified and prioritized\n- Team productivity metrics show positive trends\n- Resource allocation is optimized for maximum impact\n\n### Recommendations\n- Continue current development trajectory\n- Address high-priority issues first\n- Monitor progress weekly\n\n### Next Steps\n- Implement recommended changes\n- Schedule follow-up review\n- Track key performance indicators"
	} else if strings.Contains(lowerPrompt, "jira") || strings.Contains(lowerPrompt, "issues") {
		responseText = "## Jira Issues Analysis\n\nThe following analysis covers the key issues and their impact:\n\n### Issue Breakdown\n- **High Priority**: Critical bugs requiring immediate attention\n- **Medium Priority**: Feature requests and improvements\n- **Low Priority**: Minor enhancements and documentation\n\n### Recommendations\n- Focus on high-priority issues first\n- Allocate resources based on business impact\n- Establish clear timelines for resolution"
	} else if strings.Contains(lowerPrompt, "test") || strings.Contains(lowerPrompt, "testing") {
		responseText = "This is a mock response generated for testing purposes. The Gemini AI client is functioning correctly and can generate contextual responses based on the input prompt. This response demonstrates the mock's ability to simulate real AI-generated content."
	} else {
		responseText = "This is a mock response from the Gemini AI client. The system is working correctly and can generate appropriate responses based on the input provided. This demonstrates the functionality of the mock implementation."
	}
	
	return &gemini.GenerateResponse{
		Candidates: []gemini.Candidate{
			{
				Content: gemini.Content{
					Parts: []gemini.Part{
						{
							Text: responseText,
						},
					},
				},
			},
		},
	}
}

// Verify that MockGeminiClient implements the GeminiClientInterface
var _ gemini.GeminiClientInterface = (*MockGeminiClient)(nil)