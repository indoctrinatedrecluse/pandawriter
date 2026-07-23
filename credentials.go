package main

import (
	"errors"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	credentialService     = "pandawriter"
	deepSeekKeyAccount    = "deepseek-api-key"
	unsplashKeyAccount = "unsplash-access-key"
)

// CredentialStore is the boundary between PandaWriter and the OS credential store.
// The frontend must use configured-status methods rather than reading a secret back.
type CredentialStore struct{}

func (CredentialStore) HasDeepSeekAPIKey() (bool, error) {
	_, err := keyring.Get(credentialService, deepSeekKeyAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (CredentialStore) SaveDeepSeekAPIKey(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return errors.New("DeepSeek API key cannot be empty")
	}
	return keyring.Set(credentialService, deepSeekKeyAccount, apiKey)
}

func (CredentialStore) RemoveDeepSeekAPIKey() error {
	err := keyring.Delete(credentialService, deepSeekKeyAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (CredentialStore) HasUnsplashAPIKey() (bool, error) {
	_, err := keyring.Get(credentialService, unsplashKeyAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (CredentialStore) SaveUnsplashAPIKey(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return errors.New("Unsplash API key cannot be empty")
	}
	return keyring.Set(credentialService, unsplashKeyAccount, apiKey)
}

func (CredentialStore) RemoveUnsplashAPIKey() error {
	err := keyring.Delete(credentialService, unsplashKeyAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
