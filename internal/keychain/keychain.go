// Package keychain provides secure credential storage using platform-native keychains.
//
// On macOS, credentials are stored in the macOS Keychain. On Linux, the package
// attempts to use the Secret Service D-Bus API (GNOME Keyring, KWallet), falls back
// to the Linux kernel keyring (keyctl), and finally to an encrypted file. On Windows,
// credentials are stored in Windows Credential Manager.
//
// The backend can be overridden using the HEADJACK_KEYRING_BACKEND environment variable.
// For the encrypted file backend, the password can be provided via HEADJACK_KEYRING_PASSWORD.
package keychain

import "errors"

// ErrNotFound is returned when a credential is not found in the keychain.
var ErrNotFound = errors.New("credential not found in keychain")

// ErrNoPassword is returned when the file backend needs a password but none is available.
var ErrNoPassword = errors.New("keyring password required: set HEADJACK_KEYRING_PASSWORD or run in interactive terminal")

// Backend represents a keyring backend type.
type Backend string

// Supported keyring backends.
const (
	// BackendAuto automatically selects the best available backend for the platform.
	BackendAuto Backend = ""

	// BackendKeychain uses macOS Keychain (darwin only).
	BackendKeychain Backend = "keychain"

	// BackendSecretService uses the Secret Service D-Bus API (Linux with GNOME Keyring or KWallet).
	BackendSecretService Backend = "secret-service"

	// BackendKeyctl uses the Linux kernel keyring (headless Linux).
	BackendKeyctl Backend = "keyctl"

	// BackendWinCred uses Windows Credential Manager (Windows only).
	BackendWinCred Backend = "wincred"

	// BackendFile uses an encrypted file (universal fallback).
	BackendFile Backend = "file"
)

// Config holds configuration for the keyring.
type Config struct {
	// Backend specifies the backend to use. Empty string means auto-detect.
	Backend Backend

	// FileDir is the directory for encrypted file backend.
	// Defaults to ~/.config/headjack/
	FileDir string

	// PasswordFunc provides a password for the encrypted file backend.
	// If nil, HEADJACK_KEYRING_PASSWORD env var is checked, then interactive prompt.
	PasswordFunc func(string) (string, error)
}

// Keychain provides secure credential storage.
//
//go:generate go run github.com/matryer/moq@latest -pkg mocks -out mocks/keychain.go . Keychain
type Keychain interface {
	// Set stores a credential in the keychain.
	Set(account, secret string) error

	// Get retrieves a credential from the keychain.
	// Returns ErrNotFound if the credential does not exist.
	Get(account string) (string, error)

	// Delete removes a credential from the keychain.
	// Returns nil if the credential does not exist.
	Delete(account string) error
}
