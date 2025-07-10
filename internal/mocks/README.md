# Mock Services for ESA (Executive Summary Automation)

This directory contains comprehensive mock implementations for all external services used by the ESA application. These mocks are designed for testing purposes and provide realistic behavior simulation without requiring actual API connections.

## Overview

The mock services implement the same interfaces as the real services but provide controlled, predictable behavior for testing. Each mock service supports:

- **Configurable Failure Modes**: Simulate authentication failures, rate limiting, timeouts, and other error conditions
- **Call Tracking**: Track method calls, parameters, and usage patterns
- **Response Customization**: Set custom responses for specific test scenarios
- **Behavior Simulation**: Simulate delays, timeouts, and other real-world conditions
- **Data Storage**: Store and retrieve test data for integration testing

## Mock Services

### 1. Jira Mock Service (`internal/mocks/jira/`)

**File**: `client.go`

**Features**:
- Implements `JiraClientInterface`
- Simulates all Jira API operations (search, get issue, get user, etc.)
- Provides realistic mock data for issues, users, worklogs, and comments
- Configurable failure scenarios (auth, rate limiting, timeouts)
- Call counting and parameter tracking

**Usage Example**:
```go
jiraClient := jira.NewMockJiraClient("https://test.atlassian.net", "user", "token", logger)

// Configure behavior
jiraClient.SetFailAuth(false)
jiraClient.SetSimulateRateLimit(false)

// Use like real client
searchResult, err := jiraClient.SearchIssues(ctx, "project = TEST", 10)

// Track calls
auth, search, issue, user := jiraClient.GetCallCounts()
lastQuery := jiraClient.GetLastSearchQuery()
```

### 2. Gemini Mock Service (`internal/mocks/gemini/`)

**File**: `client.go`

**Features**:
- Implements `GeminiClientInterface`
- Generates contextual responses based on prompt content
- Supports custom response mapping for specific prompts
- Simulates token usage and metadata
- Configurable failure scenarios (auth, rate limiting, quota exceeded)

**Usage Example**:
```go
geminiClient := gemini.NewMockGeminiClient("test-api-key", logger)

// Set custom response
geminiClient.SetCustomResponse("executive summary", "Custom executive summary content")

// Configure behavior
geminiClient.SetSimulateRateLimit(false)
geminiClient.SetResponseDelay(10 * time.Millisecond)

// Use like real client
response, err := geminiClient.GenerateContent(ctx, "gemini-pro", "executive summary")

// Track calls
generate, models := geminiClient.GetCallCounts()
lastPrompt := geminiClient.GetLastPrompt()
```

### 3. Google Docs Mock Service (`internal/mocks/gdocs/`)

**File**: `client.go`

**Features**:
- Implements `GoogleDocsClientInterface`
- Simulates document creation, updating, sharing, and retrieval
- Stores documents and permissions in memory
- Generates realistic document IDs and URLs
- Supports executive summary document creation
- Configurable failure scenarios

**Usage Example**:
```go
gdocsClient := gdocs.NewMockGoogleDocsClient(logger)

// Configure behavior
gdocsClient.SetFailCreate(false)
gdocsClient.SetResponseDelay(5 * time.Millisecond)

// Use like real client
docResponse, err := gdocsClient.CreateDocument(ctx, "Test Document")
permission, err := gdocsClient.ShareDocument(ctx, docResponse.DocumentID, "test@example.com", "reader")

// Track calls and data
create, update, share, get := gdocsClient.GetCallCounts()
storedDocs := gdocsClient.GetStoredDocuments()
storedPerms := gdocsClient.GetStoredPermissions()
```

## Integration Testing

The `integration_test.go` file demonstrates how to use all mock services together in realistic test scenarios:

### Test Scenarios Covered:

1. **Individual Service Testing**: Test each mock service independently
2. **Cross-Service Integration**: Simulate full workflows (Jira → Gemini → Google Docs)
3. **Failure Scenario Testing**: Test error handling and recovery
4. **Timeout and Context Handling**: Test cancellation and timeout behavior
5. **Response Delay Testing**: Test performance under various conditions

### Running Integration Tests:

```bash
# Run all mock service tests
go test ./internal/mocks/

# Run with verbose output
go test -v ./internal/mocks/

# Run specific test
go test -v ./internal/mocks/ -run TestMockServicesIntegration
```

## Configuration Options

All mock services support the following configuration options:

