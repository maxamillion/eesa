package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "gemini-pro", config.Gemini.Model)
	assert.Equal(t, float32(0.7), config.Gemini.Temperature)
	assert.Equal(t, 4096, config.Gemini.MaxTokens)
	assert.Equal(t, "1w", config.Defaults.TimeRange)
	assert.Equal(t, "google_docs", config.Defaults.OutputFormat)
	assert.Equal(t, "1.3", config.Security.TLSMinVersion)
	assert.True(t, config.Security.VerifySSL)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errCode string
	}{
		{
			name: "valid config",
			config: &Config{
				Jira: struct {
					URL      string `yaml:"url"`
					Username string `yaml:"username"`
				}{
					URL:      "https://company.atlassian.net",
					Username: "testuser",
				},
				Google: struct {
					ClientID string `yaml:"client_id"`
				}{
					ClientID: "test-client-id",
				},
			},
			wantErr: false,
		},
		{
			name: "missing jira url",
			config: &Config{
				Jira: struct {
					URL      string `yaml:"url"`
					Username string `yaml:"username"`
				}{
					Username: "testuser",
				},
				Google: struct {
					ClientID string `yaml:"client_id"`
				}{
					ClientID: "test-client-id",
				},
			},
			wantErr: true,
			errCode: "JIRA_URL_MISSING",
		},
		{
			name: "missing jira username",
			config: &Config{
				Jira: struct {
					URL      string `yaml:"url"`
					Username string `yaml:"username"`
				}{
					URL: "https://company.atlassian.net",
				},
				Google: struct {
					ClientID string `yaml:"client_id"`
				}{
					ClientID: "test-client-id",
				},
			},
			wantErr: true,
			errCode: "JIRA_USERNAME_MISSING",
		},
		{
			name: "missing google client id",
			config: &Config{
				Jira: struct {
					URL      string `yaml:"url"`
					Username string `yaml:"username"`
				}{
					URL:      "https://company.atlassian.net",
					Username: "testuser",
				},
			},
			wantErr: true,
			errCode: "GOOGLE_CLIENT_ID_MISSING",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			
			if tt.wantErr {
				require.Error(t, err)
				configErr, ok := err.(*ConfigError)
				require.True(t, ok, "Expected ConfigError")
				assert.Equal(t, tt.errCode, configErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	
	// Set environment variable to use temp path
	oldPath := os.Getenv("ESA_CONFIG_PATH")
	os.Setenv("ESA_CONFIG_PATH", configPath)
	defer func() {
		if oldPath != "" {
			os.Setenv("ESA_CONFIG_PATH", oldPath)
		} else {
			os.Unsetenv("ESA_CONFIG_PATH")
		}
	}()
	
	// Create test config
	originalConfig := DefaultConfig()
	originalConfig.LogLevel = "debug"
	originalConfig.Jira.URL = "https://test.atlassian.net"
	originalConfig.Jira.Username = "testuser"
	originalConfig.Google.ClientID = "test-client-id"
	
	// Save config
	err := originalConfig.Save()
	require.NoError(t, err)
	
	// Load config
	loadedConfig, err := Load()
	require.NoError(t, err)
	
	// Compare configs
	assert.Equal(t, originalConfig.LogLevel, loadedConfig.LogLevel)
	assert.Equal(t, originalConfig.Jira.URL, loadedConfig.Jira.URL)
	assert.Equal(t, originalConfig.Jira.Username, loadedConfig.Jira.Username)
	assert.Equal(t, originalConfig.Google.ClientID, loadedConfig.Google.ClientID)
}

func TestParseTimeRange(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected TimeRange
	}{
		{
			name:    "one week",
			input:   "1w",
			wantErr: false,
			expected: TimeRange{
				Start: now.AddDate(0, 0, -7),
				End:   now,
			},
		},
		{
			name:    "two weeks",
			input:   "2w",
			wantErr: false,
			expected: TimeRange{
				Start: now.AddDate(0, 0, -14),
				End:   now,
			},
		},
		{
			name:    "one month",
			input:   "1m",
			wantErr: false,
			expected: TimeRange{
				Start: now.AddDate(0, -1, 0),
				End:   now,
			},
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeRange(tt.input)
			
			if tt.wantErr {
				require.Error(t, err)
				configErr, ok := err.(*ConfigError)
				require.True(t, ok, "Expected ConfigError")
				assert.Equal(t, "INVALID_TIME_RANGE", configErr.Code)
			} else {
				require.NoError(t, err)
				// Allow for small time differences due to test execution time
				assert.WithinDuration(t, tt.expected.Start, result.Start, time.Second)
				assert.WithinDuration(t, tt.expected.End, result.End, time.Second)
			}
		})
	}
}

func TestEnvOverrides(t *testing.T) {
	// Set test environment variables
	testEnvs := map[string]string{
		"ESA_LOG_LEVEL":        "debug",
		"ESA_JIRA_URL":         "https://env.atlassian.net",
		"ESA_JIRA_USERNAME":    "envuser",
		"ESA_GEMINI_MODEL":     "gemini-pro-vision",
		"ESA_GOOGLE_CLIENT_ID": "env-client-id",
	}
	
	// Set environment variables
	for key, value := range testEnvs {
		oldValue := os.Getenv(key)
		os.Setenv(key, value)
		defer func(k, v string) {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}(key, oldValue)
	}
	
	// Create config and apply overrides
	config := DefaultConfig()
	applyEnvOverrides(config)
	
	// Check overrides
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "https://env.atlassian.net", config.Jira.URL)
	assert.Equal(t, "envuser", config.Jira.Username)
	assert.Equal(t, "gemini-pro-vision", config.Gemini.Model)
	assert.Equal(t, "env-client-id", config.Google.ClientID)
}

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name: "error with cause",
			err: &ConfigError{
				Code:    "TEST_ERROR",
				Message: "Test error message",
				Cause:   assert.AnError,
			},
			expected: "Test error message: " + assert.AnError.Error(),
		},
		{
			name: "error without cause",
			err: &ConfigError{
				Code:    "TEST_ERROR",
				Message: "Test error message",
			},
			expected: "Test error message",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}