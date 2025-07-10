package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/company/eesa/internal/gdocs"
	"github.com/company/eesa/pkg/utils"
)

// MockGoogleDocsClient is a mock implementation of the Google Docs client for testing
type MockGoogleDocsClient struct {
	// Mock configuration
	logger utils.Logger
	
	// Mock behavior controls
	shouldFailAuth       bool
	shouldFailCreate     bool
	shouldFailUpdate     bool
	shouldFailShare      bool
	shouldFailGet        bool
	shouldFailValidate   bool
	simulateTimeout      bool
	simulateRateLimit    bool
	simulateQuotaExceeded bool
	
	// Mock data storage
	documents       map[string]*gdocs.DocumentResponse
	sharedWith      map[string][]string // documentID -> list of emails
	documentCounter int
	
	// Call tracking
	mu                sync.RWMutex
	createCalls       int
	updateCalls       int
	shareCalls        int
	getCalls          int
	validateCalls     int
	execSummaryCalls  int
	lastTitle         string
	lastDocumentID    string
	lastEmails        []string
	lastContent       string
	
	// Response customization
	responseDelay     time.Duration
}

// NewMockGoogleDocsClient creates a new mock Google Docs client
func NewMockGoogleDocsClient(logger utils.Logger) *MockGoogleDocsClient {
	return &MockGoogleDocsClient{
		logger:      logger,
		documents:   make(map[string]*gdocs.DocumentResponse),
		sharedWith:  make(map[string][]string),
		responseDelay: 0,
	}
}

// Configuration methods
func (m *MockGoogleDocsClient) SetFailAuth(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailAuth = fail
}

func (m *MockGoogleDocsClient) SetFailCreate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailCreate = fail
}

func (m *MockGoogleDocsClient) SetFailUpdate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailUpdate = fail
}

func (m *MockGoogleDocsClient) SetFailShare(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailShare = fail
}

func (m *MockGoogleDocsClient) SetFailGet(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailGet = fail
}

func (m *MockGoogleDocsClient) SetFailValidate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailValidate = fail
}

func (m *MockGoogleDocsClient) SetSimulateTimeout(timeout bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateTimeout = timeout
}

func (m *MockGoogleDocsClient) SetSimulateRateLimit(rateLimit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateRateLimit = rateLimit
}

func (m *MockGoogleDocsClient) SetSimulateQuotaExceeded(quotaExceeded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulateQuotaExceeded = quotaExceeded
}

func (m *MockGoogleDocsClient) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// Tracking methods
func (m *MockGoogleDocsClient) GetCallCounts() (create, update, share, get, validate, execSummary int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.createCalls, m.updateCalls, m.shareCalls, m.getCalls, m.validateCalls, m.execSummaryCalls
}

func (m *MockGoogleDocsClient) GetLastTitle() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastTitle
}

func (m *MockGoogleDocsClient) GetLastDocumentID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastDocumentID
}

func (m *MockGoogleDocsClient) GetLastEmails() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastEmails
}

func (m *MockGoogleDocsClient) GetLastContent() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastContent
}

func (m *MockGoogleDocsClient) GetStoredDocuments() map[string]*gdocs.DocumentResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*gdocs.DocumentResponse)
	for id, doc := range m.documents {
		result[id] = doc
	}
	return result
}

func (m *MockGoogleDocsClient) GetSharedWith() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string][]string)
	for id, emails := range m.sharedWith {
		result[id] = make([]string, len(emails))
		copy(result[id], emails)
	}
	return result
}

func (m *MockGoogleDocsClient) ResetCallCounts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalls = 0
	m.updateCalls = 0
	m.shareCalls = 0
	m.getCalls = 0
	m.validateCalls = 0
	m.execSummaryCalls = 0
	m.lastTitle = ""
	m.lastDocumentID = ""
	m.lastEmails = nil
	m.lastContent = ""
}

