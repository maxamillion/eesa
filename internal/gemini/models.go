package gemini

import (
	"time"

	"github.com/company/eesa/pkg/models"
)

// GenerateRequest represents a request to Gemini's generateContent API
type GenerateRequest struct {
	Contents         []Content          `json:"contents"`
	GenerationConfig *GenerationConfig  `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting    `json:"safetySettings,omitempty"`
	Tools            []Tool             `json:"tools,omitempty"`
	SystemInstruction *Content          `json:"systemInstruction,omitempty"`
}

// Content represents content in a Gemini request/response
type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

// Part represents a part of content (text, image, etc.)
type Part struct {
	Text         string     `json:"text,omitempty"`
	InlineData   *InlineData `json:"inlineData,omitempty"`
	FileData     *FileData   `json:"fileData,omitempty"`
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

// InlineData represents inline data (e.g., base64 encoded image)
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// FileData represents file data
type FileData struct {
	MimeType string `json:"mimeType"`
	FileURI  string `json:"fileUri"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// FunctionResponse represents a function response
type FunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature     *float32  `json:"temperature,omitempty"`
	TopP            *float32  `json:"topP,omitempty"`
	TopK            *int32    `json:"topK,omitempty"`
	MaxTokens       *int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string  `json:"stopSequences,omitempty"`
	CandidateCount  *int32    `json:"candidateCount,omitempty"`
	PresencePenalty *float32  `json:"presencePenalty,omitempty"`
	FrequencyPenalty *float32 `json:"frequencyPenalty,omitempty"`
}

// SafetySetting represents a safety setting
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// Tool represents a tool definition
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

// FunctionDeclaration represents a function declaration
type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GenerateResponse represents a response from Gemini's generateContent API
type GenerateResponse struct {
	Candidates     []Candidate     `json:"candidates"`
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *UsageMetadata  `json:"usageMetadata,omitempty"`
}

// Candidate represents a generated candidate
type Candidate struct {
	Content       Content         `json:"content"`
	FinishReason  string          `json:"finishReason"`
	Index         int             `json:"index"`
	SafetyRatings []SafetyRating  `json:"safetyRatings"`
	CitationMetadata *CitationMetadata `json:"citationMetadata,omitempty"`
}

// SafetyRating represents a safety rating
type SafetyRating struct {
	Category         string  `json:"category"`
	Probability      string  `json:"probability"`
	Blocked          bool    `json:"blocked"`
	BlockedReason    string  `json:"blockedReason,omitempty"`
}

// CitationMetadata represents citation metadata
type CitationMetadata struct {
	CitationSources []CitationSource `json:"citationSources"`
}

// CitationSource represents a citation source
type CitationSource struct {
	StartIndex int    `json:"startIndex"`
	EndIndex   int    `json:"endIndex"`
	URI        string `json:"uri,omitempty"`
	License    string `json:"license,omitempty"`
}

// PromptFeedback represents prompt feedback
type PromptFeedback struct {
	BlockReason   string         `json:"blockReason,omitempty"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

// UsageMetadata represents usage metadata
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// ModelsResponse represents a response from the models API
type ModelsResponse struct {
	Models []Model `json:"models"`
}

// Model represents a Gemini model
type Model struct {
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	DisplayName          string   `json:"displayName"`
	Description          string   `json:"description"`
	InputTokenLimit      int      `json:"inputTokenLimit"`
	OutputTokenLimit     int      `json:"outputTokenLimit"`
	SupportedGeneration  []string `json:"supportedGenerationMethods"`
	Temperature          *float32 `json:"temperature,omitempty"`
	TopP                 *float32 `json:"topP,omitempty"`
	TopK                 *int32   `json:"topK,omitempty"`
}

// SummaryResponse represents a response from summary generation
type SummaryResponse struct {
	Summary     string           `json:"summary"`
	TokensUsed  int              `json:"tokensUsed"`
	Model       string           `json:"model"`
	Temperature float32          `json:"temperature"`
	GeneratedAt time.Time        `json:"generatedAt"`
	Activities  []models.Activity `json:"activities"`
	Metadata    *SummaryMetadata `json:"metadata,omitempty"`
}

// SummaryMetadata represents metadata about the summary generation
type SummaryMetadata struct {
	ProjectCount       int                     `json:"projectCount"`
	CompletionRate     float64                 `json:"completionRate"`
	TotalTimeSpent     int64                   `json:"totalTimeSpent"`
	TopContributors    []models.User           `json:"topContributors"`
	ActivityBreakdown  map[string]int          `json:"activityBreakdown"`
	PriorityBreakdown  map[string]int          `json:"priorityBreakdown"`
	StatusBreakdown    map[string]int          `json:"statusBreakdown"`
	SafetyRatings      []SafetyRating          `json:"safetyRatings"`
	CitationMetadata   *CitationMetadata       `json:"citationMetadata,omitempty"`
}

// ErrorResponse represents a Gemini error response
type ErrorResponse struct {
	ErrorInfo GeminiError `json:"error"`
}

// GeminiError represents a Gemini API error
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Type        string                 `json:"@type"`
	Reason      string                 `json:"reason,omitempty"`
	Domain      string                 `json:"domain,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// Error returns the error message from GeminiError
func (e *GeminiError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Unknown Gemini API error"
}

// Error returns the error message from ErrorResponse
func (e *ErrorResponse) Error() string {
	return e.ErrorInfo.Error()
}

// PromptTemplate represents a prompt template for summary generation
type PromptTemplate struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Template     string            `json:"template"`
	Variables    []string          `json:"variables"`
	Metadata     map[string]string `json:"metadata"`
	Version      string            `json:"version"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

// SummaryRequest represents a request for summary generation
type SummaryRequest struct {
	Activities       []models.Activity `json:"activities"`
	CustomPrompt     string            `json:"customPrompt,omitempty"`
	PromptTemplate   *PromptTemplate   `json:"promptTemplate,omitempty"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
	IncludeMetadata  bool              `json:"includeMetadata"`
	Format           string            `json:"format"` // "executive", "detailed", "bullet_points"
	Audience         string            `json:"audience"` // "executives", "managers", "team"
	TimeRange        string            `json:"timeRange,omitempty"`
}

