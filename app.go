package main

import "context"

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
