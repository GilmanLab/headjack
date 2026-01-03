package keychain

import (
	"os"
	"runtime"
	"testing"
)

const testPassword = "test-password"

func testPasswordFunc(string) (string, error) {
	return testPassword, nil
}

func TestNew_DefaultBackend(t *testing.T) {
	// Use file backend for testing to avoid needing actual system keyrings
	t.Setenv(EnvKeyringBackend, string(BackendFile))
	t.Setenv(EnvKeyringPassword, testPassword)

	store, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if store == nil {
		t.Fatal("New() returned nil store")
	}
}

func TestNewWithConfig_FileBackend(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewWithConfig(Config{
		Backend:      BackendFile,
		FileDir:      tmpDir,
		PasswordFunc: testPasswordFunc,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() failed: %v", err)
	}
	if store == nil {
		t.Fatal("NewWithConfig() returned nil store")
	}
}

func TestKeyringStore_SetGet(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewWithConfig(Config{
		Backend:      BackendFile,
		FileDir:      tmpDir,
		PasswordFunc: testPasswordFunc,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() failed: %v", err)
	}

	// Set a credential
	err = store.Set("test-account", "test-secret")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get the credential
	secret, err := store.Get("test-account")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if secret != "test-secret" {
		t.Errorf("Get() = %q, want %q", secret, "test-secret")
	}
}

func TestKeyringStore_GetNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewWithConfig(Config{
		Backend:      BackendFile,
		FileDir:      tmpDir,
		PasswordFunc: testPasswordFunc,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() failed: %v", err)
	}

	_, err = store.Get("nonexistent")
	if err != ErrNotFound {
		t.Errorf("Get() error = %v, want %v", err, ErrNotFound)
	}
}

func TestKeyringStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewWithConfig(Config{
		Backend:      BackendFile,
		FileDir:      tmpDir,
		PasswordFunc: testPasswordFunc,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() failed: %v", err)
	}

	// Set then delete
	err = store.Set("delete-test", "secret")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	err = store.Delete("delete-test")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	_, err = store.Get("delete-test")
	if err != ErrNotFound {
		t.Errorf("Get() after Delete() error = %v, want %v", err, ErrNotFound)
	}
}

func TestKeyringStore_DeleteIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewWithConfig(Config{
		Backend:      BackendFile,
		FileDir:      tmpDir,
		PasswordFunc: testPasswordFunc,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() failed: %v", err)
	}

	// Delete nonexistent should succeed (idempotent)
	err = store.Delete("nonexistent")
	if err != nil {
		t.Errorf("Delete() of nonexistent key failed: %v", err)
	}
}

func TestKeyringStore_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewWithConfig(Config{
		Backend:      BackendFile,
		FileDir:      tmpDir,
		PasswordFunc: testPasswordFunc,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() failed: %v", err)
	}

	// Set initial value
	err = store.Set("overwrite-test", "initial")
	if err != nil {
		t.Fatalf("Set() initial failed: %v", err)
	}

	// Overwrite with new value
	err = store.Set("overwrite-test", "updated")
	if err != nil {
		t.Fatalf("Set() overwrite failed: %v", err)
	}

	// Verify updated value
	secret, err := store.Get("overwrite-test")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if secret != "updated" {
		t.Errorf("Get() = %q, want %q", secret, "updated")
	}
}

func TestDetectBackend(t *testing.T) {
	backend := detectBackend()

	switch runtime.GOOS {
	case "darwin":
		if backend != BackendKeychain {
			t.Errorf("detectBackend() on darwin = %v, want %v", backend, BackendKeychain)
		}
	case "windows":
		if backend != BackendWinCred {
			t.Errorf("detectBackend() on windows = %v, want %v", backend, BackendWinCred)
		}
	default:
		// Linux and other platforms
		validBackends := map[Backend]bool{
			BackendSecretService: true,
			BackendKeyctl:        true,
			BackendFile:          true,
		}
		if !validBackends[backend] {
			t.Errorf("detectBackend() on %s = %v, want one of secret-service, keyctl, or file", runtime.GOOS, backend)
		}
	}
}

func TestEnvBackendOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv(EnvKeyringBackend, string(BackendFile))
	t.Setenv(EnvKeyringPassword, testPassword)

	store, err := NewWithConfig(Config{
		FileDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewWithConfig() with env override failed: %v", err)
	}
	if store == nil {
		t.Fatal("NewWithConfig() returned nil store")
	}
}

func TestDefaultPasswordFunc_EnvVar(t *testing.T) {
	t.Setenv(EnvKeyringPassword, "env-password")

	password, err := defaultPasswordFunc("Enter password: ")
	if err != nil {
		t.Fatalf("defaultPasswordFunc() failed: %v", err)
	}
	if password != "env-password" {
		t.Errorf("defaultPasswordFunc() = %q, want %q", password, "env-password")
	}
}

func TestDefaultPasswordFunc_NoPassword(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv(EnvKeyringPassword)

	// In non-terminal context (like tests), should return ErrNoPassword
	_, err := defaultPasswordFunc("Enter password: ")
	if err != ErrNoPassword {
		t.Errorf("defaultPasswordFunc() error = %v, want %v", err, ErrNoPassword)
	}
}
