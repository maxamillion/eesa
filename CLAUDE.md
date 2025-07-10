# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

EESA (Efficacious Executive Summary Assistant) is a cross-platform desktop application that automates executive summary creation by:
- Fetching activity data from Jira
- Processing it through Google Gemini AI
- Generating formatted Google Docs summaries

**Architecture**: Go-based GUI application using Fyne framework with layered architecture:
- **Presentation Layer**: Fyne-based GUI (`internal/ui/`)
- **Business Logic**: Core workflows (`internal/processor/`)
- **Data Access**: External API integrations (`internal/jira/`, `internal/gemini/`, `internal/gdocs/`)
- **Infrastructure**: Configuration, logging, security (`internal/config/`, `internal/security/`, `pkg/utils/`)

## Development Commands

### Quick Start (Using Makefile)
```bash
# Show all available commands
make help

# Complete development workflow
make all

# Build application
make build

# Run application
make run

# Run all tests
make test

# Run tests with coverage
make test-coverage-html
```

### Manual Commands (Alternative to Makefile)

#### Building
```bash
# Build for current platform
go build -o eesa ./cmd/eesa

# Cross-compilation
GOOS=darwin GOARCH=amd64 go build -o eesa-darwin ./cmd/eesa
GOOS=windows GOARCH=amd64 go build -o eesa-windows.exe ./cmd/eesa
GOOS=linux GOARCH=amd64 go build -o eesa-linux ./cmd/eesa
```

#### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests excluding integration tests
go test -short ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Running
```bash
# Run application
go run ./cmd/eesa

# Run with environment variables
ESA_CONFIG_PATH=/path/to/config.yaml go run ./cmd/eesa
ESA_LOG_LEVEL=debug go run ./cmd/eesa
```

#### Code Quality
```bash
# Format and vet
go fmt ./...
go vet ./...

# Dependencies
go mod tidy
go mod download
```

## Key Architecture Patterns

### Dependency Injection
All external dependencies are injected through interfaces, enabling easy testing and mocking.

### Interface-Based Design
- `JiraClientInterface`: Jira API operations
- `GeminiClientInterface`: AI summarization  
- `DocsClientInterface`: Google Docs creation
- `Logger`: Structured logging interface

### Error Handling
Custom `AppError` type with error codes for consistent error handling across the application.

### Configuration Management
YAML configuration with environment variable overrides. Sensitive data (API keys, tokens) stored in OS keyring via `internal/security/keyring.go`.

## Testing Strategy

- **Unit Tests**: 90%+ coverage target with testify framework
- **Integration Tests**: Real API interactions in `internal/mocks/integration_test.go`
- **Mock Services**: Complete mock implementations for all external services
- **TDD Approach**: Write tests first, then implement features

Key test files:
- `internal/config/config_test.go`
- `internal/processor/engine_test.go`
- `internal/jira/client_test.go`
- `internal/gemini/client_test.go`
- `pkg/validation/validator_test.go`

## Security Implementation

- **Credential Storage**: OS keyring integration (no plaintext secrets)
- **API Security**: TLS 1.3, rate limiting, request validation
- **Memory Safety**: Sensitive data cleared from memory after use

## Environment Variables

- `ESA_CONFIG_PATH`: Configuration file path
- `ESA_LOG_LEVEL`: Logging level (debug, info, warn, error)
- `ESA_JIRA_URL`: Jira instance URL
- `ESA_JIRA_USERNAME`: Jira username
- `ESA_GEMINI_MODEL`: Gemini AI model name
- `ESA_GOOGLE_CLIENT_ID`: Google OAuth client ID

## External Dependencies

- **Jira REST API**: Activity data retrieval
- **Google Gemini API**: AI summarization
- **Google Docs API**: Document creation
- **OS Keyring**: Secure credential storage
- **Fyne Framework**: Cross-platform GUI

## Development Notes

- Project uses Go 1.23.9 with Fyne v2.6.1
- No CI/CD configuration files present
- Comprehensive Makefile available with all development tasks
- Entry point: `cmd/eesa/main.go`
- Comprehensive PRD and implementation plan available in `PRD.md` and `IMPLEMENTATION_PLAN.md`