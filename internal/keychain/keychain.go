// Package keychain provides access to macOS Keychain for secure credential storage.
package keychain

import (
	"errors"

	gokeychain "github.com/keybase/go-keychain"
)

// ErrNotFound is returned when a credential is not found in the keychain.
var ErrNotFound = errors.New("credential not found in keychain")

// serviceName is the service identifier used for all headjack credentials.
const serviceName = "com.headjack.cli"

// Keychain provides secure credential storage using macOS Keychain.
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

type keychain struct{}

// New creates a new Keychain backed by macOS Keychain.
func New() Keychain {
	return &keychain{}
}

func (k *keychain) Set(account, secret string) error {
	// First try to delete any existing entry to avoid duplicates.
	// Ignore errors since the item may not exist.
	_ = k.Delete(account) //nolint:errcheck // Intentionally ignoring - item may not exist

	item := gokeychain.NewItem()
	item.SetSecClass(gokeychain.SecClassGenericPassword)
	item.SetService(serviceName)
	item.SetAccount(account)
	item.SetLabel("Headjack - " + account)
	item.SetData([]byte(secret))
	item.SetSynchronizable(gokeychain.SynchronizableNo)
	item.SetAccessible(gokeychain.AccessibleWhenUnlocked)

	return gokeychain.AddItem(item)
}

func (k *keychain) Get(account string) (string, error) {
	query := gokeychain.NewItem()
	query.SetSecClass(gokeychain.SecClassGenericPassword)
	query.SetService(serviceName)
	query.SetAccount(account)
	query.SetMatchLimit(gokeychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := gokeychain.QueryItem(query)
	if err == gokeychain.ErrorItemNotFound {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", ErrNotFound
	}

	return string(results[0].Data), nil
}

func (k *keychain) Delete(account string) error {
	item := gokeychain.NewItem()
	item.SetSecClass(gokeychain.SecClassGenericPassword)
	item.SetService(serviceName)
	item.SetAccount(account)

	err := gokeychain.DeleteItem(item)
	if err == gokeychain.ErrorItemNotFound {
		return nil
	}
	return err
}