func (m *MockGoogleDocsClient) ClearStoredData() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.documents = make(map[string]*gdocs.DocumentResponse)
	m.sharedWith = make(map[string][]string)
	m.documentCounter = 0
}

// Helper method for common behavior simulation
func (m *MockGoogleDocsClient) simulateCommonBehavior(ctx context.Context) error {
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
func (m *MockGoogleDocsClient) CreateDocument(ctx context.Context, title string, content string) (*gdocs.DocumentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.createCalls++
	m.lastTitle = title
	m.lastContent = content
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailCreate {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Mock document creation failed", nil)
	}
	
	// Generate a new document ID
	m.documentCounter++
	documentID := fmt.Sprintf("mock-doc-%d", m.documentCounter)
	m.lastDocumentID = documentID
	
	// Create the document response
	document := &gdocs.DocumentResponse{
		DocumentID: documentID,
		Title:      title,
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					Paragraph: &gdocs.Paragraph{
						Elements: []*gdocs.ParagraphElement{
							{
								TextRun: &gdocs.TextRun{
									Content: title + "\n\n" + content,
								},
							},
						},
					},
				},
			},
		},
		RevisionID: fmt.Sprintf("rev-%d", m.documentCounter),
	}
	
	// Store the document
	m.documents[documentID] = document
	
	m.logger.Info("Mock Google Docs document created",
		utils.NewField("document_id", documentID),
		utils.NewField("title", title),
		utils.NewField("content_length", len(content)),
	)
	
	return document, nil
}

func (m *MockGoogleDocsClient) UpdateDocument(ctx context.Context, documentID string, requests []gdocs.Request) (*gdocs.BatchUpdateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.updateCalls++
	m.lastDocumentID = documentID
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailUpdate {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Mock document update failed", nil)
	}
	
	// Check if document exists
	document, exists := m.documents[documentID]
	if !exists {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Document not found", nil)
	}
	
	// Process update requests (simplified implementation)
	replies := make([]*gdocs.Response, len(requests))
	for i, request := range requests {
		if request.InsertText != nil {
			// Add text content to the document
			if document.Body == nil {
				document.Body = &gdocs.Body{
					Content: []*gdocs.StructuralElement{},
				}
			}
			
			// Create a new paragraph with the inserted text
			newParagraph := &gdocs.StructuralElement{
				Paragraph: &gdocs.Paragraph{
					Elements: []*gdocs.ParagraphElement{
						{
							TextRun: &gdocs.TextRun{
								Content: request.InsertText.Text,
							},
						},
					},
				},
			}
			
			document.Body.Content = append(document.Body.Content, newParagraph)
			m.lastContent += request.InsertText.Text
			
			replies[i] = &gdocs.Response{
				InsertText: &gdocs.InsertTextResponse{},
			}
		}
	}
	
	response := &gdocs.BatchUpdateResponse{
		DocumentID: documentID,
		Replies:    replies,
	}
	
	m.logger.Info("Mock Google Docs document updated",
		utils.NewField("document_id", documentID),
		utils.NewField("request_count", len(requests)),
	)
	
	return response, nil
}

func (m *MockGoogleDocsClient) GetDocument(ctx context.Context, documentID string) (*gdocs.DocumentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.getCalls++
	m.lastDocumentID = documentID
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailGet {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Mock document retrieval failed", nil)
	}
	
	// Check if document exists
	document, exists := m.documents[documentID]
	if !exists {
		return nil, utils.NewAppError(utils.ErrorCodeGoogleError, "Document not found", nil)
	}
	
	m.logger.Info("Mock Google Docs document retrieved",
		utils.NewField("document_id", documentID),
		utils.NewField("title", document.Title),
	)
	
	return document, nil
}

