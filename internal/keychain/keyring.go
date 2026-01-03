package keychain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/99designs/keyring"
	"golang.org/x/term"
)

const serviceName = "com.headjack.cli"

// Environment variable names for keyring configuration.
const (
	EnvKeyringBackend  = "HEADJACK_KEYRING_BACKEND"
	EnvKeyringPassword = "HEADJACK_KEYRING_PASSWORD"
)

type keyringStore struct {
	ring keyring.Keyring
}

// New creates a new Keychain with default configuration.
// Uses auto-detection to select the appropriate backend for the platform.
func New() (Keychain, error) {
	return NewWithConfig(Config{})
}

// NewWithConfig creates a new Keychain with the specified configuration.
func NewWithConfig(cfg Config) (Keychain, error) {
	backend := cfg.Backend
	if backend == BackendAuto {
		backend = detectBackend()
	}

	// Check for environment variable override
	if envBackend := os.Getenv(EnvKeyringBackend); envBackend != "" {
		backend = Backend(envBackend)
	}

	ring, err := openKeyring(backend, cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring (%s): %w", backend, err)
	}

	return &keyringStore{ring: ring}, nil
}

// detectBackend returns the best available backend for the current platform.
func detectBackend() Backend {
	switch runtime.GOOS {
	case "darwin":
		return BackendKeychain
	case "windows":
		return BackendWinCred
	case "linux":
		// Try secret-service first (works with GNOME Keyring, KWallet via D-Bus)
		if isSecretServiceAvailable() {
			return BackendSecretService
		}
		// Fall back to keyctl (Linux kernel keyring, works headless)
		if isKeyctlAvailable() {
			return BackendKeyctl
		}
		// Last resort: encrypted file
		return BackendFile
	default:
		return BackendFile
	}
}

// isSecretServiceAvailable checks if the Secret Service D-Bus API is available.
func isSecretServiceAvailable() bool {
	// Check if D-Bus session is available by looking for the socket
	if dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS"); dbusAddr != "" {
		return true
	}
	// Also check for the default socket path
	if xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntimeDir != "" {
		socketPath := filepath.Join(xdgRuntimeDir, "bus")
		if _, err := os.Stat(socketPath); err == nil {
			return true
		}
	}
	return false
}

// isKeyctlAvailable checks if the Linux kernel keyring is available.
func isKeyctlAvailable() bool {
	// keyctl is available on all modern Linux kernels
	return runtime.GOOS == "linux"
}

// openKeyring opens a keyring with the specified backend and configuration.
func openKeyring(backend Backend, cfg Config) (keyring.Keyring, error) {
	fileDir := cfg.FileDir
	if fileDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		fileDir = filepath.Join(home, ".config", "headjack")
	}

	passwordFunc := cfg.PasswordFunc
	if passwordFunc == nil {
		passwordFunc = defaultPasswordFunc
	}

	config := keyring.Config{
		ServiceName: serviceName,

		// macOS Keychain options
		KeychainName:                   "login",
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,

		// Encrypted file backend options
		FileDir:          fileDir,
		FilePasswordFunc: passwordFunc,

		// Restrict to specified backend
		AllowedBackends: []keyring.BackendType{keyring.BackendType(backend)},
	}

	return keyring.Open(config)
}

// defaultPasswordFunc provides a password for the encrypted file backend.
// It first checks the environment variable, then falls back to interactive prompt.
func defaultPasswordFunc(prompt string) (string, error) {
	// Check environment variable first (for CI/headless use)
	if pw := os.Getenv(EnvKeyringPassword); pw != "" {
		return pw, nil
	}

	// Try interactive prompt if running in a terminal
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return terminalPrompt(prompt)
	}

	return "", ErrNoPassword
}

// terminalPrompt reads a password from the terminal without echoing.
func terminalPrompt(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // Print newline after password entry
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(password), nil
}

func (k *keyringStore) Set(account, secret string) error {
	return k.ring.Set(keyring.Item{
		Key:         account,
		Data:        []byte(secret),
		Label:       "Headjack - " + account,
		Description: "Headjack CLI credential",
	})
}

func (k *keyringStore) Get(account string) (string, error) {
	item, err := k.ring.Get(account)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

func (k *keyringStore) Delete(account string) error {
	err := k.ring.Remove(account)
	if err == nil {
		return nil
	}
	// Handle both keyring.ErrKeyNotFound and filesystem "no such file" errors
	// for idempotent delete behavior
	if errors.Is(err, keyring.ErrKeyNotFound) || os.IsNotExist(err) {
		return nil
	}
	return err
}
