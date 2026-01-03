package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// geminiConfigDir is the path where Gemini CLI stores its configuration.
var geminiConfigDir = filepath.Join(os.Getenv("HOME"), ".gemini")

// Gemini provider configuration.
var geminiInfo = ProviderInfo{
	Name:                   "gemini",
	SubscriptionEnvVar:     "GEMINI_OAUTH_CREDS",
	APIKeyEnvVar:           "GEMINI_API_KEY",
	KeychainAccount:        "gemini-credential",
	RequiresContainerSetup: true,
}

// GeminiConfig holds all configuration needed to authenticate Gemini CLI.
type GeminiConfig struct {
	OAuthCreds     json.RawMessage `json:"oauth_creds"`
	GoogleAccounts json.RawMessage `json:"google_accounts"`
}

// GeminiProvider authenticates with Gemini CLI.
type GeminiProvider struct{}

// NewGeminiProvider creates a new Gemini authentication provider.
func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{}
}

// Info returns metadata about the Gemini provider.
func (p *GeminiProvider) Info() ProviderInfo {
	return geminiInfo
}

// CheckSubscription reads cached Gemini CLI credentials from ~/.gemini/.
// If credentials exist and are valid, returns them as a JSON string.
// If credentials don't exist, returns an error with instructions.
func (p *GeminiProvider) CheckSubscription() (string, error) {
	config, err := readGeminiConfig()
	if err != nil {
		return "", err
	}

	// Marshal the config to JSON for storage
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	return string(configJSON), nil
}

// ValidateSubscription validates Gemini OAuth credentials.
func (p *GeminiProvider) ValidateSubscription(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("credentials cannot be empty")
	}

	// Try to parse as GeminiConfig JSON
	var config GeminiConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	if len(config.OAuthCreds) == 0 {
		return errors.New("missing oauth_creds in credentials")
	}
	if len(config.GoogleAccounts) == 0 {
		return errors.New("missing google_accounts in credentials")
	}

	// Validate OAuth creds have a refresh token
	var oauthCreds struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(config.OAuthCreds, &oauthCreds); err != nil {
		return fmt.Errorf("parse oauth_creds: %w", err)
	}
	if oauthCreds.RefreshToken == "" {
		return errors.New("missing refresh_token in oauth_creds")
	}

	return nil
}

// ValidateAPIKey validates a Google AI API key.
func (p *GeminiProvider) ValidateAPIKey(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("API key cannot be empty")
	}
	// Google AI API keys typically start with "AIza"
	if !strings.HasPrefix(value, "AIza") {
		return errors.New("invalid Google AI API key: must start with 'AIza'")
	}
	return nil
}

// Store saves a credential to storage.
func (p *GeminiProvider) Store(storage Storage, cred Credential) error {
	return StoreCredential(storage, geminiInfo.KeychainAccount, cred)
}

// Load retrieves the stored credential for Gemini.
func (p *GeminiProvider) Load(storage Storage) (*Credential, error) {
	return LoadCredential(storage, geminiInfo.KeychainAccount)
}

// readGeminiConfig reads OAuth credentials and account info from Gemini CLI's cache.
func readGeminiConfig() (*GeminiConfig, error) {
	// Read oauth_creds.json (required)
	oauthPath := filepath.Join(geminiConfigDir, "oauth_creds.json")
	oauthData, err := os.ReadFile(oauthPath) //nolint:gosec // Path is constructed from HOME env var
	if err != nil {
		if os.IsNotExist(err) {
			//nolint:staticcheck // ST1005: Intentionally capitalized - user-facing instructions
			return nil, errors.New(`Gemini credentials not found.

To authenticate with your Gemini subscription:
  1. Run: gemini
  2. Complete the Google OAuth login
  3. Run: hjk auth gemini`)
		}
		return nil, fmt.Errorf("read oauth_creds.json: %w", err)
	}

	// Validate OAuth creds have a refresh token
	var oauthCreds struct {
		RefreshToken string `json:"refresh_token"`
	}
	if unmarshalErr := json.Unmarshal(oauthData, &oauthCreds); unmarshalErr != nil {
		return nil, fmt.Errorf("parse oauth_creds.json: %w", unmarshalErr)
	}
	if oauthCreds.RefreshToken == "" {
		return nil, errors.New("gemini credentials missing refresh token: please run 'gemini' and complete the OAuth login")
	}

	// Read google_accounts.json (required for OAuth)
	accountsPath := filepath.Join(geminiConfigDir, "google_accounts.json")
	accountsData, err := os.ReadFile(accountsPath) //nolint:gosec // Path is constructed from HOME env var
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("google_accounts.json not found: please run 'gemini' and complete the OAuth login first")
		}
		return nil, fmt.Errorf("read google_accounts.json: %w", err)
	}

	return &GeminiConfig{
		OAuthCreds:     oauthData,
		GoogleAccounts: accountsData,
	}, nil
}
