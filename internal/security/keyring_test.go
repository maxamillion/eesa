package security

import (
	"testing"

	"github.com/company/eesa/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyringManager_StoreAndGetCredential(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	testKey := "test_key"
	testValue := "test_value"
	
	// Store credential
	err := manager.StoreCredential(testKey, testValue)
	require.NoError(t, err)
	
	// Retrieve credential
	retrievedValue, err := manager.GetCredential(testKey)
	require.NoError(t, err)
	assert.Equal(t, testValue, retrievedValue)
	
	// Clean up
	err = manager.DeleteCredential(testKey)
	require.NoError(t, err)
}

func TestKeyringManager_StoreCredential_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	tests := []struct {
		name        string
		key         string
		value       string
		expectError bool
		errorCode   utils.ErrorCode
	}{
		{
			name:        "empty key",
			key:         "",
			value:       "test_value",
			expectError: true,
			errorCode:   utils.ErrorCodeKeyringError,
		},
		{
			name:        "empty value",
			key:         "test_key",
			value:       "",
			expectError: true,
			errorCode:   utils.ErrorCodeKeyringError,
		},
		{
			name:        "valid key and value",
			key:         "test_key",
			value:       "test_value",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.StoreCredential(tt.key, tt.value)
			
			if tt.expectError {
				require.Error(t, err)
				appErr, ok := err.(*utils.AppError)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				require.NoError(t, err)
				// Clean up
				manager.DeleteCredential(tt.key)
			}
		})
	}
}

func TestKeyringManager_GetCredential_ValidationErrors(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	// Test empty key
	_, err := manager.GetCredential("")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeKeyringError, appErr.Code)
	
	// Test non-existent key
	_, err = manager.GetCredential("non_existent_key")
	require.Error(t, err)
	appErr, ok = err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeCredentialsMissing, appErr.Code)
}

func TestKeyringManager_DeleteCredential(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	testKey := "test_key_for_deletion"
	testValue := "test_value"
	
	// Store credential
	err := manager.StoreCredential(testKey, testValue)
	require.NoError(t, err)
	
	// Verify it exists
	assert.True(t, manager.HasCredential(testKey))
	
	// Delete credential
	err = manager.DeleteCredential(testKey)
	require.NoError(t, err)
	
	// Verify it's gone
	assert.False(t, manager.HasCredential(testKey))
	
	// Test deleting non-existent key
	err = manager.DeleteCredential("non_existent_key")
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeCredentialsMissing, appErr.Code)
}

func TestKeyringManager_HasCredential(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	testKey := "test_key_for_existence"
	testValue := "test_value"
	
	// Should not exist initially
	assert.False(t, manager.HasCredential(testKey))
	
	// Store credential
	err := manager.StoreCredential(testKey, testValue)
	require.NoError(t, err)
	
	// Should exist now
	assert.True(t, manager.HasCredential(testKey))
	
	// Clean up
	err = manager.DeleteCredential(testKey)
	require.NoError(t, err)
	
	// Should not exist after deletion
	assert.False(t, manager.HasCredential(testKey))
}

func TestKeyringManager_ListCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	// Initially should be empty or have some existing keys
	_, err := manager.ListCredentials()
	require.NoError(t, err)
	
	// Store some test credentials
	testCredentials := map[string]string{
		KeyJiraToken:    "test_jira_token",
		KeyGeminiAPIKey: "test_gemini_key",
	}
	
	for key, value := range testCredentials {
		err := manager.StoreCredential(key, value)
		require.NoError(t, err)
	}
	
	// List credentials
	keys, err := manager.ListCredentials()
	require.NoError(t, err)
	assert.True(t, len(keys) >= len(testCredentials))
	
	// Check that our test keys are present
	for testKey := range testCredentials {
		assert.Contains(t, keys, testKey)
	}
	
	// Clean up
	for key := range testCredentials {
		manager.DeleteCredential(key)
	}
}

