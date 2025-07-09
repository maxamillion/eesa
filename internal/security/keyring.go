package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"runtime"

	"github.com/company/eesa/pkg/utils"
	"github.com/zalando/go-keyring"
)

const (
	// Service name for keyring entries
	ServiceName = "eesa"
	
	// Key names for different credentials
	KeyJiraToken        = "jira_token"
	KeyGeminiAPIKey     = "gemini_api_key"
	KeyGoogleClientSecret = "google_client_secret"
	KeyGoogleAccessToken  = "google_access_token"
	KeyGoogleRefreshToken = "google_refresh_token"
	KeyEncryptionKey    = "encryption_key"
)

// KeyringManager handles secure storage and retrieval of credentials
type KeyringManager struct {
	logger utils.Logger
}

// NewKeyringManager creates a new keyring manager
func NewKeyringManager(logger utils.Logger) *KeyringManager {
	return &KeyringManager{
		logger: logger,
	}
}

// StoreCredential stores a credential in the OS keyring
func (k *KeyringManager) StoreCredential(key, value string) error {
	if key == "" {
		return utils.NewAppError(utils.ErrorCodeKeyringError, "Key cannot be empty", nil)
	}
	
	if value == "" {
		return utils.NewAppError(utils.ErrorCodeKeyringError, "Value cannot be empty", nil)
	}
	
	// Use platform-specific service name
	serviceName := k.getServiceName()
	
	err := keyring.Set(serviceName, key, value)
	if err != nil {
		k.logger.Error("Failed to store credential", err,
			utils.NewField("key", key),
			utils.NewField("service", serviceName),
			utils.NewField("platform", runtime.GOOS),
		)
		return utils.NewAppError(utils.ErrorCodeKeyringError, "Failed to store credential", err).
			WithService("keyring").
			WithExtra("key", key)
	}
	
	k.logger.Info("Credential stored successfully",
		utils.NewField("key", key),
		utils.NewField("service", serviceName),
		utils.NewField("platform", runtime.GOOS),
	)
	
	return nil
}

// GetCredential retrieves a credential from the OS keyring
func (k *KeyringManager) GetCredential(key string) (string, error) {
	if key == "" {
		return "", utils.NewAppError(utils.ErrorCodeKeyringError, "Key cannot be empty", nil)
	}
	
	serviceName := k.getServiceName()
	
	value, err := keyring.Get(serviceName, key)
	if err != nil {
		if err == keyring.ErrNotFound {
			k.logger.Debug("Credential not found",
				utils.NewField("key", key),
				utils.NewField("service", serviceName),
			)
			return "", utils.NewAppError(utils.ErrorCodeCredentialsMissing, "Credential not found", err).
				WithService("keyring").
				WithExtra("key", key)
		}
		
		k.logger.Error("Failed to retrieve credential", err,
			utils.NewField("key", key),
			utils.NewField("service", serviceName),
			utils.NewField("platform", runtime.GOOS),
		)
		return "", utils.NewAppError(utils.ErrorCodeKeyringError, "Failed to retrieve credential", err).
			WithService("keyring").
			WithExtra("key", key)
	}
	
	k.logger.Debug("Credential retrieved successfully",
		utils.NewField("key", key),
		utils.NewField("service", serviceName),
	)
	
	return value, nil
}

// DeleteCredential removes a credential from the OS keyring
func (k *KeyringManager) DeleteCredential(key string) error {
	if key == "" {
		return utils.NewAppError(utils.ErrorCodeKeyringError, "Key cannot be empty", nil)
	}
	
	serviceName := k.getServiceName()
	
	err := keyring.Delete(serviceName, key)
	if err != nil {
		if err == keyring.ErrNotFound {
			k.logger.Debug("Credential not found for deletion",
				utils.NewField("key", key),
				utils.NewField("service", serviceName),
			)
			return utils.NewAppError(utils.ErrorCodeCredentialsMissing, "Credential not found", err).
				WithService("keyring").
				WithExtra("key", key)
		}
		
		k.logger.Error("Failed to delete credential", err,
			utils.NewField("key", key),
			utils.NewField("service", serviceName),
			utils.NewField("platform", runtime.GOOS),
		)
		return utils.NewAppError(utils.ErrorCodeKeyringError, "Failed to delete credential", err).
			WithService("keyring").
			WithExtra("key", key)
	}
	
	k.logger.Info("Credential deleted successfully",
		utils.NewField("key", key),
		utils.NewField("service", serviceName),
	)
	
	return nil
}

