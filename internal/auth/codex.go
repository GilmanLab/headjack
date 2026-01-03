package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// codexConfigDir is the path where Codex CLI stores its configuration.
var codexConfigDir = filepath.Join(os.Getenv("HOME"), ".codex")

// Codex provider configuration.
var codexInfo = ProviderInfo{
	Name:                   "codex",
	SubscriptionEnvVar:     "CODEX_AUTH_JSON",
	APIKeyEnvVar:           "OPENAI_API_KEY",
	KeychainAccount:        "codex-credential",
	RequiresContainerSetup: true,
}

// CodexProvider authenticates with OpenAI Codex CLI.
type CodexProvider struct{}

// NewCodexProvider creates a new Codex authentication provider.
func NewCodexProvider() *CodexProvider {
	return &CodexProvider{}
}

// Info returns metadata about the Codex provider.
func (p *CodexProvider) Info() ProviderInfo {
	return codexInfo
}

// CheckSubscription reads cached Codex CLI credentials from ~/.codex/auth.json.
// If credentials exist and are valid, returns them as a JSON string.
// If credentials don't exist, returns an error with instructions.
func (p *CodexProvider) CheckSubscription() (string, error) {
	authData, err := readCodexAuth()
	if err != nil {
		return "", err
	}
	return string(authData), nil
}

// ValidateSubscription validates Codex auth.json credentials.
func (p *CodexProvider) ValidateSubscription(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("credentials cannot be empty")
	}
	// Codex auth.json should be valid JSON
	if !strings.HasPrefix(value, "{") {
		return errors.New("invalid auth.json: must be a JSON object")
	}
	return nil
}

// ValidateAPIKey validates an OpenAI API key.
func (p *CodexProvider) ValidateAPIKey(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("API key cannot be empty")
	}
	// OpenAI API keys start with "sk-"
	if !strings.HasPrefix(value, "sk-") {
		return errors.New("invalid OpenAI API key: must start with 'sk-'")
	}
	return nil
}

// Store saves a credential to storage.
func (p *CodexProvider) Store(storage Storage, cred Credential) error {
	return StoreCredential(storage, codexInfo.KeychainAccount, cred)
}

// Load retrieves the stored credential for Codex.
func (p *CodexProvider) Load(storage Storage) (*Credential, error) {
	return LoadCredential(storage, codexInfo.KeychainAccount)
}

// readCodexAuth reads the auth.json file from the Codex config directory.
func readCodexAuth() ([]byte, error) {
	authPath := filepath.Join(codexConfigDir, "auth.json")
	data, err := os.ReadFile(authPath) //nolint:gosec // Path is constructed from HOME env var
	if err != nil {
		if os.IsNotExist(err) {
			//nolint:staticcheck // ST1005: Intentionally capitalized - user-facing instructions
			return nil, errors.New(`Codex credentials not found.

To authenticate with your ChatGPT subscription:
  1. Run: codex login
  2. Complete the OAuth flow in your browser
  3. Run: hjk auth codex`)
		}
		return nil, fmt.Errorf("read auth.json: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("codex auth.json is empty: login may have failed")
	}

	return data, nil
}