func TestKeyringManager_GenerateEncryptionKey(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	// Clean up any existing encryption key
	manager.DeleteCredential(KeyEncryptionKey)
	
	// Generate encryption key
	err := manager.GenerateEncryptionKey()
	require.NoError(t, err)
	
	// Verify key was stored
	assert.True(t, manager.HasCredential(KeyEncryptionKey))
	
	// Retrieve key
	key, err := manager.GetOrGenerateEncryptionKey()
	require.NoError(t, err)
	assert.Len(t, key, 32) // Should be 32 bytes for AES-256
	
	// Clean up
	err = manager.DeleteCredential(KeyEncryptionKey)
	require.NoError(t, err)
}

func TestKeyringManager_GetOrGenerateEncryptionKey(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	// Clean up any existing encryption key
	manager.DeleteCredential(KeyEncryptionKey)
	
	// Should generate new key if none exists
	key1, err := manager.GetOrGenerateEncryptionKey()
	require.NoError(t, err)
	assert.Len(t, key1, 32)
	
	// Should return the same key on subsequent calls
	key2, err := manager.GetOrGenerateEncryptionKey()
	require.NoError(t, err)
	assert.Equal(t, key1, key2)
	
	// Clean up
	err = manager.DeleteCredential(KeyEncryptionKey)
	require.NoError(t, err)
}

func TestKeyringManager_ValidateCredential(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	testKey := "test_key_for_validation"
	testValue := "test_value"
	
	// Should fail for non-existent key
	err := manager.ValidateCredential(testKey)
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeCredentialsMissing, appErr.Code)
	
	// Store credential
	err = manager.StoreCredential(testKey, testValue)
	require.NoError(t, err)
	
	// Should pass for existing key
	err = manager.ValidateCredential(testKey)
	require.NoError(t, err)
	
	// Clean up
	err = manager.DeleteCredential(testKey)
	require.NoError(t, err)
}

func TestKeyringManager_SecureCompare(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	// Test equal strings
	assert.True(t, manager.SecureCompare("test", "test"))
	assert.True(t, manager.SecureCompare("", ""))
	
	// Test different strings
	assert.False(t, manager.SecureCompare("test", "test2"))
	assert.False(t, manager.SecureCompare("test", ""))
	assert.False(t, manager.SecureCompare("", "test"))
}

func TestKeyringManager_ClearSensitiveData(t *testing.T) {
	logger := utils.NewMockLogger()
	manager := NewKeyringManager(logger)
	
	data := []byte("sensitive data")
	originalData := make([]byte, len(data))
	copy(originalData, data)
	
	// Clear the data
	manager.ClearSensitiveData(data)
	
	// Verify data is cleared
	for i, b := range data {
		assert.Equal(t, byte(0), b, "Byte at index %d should be zero", i)
	}
	
	// Verify original data was actually different
	assert.NotEqual(t, originalData, data)
}

func TestCredentialStore_JiraCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	store := NewCredentialStore(logger)
	
	// Clean up any existing credentials
	store.keyring.DeleteCredential(KeyJiraToken)
	
	testCreds := JiraCredentials{
		Token: "test_jira_token",
	}
	
	// Store credentials
	err := store.SetJiraCredentials(testCreds)
	require.NoError(t, err)
	
	// Retrieve credentials
	retrievedCreds, err := store.GetJiraCredentials()
	require.NoError(t, err)
	assert.Equal(t, testCreds.Token, retrievedCreds.Token)
	
	// Test empty token
	err = store.SetJiraCredentials(JiraCredentials{Token: ""})
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeValidationError, appErr.Code)
	
	// Clean up
	store.keyring.DeleteCredential(KeyJiraToken)
}