func (m *MockGoogleDocsClient) ShareDocument(ctx context.Context, documentID string, emails []string, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.shareCalls++
	m.lastDocumentID = documentID
	m.lastEmails = emails
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return err
	}
	
	if m.shouldFailShare {
		return utils.NewAppError(utils.ErrorCodeGoogleError, "Mock document sharing failed", nil)
	}
	
	// Check if document exists
	if _, exists := m.documents[documentID]; !exists {
		return utils.NewAppError(utils.ErrorCodeGoogleError, "Document not found", nil)
	}
	
	// Store sharing information
	if m.sharedWith[documentID] == nil {
		m.sharedWith[documentID] = []string{}
	}
	m.sharedWith[documentID] = append(m.sharedWith[documentID], emails...)
	
	m.logger.Info("Mock Google Docs document shared",
		utils.NewField("document_id", documentID),
		utils.NewField("emails", emails),
		utils.NewField("role", role),
	)
	
	return nil
}

func (m *MockGoogleDocsClient) ValidateCredentials(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.validateCalls++
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return err
	}
	
	if m.shouldFailValidate {
		return utils.NewAppError(utils.ErrorCodeAuthFailed, "Mock credential validation failed", nil)
	}
	
	m.logger.Info("Mock Google Docs credential validation successful")
	return nil
}

func (m *MockGoogleDocsClient) CreateExecutiveSummaryDocument(ctx context.Context, title, summary string, metadata map[string]interface{}) (*gdocs.DocumentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.execSummaryCalls++
	m.lastTitle = title
	m.lastContent = summary
	
	if err := m.simulateCommonBehavior(ctx); err != nil {
		return nil, err
	}
	
	if m.shouldFailCreate {
		return utils.NewAppError(utils.ErrorCodeGoogleError, "Mock executive summary document creation failed", nil)
	}
	
	// Generate a new document ID
	m.documentCounter++
	documentID := fmt.Sprintf("mock-exec-summary-%d", m.documentCounter)
	m.lastDocumentID = documentID
	
	// Create formatted content with metadata
	formattedContent := fmt.Sprintf("# %s\n\n%s\n\n---\n\nGenerated on: %s\n", 
		title, summary, time.Now().Format("2006-01-02 15:04:05"))
	
	// Add metadata if provided
	if metadata != nil {
		formattedContent += "\n## Metadata\n"
		for key, value := range metadata {
			formattedContent += fmt.Sprintf("- **%s**: %v\n", key, value)
		}
	}
	
	// Create the document with formatted content
	document := &gdocs.DocumentResponse{
		DocumentID: documentID,
		Title:      title,
		Body: &gdocs.Body{
			Content: []*gdocs.StructuralElement{
				{
					Paragraph: &gdocs.Paragraph{
						Elements: []*gdocs.ParagraphElement{
							{
								TextRun: &gdocs.TextRun{
									Content: formattedContent,
								},
							},
						},
					},
				},
			},
		},
		RevisionID: fmt.Sprintf("rev-%d", m.documentCounter),
	}
	
	// Store the document
	m.documents[documentID] = document
	
	m.logger.Info("Mock Google Docs executive summary document created",
		utils.NewField("document_id", documentID),
		utils.NewField("title", title),
		utils.NewField("summary_length", len(summary)),
		utils.NewField("metadata_count", len(metadata)),
	)
	
	return document, nil
}

// Helper methods
func (m *MockGoogleDocsClient) GetDocumentURL(documentID string) string {
	return fmt.Sprintf("https://docs.google.com/document/d/%s/edit", documentID)
}

func (m *MockGoogleDocsClient) ExtractPlainText(document *gdocs.DocumentResponse) string {
	if document == nil || document.Body == nil {
		return ""
	}
	
	var text strings.Builder
	for _, element := range document.Body.Content {
		if element.Paragraph != nil {
			for _, paragraphElement := range element.Paragraph.Elements {
				if paragraphElement.TextRun != nil {
					text.WriteString(paragraphElement.TextRun.Content)
				}
			}
		}
	}
	
	return text.String()
}

// Verify that MockGoogleDocsClient implements the GoogleDocsClientInterface
var _ gdocs.GoogleDocsClientInterface = (*MockGoogleDocsClient)(nil)