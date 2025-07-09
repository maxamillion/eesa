package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	LogLevel string `yaml:"log_level"`
	
	Jira struct {
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		// Token stored in keyring, not in config file
	} `yaml:"jira"`
	
	Gemini struct {
		// APIKey stored in keyring, not in config file
		Model       string  `yaml:"model"`
		Temperature float32 `yaml:"temperature"`
		MaxTokens   int     `yaml:"max_tokens"`
	} `yaml:"gemini"`
	
	Google struct {
		ClientID string `yaml:"client_id"`
		// ClientSecret stored in keyring, not in config file
	} `yaml:"google"`
	
	Defaults struct {
		TimeRange    string   `yaml:"time_range"`
		Users        []string `yaml:"users"`
		OutputFormat string   `yaml:"output_format"`
	} `yaml:"defaults"`
	
	Security struct {
		TLSMinVersion string `yaml:"tls_min_version"`
		VerifySSL     bool   `yaml:"verify_ssl"`
	} `yaml:"security"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		Gemini: struct {
			Model       string  `yaml:"model"`
			Temperature float32 `yaml:"temperature"`
			MaxTokens   int     `yaml:"max_tokens"`
		}{
			Model:       "gemini-pro",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Defaults: struct {
			TimeRange    string   `yaml:"time_range"`
			Users        []string `yaml:"users"`
			OutputFormat string   `yaml:"output_format"`
		}{
			TimeRange:    "1w",
			Users:        []string{},
			OutputFormat: "google_docs",
		},
		Security: struct {
			TLSMinVersion string `yaml:"tls_min_version"`
			VerifySSL     bool   `yaml:"verify_ssl"`
		}{
			TLSMinVersion: "1.3",
			VerifySSL:     true,
		},
	}
}

// Load loads configuration from file, with environment variable overrides
func Load() (*Config, error) {
	config := DefaultConfig()
	
	// Try to load from config file
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, &ConfigError{
				Code:    "CONFIG_READ_FAILED",
				Message: "Failed to read configuration file",
				Cause:   err,
			}
		}
		
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, &ConfigError{
				Code:    "CONFIG_PARSE_FAILED",
				Message: "Failed to parse configuration file",
				Cause:   err,
			}
		}
	}
	
	// Override with environment variables
	applyEnvOverrides(config)
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()
	
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return &ConfigError{
			Code:    "CONFIG_DIR_CREATE_FAILED",
			Message: "Failed to create configuration directory",
			Cause:   err,
		}
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return &ConfigError{
			Code:    "CONFIG_MARSHAL_FAILED",
			Message: "Failed to marshal configuration",
			Cause:   err,
		}
	}
	
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return &ConfigError{
			Code:    "CONFIG_WRITE_FAILED",
			Message: "Failed to write configuration file",
			Cause:   err,
		}
	}
	
	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Jira.URL == "" {
		return &ConfigError{
			Code:    "JIRA_URL_MISSING",
			Message: "Jira URL is required",
		}
	}
	
	if c.Jira.Username == "" {
		return &ConfigError{
			Code:    "JIRA_USERNAME_MISSING",
			Message: "Jira username is required",
		}
	}
	
	if c.Google.ClientID == "" {
		return &ConfigError{
			Code:    "GOOGLE_CLIENT_ID_MISSING",
			Message: "Google Client ID is required",
		}
	}
	
	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	if path := os.Getenv("ESA_CONFIG_PATH"); path != "" {
		return path
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./config.yaml"
	}
	
	return filepath.Join(homeDir, ".config", "eesa", "config.yaml")
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(config *Config) {
	if logLevel := os.Getenv("ESA_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}
	
	if jiraURL := os.Getenv("ESA_JIRA_URL"); jiraURL != "" {
		config.Jira.URL = jiraURL
	}
	
	if jiraUsername := os.Getenv("ESA_JIRA_USERNAME"); jiraUsername != "" {
		config.Jira.Username = jiraUsername
	}
	
	if geminiModel := os.Getenv("ESA_GEMINI_MODEL"); geminiModel != "" {
		config.Gemini.Model = geminiModel
	}
	
	if googleClientID := os.Getenv("ESA_GOOGLE_CLIENT_ID"); googleClientID != "" {
		config.Google.ClientID = googleClientID
	}
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// ParseTimeRange parses a time range string (e.g., "1w", "2d", "1m")
func ParseTimeRange(s string) (TimeRange, error) {
	now := time.Now()
	
	switch s {
	case "1w":
		return TimeRange{
			Start: now.AddDate(0, 0, -7),
			End:   now,
		}, nil
	case "2w":
		return TimeRange{
			Start: now.AddDate(0, 0, -14),
			End:   now,
		}, nil
	case "1m":
		return TimeRange{
			Start: now.AddDate(0, -1, 0),
			End:   now,
		}, nil
	default:
		return TimeRange{}, &ConfigError{
			Code:    "INVALID_TIME_RANGE",
			Message: "Invalid time range format",
		}
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Code    string
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}