# PandaWriter POC Dependencies

This document is the approved dependency baseline for the first proof of concept.
Versions should be pinned in `go.mod` and `package.json` when the project is scaffolded; do not use a dependency solely because it appears here.

## Runtime dependencies

### Go backend

| Dependency | Purpose | POC status |
| --- | --- | --- |
| `github.com/wailsapp/wails/v2` | Windows desktop application shell and Go-to-frontend bindings. | Required |
| `github.com/zalando/go-keyring` | Stores the DeepSeek API key in Windows Credential Manager. | Required |
| Go standard library: `net/http`, `encoding/json`, `os`, `path/filepath` | DeepSeek requests and local story/settings persistence. | Required |
| `golang.org/x/oauth2` | Google OAuth flow for Blogger publishing. | Deferred |
| `google.golang.org/api/blogger/v3` | Blogger API client. | Deferred |

### Frontend

| Dependency | Purpose | POC status |
| --- | --- | --- |
| `svelte` | Reactive desktop UI. | Required |
| `vite` | Frontend development server and production bundler. | Required |
| `@sveltejs/vite-plugin-svelte` | Vite integration for Svelte. | Required |
| `typescript` | Type-safe frontend code. | Required |
| `@tiptap/core` | Headless rich-text editor. | Required |
| `@tiptap/starter-kit` | Essential TipTap editing extensions. | Required |
| `@tiptap/extension-placeholder` | Empty-editor prompt. | Required |
| `@tiptap/extension-link` | Link formatting. | Required |
| `@tiptap/extension-underline` | Underline formatting. | Required |
| `@tiptap/extension-text-style` | Text-style base extension. | Required |
| `@tiptap/extension-color` | Text-colour styling. | Required |

## Development dependencies

| Dependency / tool | Purpose | POC status |
| --- | --- | --- |
| Wails CLI (`github.com/wailsapp/wails/v2/cmd/wails`) | Development and Windows build commands. | Required |
| `vitest` | Frontend unit tests. | Add with UI tests |
| `@testing-library/svelte` | Svelte component test helpers. | Add with UI tests |
| ESLint and Prettier | Frontend linting and formatting. | Recommended |
| `gofmt` | Go formatting (included with Go). | Required |

## External services (not code dependencies)

| Service | Purpose | POC status |
| --- | --- | --- |
| DeepSeek Chat Completions API | Classifies completed paragraphs and returns structured scene metadata. | Required; user supplies the key |
| Image provider API | Supplies a scene illustration or background image. Provider to be selected. | Deferred |
| Google Blogger API | One-click publishing. Requires OAuth credentials and consent configuration. | Deferred |

## Credential storage policy

The DeepSeek key must never be hardcoded, committed, or stored in the story files. The Go backend stores it in the operating-system credential store using `go-keyring` under:

```text
service: pandawriter
account: deepseek-api-key
```

The frontend may report whether a key is configured, but must not persist or reveal the full key.

## Deliberate POC exclusions

- No database initially: story documents and non-secret settings are saved as local JSON files.
- No DeepSeek SDK: the Go standard HTTP client is sufficient for the API calls.
- No image-generation SDK until an image provider and cost model are selected.
- No Blogger dependencies until the editor and local-save path are working.
