package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App owns desktop lifecycle state and will expose backend services to the UI.
type App struct {
	ctx         context.Context
	drafts      DraftStore
	credentials CredentialStore
	ai          *AI
}

func NewApp() *App {
	ai, err := NewAI()
	if err != nil {
		// Log the error but don't prevent startup.
		// The user can configure the key later.
		println("Could not initialize AI client:", err.Error())
	}
	return &App{
		drafts:      DraftStore{appName: "PandaWriter"},
		credentials: CredentialStore{},
		ai:          ai,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// HasAnyAPIKey returns true if the DeepSeek API key is configured.
func (a *App) HasAnyAPIKey() bool {
	return a.ai != nil
}

// HasUnsplashAPIKey returns true if the Unsplash access key is configured.
func (a *App) HasUnsplashAPIKey() bool {
	if a.ai == nil {
		return false
	}
	return a.ai.unsplashAccessKey != ""
}

// AnalyzeParagraph analyzes a paragraph of text and returns an analysis.
func (a *App) AnalyzeParagraph(text string) (*Analysis, error) {
	if a.ai == nil {
		return nil, errors.New("AI client not initialized")
	}
	return a.ai.AnalyzeParagraph(a.ctx, text)
}

// CompleteWord returns suggestions to complete a partial word.
func (a *App) CompleteWord(partialWord string, precedingText string) ([]string, error) {
	if a.ai == nil {
		return nil, errors.New("AI client not initialized")
	}
	return a.ai.CompleteWord(a.ctx, partialWord, precedingText)
}

// CanIllustrate checks whether the given paragraph qualifies for illustration analysis.
// Enforces minimum length and rate-limiting cooldown.
func (a *App) CanIllustrate(text string) bool {
	if a.ai == nil {
		return false
	}
	return a.ai.CanIllustrate(text)
}

// GetIllustration returns an illustration description for the given text.
// Performs a single lightweight API call (no theme/font/word-error analysis).
// The caller should check CanIllustrate first.
func (a *App) GetIllustration(text string) (string, error) {
	if a.ai == nil {
		return "", errors.New("AI client not initialized")
	}
	return a.ai.GetIllustration(a.ctx, text)
}

// CompleteParagraph returns a suggested continuation sentence.
func (a *App) CompleteParagraph(precedingText string) (string, error) {
	if a.ai == nil {
		return "", errors.New("AI client not initialized")
	}
	return a.ai.CompleteParagraph(a.ctx, precedingText)
}

// LoadDraft returns the locally saved Step 1 draft, if one exists.
func (a *App) LoadDraft() (Draft, error) {
	return a.drafts.Load()
}

// SaveDraft persists the editor's document and appearance in the current user's
// application-data directory. It intentionally never stores credentials.
func (a *App) SaveDraft(draft Draft) error {
	return a.drafts.Save(draft)
}

// OpenFile shows a dialog to open a .pwr file and returns its contents.
func (a *App) OpenFile() (Draft, string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open PandaWriter Draft",
		Filters: []runtime.FileFilter{
			{DisplayName: "PandaWriter Drafts (*.pwr)", Pattern: "*.pwr"},
		},
	})
	if err != nil {
		return Draft{}, "", err
	}
	if path == "" {
		return Draft{}, "", nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Draft{}, "", err
	}

	var draft Draft
	if err := json.Unmarshal(content, &draft); err != nil {
		return Draft{}, "", err
	}

	return draft, path, nil
}

// SaveFile writes the given draft to the specified path.
func (a *App) SaveFile(path string, draft Draft) error {
	content, err := json.MarshalIndent(&draft, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

// SaveFileAs shows a dialog to save the given draft to a new .pwr file.
func (a *App) SaveFileAs(draft Draft) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Save PandaWriter Draft",
		Filters: []runtime.FileFilter{
			{DisplayName: "PandaWriter Drafts (*.pwr)", Pattern: "*.pwr"},
		},
		DefaultFilename: "story.pwr",
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	err = a.SaveFile(path, draft)
	if err != nil {
		return "", err
	}
	return path, nil
}