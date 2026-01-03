package auth

import (
	"errors"
	"strings"
)

// Claude provider configuration.
var claudeInfo = ProviderInfo{
	Name:                   "claude",
	SubscriptionEnvVar:     "CLAUDE_CODE_OAUTH_TOKEN",
	APIKeyEnvVar:           "ANTHROPIC_API_KEY",
	KeychainAccount:        "claude-credential",
	RequiresContainerSetup: true,
}

// ClaudeProvider authenticates with Claude Code CLI.
type ClaudeProvider struct{}

// NewClaudeProvider creates a new Claude authentication provider.
func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{}
}

// Info returns metadata about the Claude provider.
func (p *ClaudeProvider) Info() ProviderInfo {
	return claudeInfo
}

// CheckSubscription returns instructions for obtaining a Claude OAuth token.
// Unlike Gemini/Codex, Claude requires manual token retrieval via `claude setup-token`.
func (p *ClaudeProvider) CheckSubscription() (string, error) {
	//nolint:staticcheck // ST1005: Intentionally capitalized - user-facing instructions
	return "", errors.New(`Claude subscription credentials must be entered manually.

To get your OAuth token:
  1. Run: claude setup-token
  2. Complete the browser login flow
  3. Copy the token (starts with sk-ant-)`)
}

// ValidateSubscription validates a Claude OAuth token.
func (p *ClaudeProvider) ValidateSubscription(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("token cannot be empty")
	}
	if !strings.HasPrefix(value, "sk-ant-") {
		return errors.New("invalid Claude OAuth token: must start with 'sk-ant-'")
	}
	return nil
}

// ValidateAPIKey validates an Anthropic API key.
func (p *ClaudeProvider) ValidateAPIKey(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("API key cannot be empty")
	}
	// Anthropic API keys start with sk-ant-api
	if !strings.HasPrefix(value, "sk-ant-api") {
		return errors.New("invalid Anthropic API key: must start with 'sk-ant-api'")
	}
	return nil
}

// Store saves a credential to storage.
func (p *ClaudeProvider) Store(storage Storage, cred Credential) error {
	return StoreCredential(storage, claudeInfo.KeychainAccount, cred)
}

// Load retrieves the stored credential for Claude.
func (p *ClaudeProvider) Load(storage Storage) (*Credential, error) {
	return LoadCredential(storage, claudeInfo.KeychainAccount)
}

// isClaudeToken checks if a string looks like a Claude OAuth token.
// Claude tokens have the format: sk-ant-oat01-...
func isClaudeToken(s string) bool {
	return strings.HasPrefix(s, "sk-ant-")
}

// extractToken extracts a Claude OAuth token from command output.
// It scans each line, strips ANSI codes, and returns the first token found.
func extractToken(output string) string {
	// Split on any newline type
	lines := strings.FieldsFunc(output, func(r rune) bool {
		return r == '\n' || r == '\r'
	})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = stripANSI(line)
		if isClaudeToken(line) {
			return line
		}
	}
	return ""
}

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false

	for i := range len(s) {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			// ANSI sequences end with a letter
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