// HasCredential checks if a credential exists in the keyring
func (k *KeyringManager) HasCredential(key string) bool {
	_, err := k.GetCredential(key)
	return err == nil
}

// ListCredentials returns a list of all stored credential keys
func (k *KeyringManager) ListCredentials() ([]string, error) {
	// Note: The go-keyring library doesn't provide a native list function
	// We check for known keys instead
	knownKeys := []string{
		KeyJiraToken,
		KeyGeminiAPIKey,
		KeyGoogleClientSecret,
		KeyGoogleAccessToken,
		KeyGoogleRefreshToken,
		KeyEncryptionKey,
	}
	
	var existingKeys []string
	for _, key := range knownKeys {
		if k.HasCredential(key) {
			existingKeys = append(existingKeys, key)
		}
	}
	
	return existingKeys, nil
}

// GenerateEncryptionKey generates a new encryption key and stores it
func (k *KeyringManager) GenerateEncryptionKey() error {
	// Generate a 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return utils.NewAppError(utils.ErrorCodeEncryptionError, "Failed to generate encryption key", err)
	}
	
	// Encode to base64 for storage
	encodedKey := base64.StdEncoding.EncodeToString(key)
	
	// Store in keyring
	err := k.StoreCredential(KeyEncryptionKey, encodedKey)
	if err != nil {
		return utils.WrapError(err, utils.ErrorCodeEncryptionError, "Failed to store encryption key")
	}
	
	k.logger.Info("Encryption key generated and stored")
	
	return nil
}

// GetOrGenerateEncryptionKey gets the encryption key or generates one if it doesn't exist
func (k *KeyringManager) GetOrGenerateEncryptionKey() ([]byte, error) {
	encodedKey, err := k.GetCredential(KeyEncryptionKey)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == utils.ErrorCodeCredentialsMissing {
			// Generate new key
			if genErr := k.GenerateEncryptionKey(); genErr != nil {
				return nil, genErr
			}
			
			// Retrieve the newly generated key
			encodedKey, err = k.GetCredential(KeyEncryptionKey)
			if err != nil {
				return nil, utils.WrapError(err, utils.ErrorCodeEncryptionError, "Failed to retrieve newly generated encryption key")
			}
		} else {
			return nil, err
		}
	}
	
	// Decode from base64
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, utils.NewAppError(utils.ErrorCodeEncryptionError, "Failed to decode encryption key", err)
	}
	
	return key, nil
}

// ValidateCredential validates that a credential exists and is not empty
func (k *KeyringManager) ValidateCredential(key string) error {
	value, err := k.GetCredential(key)
	if err != nil {
		return err
	}
	
	if value == "" {
		return utils.NewAppError(utils.ErrorCodeCredentialsMissing, "Credential is empty", nil).
			WithService("keyring").
			WithExtra("key", key)
	}
	
	return nil
}