### Failure Simulation:
- `SetFailAuth(bool)`: Simulate authentication failures
- `SetFailXXX(bool)`: Simulate specific operation failures
- `SetSimulateTimeout(bool)`: Simulate timeout conditions
- `SetSimulateRateLimit(bool)`: Simulate rate limiting
- `SetSimulateQuotaExceeded(bool)`: Simulate quota exceeded (where applicable)

### Behavior Customization:
- `SetResponseDelay(time.Duration)`: Add delays to responses
- `SetCustomResponse(string, string)`: Set custom responses for specific inputs
- `SetMockXXXData(interface{})`: Set custom mock data

### Call Tracking:
- `GetCallCounts()`: Get number of calls to each method
- `GetLastXXX()`: Get last parameters passed to methods
- `ResetCallCounts()`: Reset all call counters

### Data Management:
- `GetStoredXXX()`: Get stored data (documents, permissions, etc.)
- `ClearStoredData()`: Clear all stored data

## Best Practices

### 1. Test Setup:
```go
func TestMyFeature(t *testing.T) {
    logger := utils.NewMockLogger()
    ctx := context.Background()
    
    // Initialize mocks
    jiraClient := jira.NewMockJiraClient("https://test.atlassian.net", "user", "token", logger)
    geminiClient := gemini.NewMockGeminiClient("test-api-key", logger)
    gdocsClient := gdocs.NewMockGoogleDocsClient(logger)
    
    // Configure as needed
    jiraClient.SetFailAuth(false)
    geminiClient.SetResponseDelay(10 * time.Millisecond)
    
    // Your test logic here
}
```

### 2. Failure Testing:
```go
func TestErrorHandling(t *testing.T) {
    client := jira.NewMockJiraClient("https://test.atlassian.net", "user", "token", logger)
    
    // Test authentication failure
    client.SetFailAuth(true)
    err := client.TestConnection(ctx)
    assert.Error(t, err)
    
    // Test rate limiting
    client.SetFailAuth(false)
    client.SetSimulateRateLimit(true)
    _, err = client.SearchIssues(ctx, "project = TEST", 10)
    assert.Error(t, err)
}
```

### 3. Integration Testing:
```go
func TestFullWorkflow(t *testing.T) {
    // 1. Get data from Jira
    searchResult, err := jiraClient.SearchIssues(ctx, "project = TEST", 10)
    require.NoError(t, err)
    
    // 2. Generate summary with Gemini
    response, err := geminiClient.GenerateContent(ctx, "gemini-pro", "Generate summary")
    require.NoError(t, err)
    
    // 3. Create document with Google Docs
    docResponse, err := gdocsClient.CreateExecutiveSummaryDocument(ctx, "Summary", response.Candidates[0].Content.Parts[0].Text)
    require.NoError(t, err)
    
    // 4. Verify workflow
    assert.NotEmpty(t, docResponse.DocumentID)
}
```

### 4. Cleanup:
```go
func TestWithCleanup(t *testing.T) {
    client := gdocs.NewMockGoogleDocsClient(logger)
    
    // Your test logic here
    
    // Cleanup
    client.ResetCallCounts()
    client.ClearStoredData()
}
```

## Mock Data

The mock services come with realistic default data:

### Jira Mock Data:
- 2 sample issues (TEST-1, TEST-2)
- User data (user123, "Test User")
- Worklogs and comments
- Various issue types and statuses

### Gemini Mock Data:
- 2 sample models (gemini-pro, gemini-pro-vision)
- Contextual response generation
- Token usage metadata

### Google Docs Mock Data:
- Dynamic document creation
- Permission management
- Document content storage
- URL generation

## Contributing

When adding new mock functionality:

1. **Maintain Interface Compatibility**: Ensure mocks implement the same interfaces as real services
2. **Add Comprehensive Tests**: Include tests for both success and failure scenarios
3. **Document Configuration Options**: Update this README with new configuration options
4. **Follow Existing Patterns**: Use the same patterns for call tracking, failure simulation, etc.
5. **Update Integration Tests**: Add new functionality to the integration test suite

## Troubleshooting

### Common Issues:

1. **Interface Mismatch**: Ensure mock implements all required interface methods
2. **Race Conditions**: Use proper mutex locking when accessing shared data
3. **Context Handling**: Always respect context cancellation and timeouts
4. **Memory Management**: Clear stored data between tests to avoid interference

### Debugging Tips:

1. Use `GetCallCounts()` to verify mock methods are being called
2. Check `GetLastXXX()` methods to verify correct parameters
3. Enable verbose logging in tests to see mock behavior
4. Use `SetResponseDelay()` to test timing-sensitive code