func TestCredentialStore_GeminiCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	store := NewCredentialStore(logger)
	
	// Clean up any existing credentials
	store.keyring.DeleteCredential(KeyGeminiAPIKey)
	
	testCreds := GeminiCredentials{
		APIKey: "test_gemini_api_key",
	}
	
	// Store credentials
	err := store.SetGeminiCredentials(testCreds)
	require.NoError(t, err)
	
	// Retrieve credentials
	retrievedCreds, err := store.GetGeminiCredentials()
	require.NoError(t, err)
	assert.Equal(t, testCreds.APIKey, retrievedCreds.APIKey)
	
	// Test empty API key
	err = store.SetGeminiCredentials(GeminiCredentials{APIKey: ""})
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeValidationError, appErr.Code)
	
	// Clean up
	store.keyring.DeleteCredential(KeyGeminiAPIKey)
}

func TestCredentialStore_GoogleCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	store := NewCredentialStore(logger)
	
	// Clean up any existing credentials
	store.keyring.DeleteCredential(KeyGoogleClientSecret)
	store.keyring.DeleteCredential(KeyGoogleAccessToken)
	store.keyring.DeleteCredential(KeyGoogleRefreshToken)
	
	testCreds := GoogleCredentials{
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	}
	
	// Store credentials
	err := store.SetGoogleCredentials(testCreds)
	require.NoError(t, err)
	
	// Retrieve credentials
	retrievedCreds, err := store.GetGoogleCredentials()
	require.NoError(t, err)
	assert.Equal(t, testCreds.ClientSecret, retrievedCreds.ClientSecret)
	assert.Equal(t, testCreds.AccessToken, retrievedCreds.AccessToken)
	assert.Equal(t, testCreds.RefreshToken, retrievedCreds.RefreshToken)
	
	// Test empty client secret
	err = store.SetGoogleCredentials(GoogleCredentials{ClientSecret: ""})
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeValidationError, appErr.Code)
	
	// Clean up
	store.keyring.DeleteCredential(KeyGoogleClientSecret)
	store.keyring.DeleteCredential(KeyGoogleAccessToken)
	store.keyring.DeleteCredential(KeyGoogleRefreshToken)
}

func TestCredentialStore_ValidateAllCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	store := NewCredentialStore(logger)
	
	// Clean up any existing credentials
	store.ClearAllCredentials()
	
	// Should fail when no credentials are set
	err := store.ValidateAllCredentials()
	require.Error(t, err)
	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrorCodeValidationError, appErr.Code)
	
	// Set all required credentials
	err = store.SetJiraCredentials(JiraCredentials{Token: "test_jira_token"})
	require.NoError(t, err)
	
	err = store.SetGeminiCredentials(GeminiCredentials{APIKey: "test_gemini_key"})
	require.NoError(t, err)
	
	err = store.SetGoogleCredentials(GoogleCredentials{ClientSecret: "test_client_secret"})
	require.NoError(t, err)
	
	// Should pass when all credentials are set
	err = store.ValidateAllCredentials()
	require.NoError(t, err)
	
	// Clean up
	store.ClearAllCredentials()
}

func TestCredentialStore_ClearAllCredentials(t *testing.T) {
	logger := utils.NewMockLogger()
	store := NewCredentialStore(logger)
	
	// Set some credentials
	err := store.SetJiraCredentials(JiraCredentials{Token: "test_jira_token"})
	require.NoError(t, err)
	
	err = store.SetGeminiCredentials(GeminiCredentials{APIKey: "test_gemini_key"})
	require.NoError(t, err)
	
	// Verify credentials exist
	assert.True(t, store.keyring.HasCredential(KeyJiraToken))
	assert.True(t, store.keyring.HasCredential(KeyGeminiAPIKey))
	
	// Clear all credentials
	err = store.ClearAllCredentials()
	require.NoError(t, err)
	
	// Verify credentials are cleared
	assert.False(t, store.keyring.HasCredential(KeyJiraToken))
	assert.False(t, store.keyring.HasCredential(KeyGeminiAPIKey))
	assert.False(t, store.keyring.HasCredential(KeyGoogleClientSecret))
	assert.False(t, store.keyring.HasCredential(KeyGoogleAccessToken))
	assert.False(t, store.keyring.HasCredential(KeyGoogleRefreshToken))
	assert.False(t, store.keyring.HasCredential(KeyEncryptionKey))
}