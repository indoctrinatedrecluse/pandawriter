# PandaWriter

PandaWriter is a personal, Windows desktop story editor with rich-text writing tools, mood-setting visual themes, and optional AI-assisted illustration.

It is a single-user, bring-your-own-key (BYOK) application: the writer configures their own DeepSeek API key locally, and the application uses it to identify emotionally significant finished paragraphs for illustration suggestions.

## POC brief

The first proof of concept will prove four things:

1. A writer can create, style, and locally save a story in a desktop editor.
2. The editor can detect a completed paragraph and request structured scene analysis from DeepSeek.
3. The app can present a matching illustration/background suggestion without interrupting writing.
4. A DeepSeek API key can be configured in-app and stored securely in Windows Credential Manager.

Publishing to Blogger and a searchable story library are intentionally deferred until this POC workflow is reliable.

## Stack

| Layer | Technology | Responsibility |
| --- | --- | --- |
| Desktop shell | Wails v2 | Packages the web UI as a native Windows application and exposes Go methods to the frontend. |
| Backend | Go | Local file access, credential storage, DeepSeek requests, and future publishing integrations. |
| Frontend | Svelte, Vite, TypeScript | Application layout, settings, theme selection, and editor state. |
| Editor | TipTap | Extensible WYSIWYG rich-text editing and paragraph hooks. |
| Secret storage | `go-keyring` / Windows Credential Manager | Keeps the user-provided DeepSeek key outside the repository and story files. |
| AI | DeepSeek Chat Completions API | Returns structured scene/importance analysis for finished paragraphs. |

The full dependency register and deferred integrations are in [DEPENDENCIES.md](DEPENDENCIES.md).

## Planned architecture

```text
Svelte + TipTap UI
        │ Wails bindings
        ▼
Go application service ──► Windows Credential Manager
        │
        ├──► local JSON story files
        └──► DeepSeek API ──► Unsplash keyword search ──► Unsplash API ──► photos
```

The API key is accepted only through the Settings UI. Go stores it under the `pandawriter` service in Windows Credential Manager; the frontend only receives a configured/not-configured status.

## Development prerequisites

- Windows 10 or later, with the Microsoft Edge WebView2 Runtime available.
- A supported Go toolchain.
- Node.js LTS and npm.
- Git.
- The Wails CLI, installed with:

  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

Ensure your Go bin directory is on `PATH` so `wails` is available in PowerShell.

## Run in development

These commands apply after the Wails/Svelte project scaffold has been added:

```powershell
npm install --prefix frontend
wails dev
```

`wails dev` starts the Vite development server and launches the desktop app with live reload.

### Development diagnostics

`wails dev` enables Wails debug logging and opens the Web Inspector on startup. Frontend lifecycle messages, draft save/restore errors, uncaught errors, and unhandled promise rejections are logged with a `[PandaWriter]` prefix. An error during frontend startup also replaces the blank window with a development-only diagnostic screen.

## Build and launch the Step 1 POC

From PowerShell at the repository root:

```powershell
.\run.ps1
```

The script removes only the previous `build/bin` contents, builds the Windows executable, and launches it. It uses a Wails CLI found on `PATH`; if one is unavailable, it invokes the project-pinned Wails v2.13.0 CLI through Go instead.

To build without launching the desktop application:

```powershell
.\run.ps1 -NoLaunch
```

## Build a Windows executable

From the repository root:

```powershell
wails build
```

The resulting executable is written to `build/bin/`. This directory is intentionally ignored by Git.

## Repository conventions

- `.idea/` is committed so GoLand settings can be replicated on another machine.
- Do not commit real API keys, OAuth client secrets, or generated binaries.
- Use `gofmt` for Go files and the configured frontend formatter for Svelte/TypeScript files.

## Status

**Step 1** is complete: the local writing room provides a TipTap editor, appearance controls, automatic local draft saving, and draft restoration on the next launch. The draft is saved under the current user's OS application-data directory, outside the repository.

**Step 2** is complete: the editor detects finished paragraphs (Enter at end of a paragraph) and, when enabled, requests structured scene analysis from DeepSeek. The analysis returns theme/font suggestions and an illustration description. Word autocomplete (suggests words as you type after 3+ characters) and sentence autocomplete (Ctrl+Space to complete a sentence) are also available as togglable AI features. A DeepSeek API key can be configured in-app and stored securely in Windows Credential Manager. All AI features are toggled via a sleek on/off menu bar that appears only when an API key is configured.

**Step 3** is complete: when the Illustration feature is enabled, the editor buffers finished paragraphs and, after a 5-second pause, requests keyword analysis from DeepSeek followed by an Unsplash image search. The resulting scenic photograph is displayed in the side panel without interrupting the writing flow. The user can optionally insert the image inline into the document via a one-click button. The Illustration toggle is gated behind an Unsplash API key and rate-limited to one request per 30 seconds to conserve tokens.

**Step 4** is complete: DeepSeek and Unsplash API keys are accepted only through the Settings UI. Each key is stored under the `pandawriter` service in Windows Credential Manager via `go-keyring`. The frontend only receives a configured/not-configured status; secrets never leave the OS credential store.

Publishing to Blogger and a searchable story library remain deferred.
