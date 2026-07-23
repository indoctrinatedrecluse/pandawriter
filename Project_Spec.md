# Project Specification: Personal AI-Assisted Story Writer

## 1. Overview & Purpose
A lightweight, single-user desktop editor designed for writing stories with real-time visual styling and automated AI-driven illustration features.

*   **Live Preview & Custom Styling:** WYSIWYG editing with a selection of customizable backgrounds and funky typography to set the writing mood.
*   **Automated Contextual Illustration:** Evaluates finished paragraphs using a user-provided DeepSeek API key (BYOK) to identify emotionally intense/serious scenes, automatically fetching or generating relevant background visual illustrations.
*   **1-Click Publishing:** Quick export/publishing pipeline to Google Blogger or a personal portfolio website using configured APIs.

---

## 2. Technical Stack

*   **Wails (Framework & Wrapper):** Serves as the desktop container application, providing a lightweight Native UI wrapper around Webview without the overhead of Electron. Bridges frontend events directly with Go methods.
*   **Svelte / Vue.js (Frontend UI):** Powers the interactive user interface, handling layout state, font selections, visual theme overlays, and API settings modals.
*   **Go (Backend Engine):** Handles local filesystem access, secure API key storage, orchestrating LLM calls (DeepSeek), executing web image search/generation pipelines, and submitting publish payloads to remote APIs.
*   **TipTap (WYSIWYG Rich Text Editor):** Headless, extensible rich-text editor core integrated into the frontend to manage real-time text formatting, paragraph detection hooks, and live visual rendering.

---

## 3. Scope & Constraints

*   **Target Platform:** Desktop application compiled into a standalone Windows binary (`.exe`).
*   **Audience:** Personal project built for a single local user (non-collaborative, no multi-tenant auth or real-time co-editing).
*   **Key Model:** Bring Your Own Key (BYOK) for DeepSeek API requests, keeping infrastructure cost at zero.