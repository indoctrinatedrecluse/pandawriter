package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const draftFileName = "step-one-draft.pwr"

// Draft is the full local state required to restore the writing room.
type Draft struct {
	Exists    bool   `json:"exists"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Theme     string `json:"theme"`
	Font      string `json:"font"`
	FontSize  string `json:"fontSize"`
	Spacing   string `json:"spacing"`
	UpdatedAt string `json:"updatedAt"`
}

type DraftStore struct {
	appName string
}

func (s DraftStore) Load() (Draft, error) {
	path, err := s.path()
	if err != nil {
		return Draft{}, err
	}

	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Draft{}, nil
	}
	if err != nil {
		return Draft{}, err
	}

	var draft Draft
	if err := json.Unmarshal(contents, &draft); err != nil {
		return Draft{}, err
	}
	draft.Exists = true
	return draft, nil
}

func (s DraftStore) Save(draft Draft) error {
	path, err := s.path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	draft.Title = strings.TrimSpace(draft.Title)
	draft.Content = strings.TrimSpace(draft.Content)
	draft.Theme = strings.TrimSpace(draft.Theme)
	draft.Font = strings.TrimSpace(draft.Font)
	draft.FontSize = strings.TrimSpace(draft.FontSize)
	draft.Spacing = strings.TrimSpace(draft.Spacing)
	draft.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	draft.Exists = true

	contents, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return err
	}

	temporaryFile, err := os.CreateTemp(filepath.Dir(path), ".draft-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporaryFile.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporaryFile.Write(contents); err != nil {
		temporaryFile.Close()
		return err
	}
	if err := temporaryFile.Chmod(0o600); err != nil {
		temporaryFile.Close()
		return err
	}
	if err := temporaryFile.Close(); err != nil {
		return err
	}

	return os.Rename(temporaryPath, path)
}

func (s DraftStore) path() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, s.appName, draftFileName), nil
}