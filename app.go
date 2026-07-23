package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App owns desktop lifecycle state and will expose backend services to the UI.
type App struct {
	ctx    context.Context
	drafts DraftStore
}

func NewApp() *App {
	return &App{drafts: DraftStore{appName: "PandaWriter"}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