// ValidationResult represents the result of validating Gemini connection
type ValidationResult struct {
	Valid         bool      `json:"valid"`
	Message       string    `json:"message"`
	ModelInfo     *Model    `json:"modelInfo,omitempty"`
	Capabilities  []string  `json:"capabilities,omitempty"`
	TokenLimits   *TokenLimits `json:"tokenLimits,omitempty"`
	TestedAt      time.Time `json:"testedAt"`
}

// TokenLimits represents token limits for the model
type TokenLimits struct {
	InputTokenLimit  int `json:"inputTokenLimit"`
	OutputTokenLimit int `json:"outputTokenLimit"`
	ContextWindow    int `json:"contextWindow"`
}

// BatchRequest represents a batch processing request
type BatchRequest struct {
	Requests []SummaryRequest `json:"requests"`
	Metadata map[string]string `json:"metadata"`
}

// BatchResponse represents a batch processing response
type BatchResponse struct {
	Responses []SummaryResponse `json:"responses"`
	Success   int               `json:"success"`
	Failed    int               `json:"failed"`
	Errors    []BatchError      `json:"errors,omitempty"`
	Metadata  map[string]string `json:"metadata"`
}

// BatchError represents an error in batch processing
type BatchError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Constants for safety categories
const (
	SafetyCategoryHarassment        = "HARM_CATEGORY_HARASSMENT"
	SafetyCategoryHateSpeech        = "HARM_CATEGORY_HATE_SPEECH"
	SafetyCategorySexuallyExplicit  = "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	SafetyCategoryDangerousContent  = "HARM_CATEGORY_DANGEROUS_CONTENT"
	SafetyCategoryMedicalAdvice     = "HARM_CATEGORY_MEDICAL"
	SafetyCategoryDerogatory        = "HARM_CATEGORY_DEROGATORY"
	SafetyCategoryToxicity          = "HARM_CATEGORY_TOXICITY"
	SafetyCategoryViolence          = "HARM_CATEGORY_VIOLENCE"
)

// Constants for safety thresholds
const (
	SafetyThresholdBlockNone         = "BLOCK_NONE"
	SafetyThresholdBlockLowAndAbove  = "BLOCK_LOW_AND_ABOVE"
	SafetyThresholdBlockMedAndAbove  = "BLOCK_MEDIUM_AND_ABOVE"
	SafetyThresholdBlockHighAndAbove = "BLOCK_HIGH_AND_ABOVE"
	SafetyThresholdUnspecified       = "HARM_BLOCK_THRESHOLD_UNSPECIFIED"
)

// Constants for finish reasons
const (
	FinishReasonStop             = "STOP"
	FinishReasonMaxTokens        = "MAX_TOKENS"
	FinishReasonSafety           = "SAFETY"
	FinishReasonRecitation       = "RECITATION"
	FinishReasonOther            = "OTHER"
	FinishReasonUnspecified      = "FINISH_REASON_UNSPECIFIED"
)

// Constants for block reasons
const (
	BlockReasonSafety      = "SAFETY"
	BlockReasonOther       = "OTHER"
	BlockReasonUnspecified = "BLOCKED_REASON_UNSPECIFIED"
)

// Constants for model names
const (
	ModelGeminiPro       = "gemini-pro"
	ModelGeminiProVision = "gemini-pro-vision"
	ModelGemini15Pro     = "gemini-1.5-pro"
	ModelGemini15Flash   = "gemini-1.5-flash"
)

// Constants for content roles
const (
	RoleUser      = "user"
	RoleModel     = "model"
	RoleSystem    = "system"
	RoleFunction  = "function"
)

// Constants for summary formats
const (
	FormatExecutive    = "executive"
	FormatDetailed     = "detailed"
	FormatBulletPoints = "bullet_points"
	FormatNarrative    = "narrative"
	FormatMetrics      = "metrics"
)

// Constants for audience types
const (
	AudienceExecutives = "executives"
	AudienceManagers   = "managers"
	AudienceTeam       = "team"
	AudienceTechnical  = "technical"
	AudienceStakeholders = "stakeholders"
)