// SecureCompare performs constant-time comparison of two strings
func (k *KeyringManager) SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ClearSensitiveData securely clears sensitive data from memory
func (k *KeyringManager) ClearSensitiveData(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// getServiceName returns the platform-specific service name
func (k *KeyringManager) getServiceName() string {
	return ServiceName
}

// CredentialStore provides a high-level interface for credential management
type CredentialStore struct {
	keyring *KeyringManager
	logger  utils.Logger
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(logger utils.Logger) *CredentialStore {
	return &CredentialStore{
		keyring: NewKeyringManager(logger),
		logger:  logger,
	}
}

// JiraCredentials represents Jira authentication credentials
type JiraCredentials struct {
	Token string
}

// GeminiCredentials represents Gemini API credentials
type GeminiCredentials struct {
	APIKey string
}

// GoogleCredentials represents Google API credentials
type GoogleCredentials struct {
	ClientSecret string
	AccessToken  string
	RefreshToken string
}

// SetJiraCredentials stores Jira credentials
func (c *CredentialStore) SetJiraCredentials(creds JiraCredentials) error {
	if creds.Token == "" {
		return utils.NewAppError(utils.ErrorCodeValidationError, "Jira token cannot be empty", nil)
	}
	
	return c.keyring.StoreCredential(KeyJiraToken, creds.Token)
}

// GetJiraCredentials retrieves Jira credentials
func (c *CredentialStore) GetJiraCredentials() (JiraCredentials, error) {
	token, err := c.keyring.GetCredential(KeyJiraToken)
	if err != nil {
		return JiraCredentials{}, err
	}
	
	return JiraCredentials{Token: token}, nil
}

// SetGeminiCredentials stores Gemini API credentials
func (c *CredentialStore) SetGeminiCredentials(creds GeminiCredentials) error {
	if creds.APIKey == "" {
		return utils.NewAppError(utils.ErrorCodeValidationError, "Gemini API key cannot be empty", nil)
	}
	
	return c.keyring.StoreCredential(KeyGeminiAPIKey, creds.APIKey)
}

// GetGeminiCredentials retrieves Gemini API credentials
func (c *CredentialStore) GetGeminiCredentials() (GeminiCredentials, error) {
	apiKey, err := c.keyring.GetCredential(KeyGeminiAPIKey)
	if err != nil {
		return GeminiCredentials{}, err
	}
	
	return GeminiCredentials{APIKey: apiKey}, nil
}

// SetGoogleCredentials stores Google API credentials
func (c *CredentialStore) SetGoogleCredentials(creds GoogleCredentials) error {
	if creds.ClientSecret == "" {
		return utils.NewAppError(utils.ErrorCodeValidationError, "Google client secret cannot be empty", nil)
	}
	
	// Store all credentials
	if err := c.keyring.StoreCredential(KeyGoogleClientSecret, creds.ClientSecret); err != nil {
		return err
	}
	
	if creds.AccessToken != "" {
		if err := c.keyring.StoreCredential(KeyGoogleAccessToken, creds.AccessToken); err != nil {
			return err
		}
	}
	
	if creds.RefreshToken != "" {
		if err := c.keyring.StoreCredential(KeyGoogleRefreshToken, creds.RefreshToken); err != nil {
			return err
		}
	}
	
	return nil
}

// GetGoogleCredentials retrieves Google API credentials
func (c *CredentialStore) GetGoogleCredentials() (GoogleCredentials, error) {
	clientSecret, err := c.keyring.GetCredential(KeyGoogleClientSecret)
	if err != nil {
		return GoogleCredentials{}, err
	}
	
	// Access token and refresh token are optional
	accessToken, _ := c.keyring.GetCredential(KeyGoogleAccessToken)
	refreshToken, _ := c.keyring.GetCredential(KeyGoogleRefreshToken)
	
	return GoogleCredentials{
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAllCredentials validates that all required credentials are present
func (c *CredentialStore) ValidateAllCredentials() error {
	var validationErrors utils.ValidationErrors
	
	// Validate Jira credentials
	if err := c.keyring.ValidateCredential(KeyJiraToken); err != nil {
		validationErrors.Add("jira_token", "Jira token is required", nil)
	}
	
	// Validate Gemini credentials
	if err := c.keyring.ValidateCredential(KeyGeminiAPIKey); err != nil {
		validationErrors.Add("gemini_api_key", "Gemini API key is required", nil)
	}
	
	// Validate Google credentials
	if err := c.keyring.ValidateCredential(KeyGoogleClientSecret); err != nil {
		validationErrors.Add("google_client_secret", "Google client secret is required", nil)
	}
	
	if validationErrors.HasErrors() {
		return validationErrors.ToAppError()
	}
	
	return nil
}

// ClearAllCredentials removes all stored credentials
func (c *CredentialStore) ClearAllCredentials() error {
	keys := []string{
		KeyJiraToken,
		KeyGeminiAPIKey,
		KeyGoogleClientSecret,
		KeyGoogleAccessToken,
		KeyGoogleRefreshToken,
		KeyEncryptionKey,
	}
	
	var errors []error
	for _, key := range keys {
		if err := c.keyring.DeleteCredential(key); err != nil {
			// Only log errors for keys that actually exist
			if appErr, ok := err.(*utils.AppError); ok && appErr.Code != utils.ErrorCodeCredentialsMissing {
				errors = append(errors, err)
			}
		}
	}
	
	if len(errors) > 0 {
		return utils.NewAppError(utils.ErrorCodeKeyringError, "Failed to clear some credentials", errors[0])
	}
	
	c.logger.Info("All credentials cleared successfully")
	
	return nil
}