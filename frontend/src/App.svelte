<script lang="ts">
  import { onMount } from 'svelte'
  import { Editor } from '@tiptap/core'
  import StarterKit from '@tiptap/starter-kit'
  import Placeholder from '@tiptap/extension-placeholder'
  import Link from '@tiptap/extension-link'
  import Underline from '@tiptap/extension-underline'
  import Image from '@tiptap/extension-image'
  import { TextStyle } from '@tiptap/extension-text-style'
  import Color from '@tiptap/extension-color'
  import {
    LoadDraft,
    SaveDraft,
    OpenFile,
    SaveFile,
    SaveFileAs,
    AnalyzeParagraph,
    CompleteWord,
    CompleteParagraph,
    HasAnyAPIKey,
    HasUnsplashAPIKey,
    CanIllustrate,
    GetIllustration
  } from '../wailsjs/go/main/App'
  import { logDebug, logError, logInfo } from './logger'
  import { EventsOn } from '../wailsjs/runtime/runtime'
  import ApiKeyModal from './ApiKeyModal.svelte'
  import { ParagraphHandler, type ParagraphCompletedDetail } from './paragraph-handler'
  import { WordError } from './word-error'
  import type { Analysis, Draft, Theme, Font, WordError as WordErrorType } from './types'

  const themes: Theme[] = [
    { id: 'midnight', name: 'Midnight ink', caption: 'Quiet and cinematic' },
    { id: 'parchment', name: 'Soft parchment', caption: 'Warm and literary' },
    { id: 'blossom', name: 'Electric bloom', caption: 'Playful and bright' },
    { id: 'studio', name: 'Studio blue', caption: 'Focused and clean' },
    { id: 'crimson', name: 'Crimson study', caption: 'Deep red, scholarly' },
    { id: 'seafoam', name: 'Sea foam', caption: 'Coastal aqua calm' },
    { id: 'ember', name: 'Ember glow', caption: 'Warm firelight' },
    { id: 'viola', name: 'Viola dusk', caption: 'Purple twilight' },
    { id: 'moss', name: 'Moss garden', caption: 'Earthy woodland' },
    { id: 'frost', name: 'Frost', caption: 'Icy winter morning' }
  ]

  const fonts: Font[] = [
    { id: 'literary', name: 'Literary', sample: 'Cormorant Garamond' },
    { id: 'editorial', name: 'Editorial', sample: 'DM Sans' },
    { id: 'typewriter', name: 'Typewriter', sample: 'Space Mono' },
    { id: 'playfair', name: 'Playfair', sample: 'Playfair Display' },
    { id: 'inter', name: 'Inter', sample: 'Inter' },
    { id: 'merriweather', name: 'Merriweather', sample: 'Merriweather' },
    { id: 'monoton', name: 'Monoton', sample: 'Monoton' },
    { id: 'bebas', name: 'Bebas Neue', sample: 'Bebas Neue' }
  ]

  let editorElement: HTMLDivElement
  let editor: Editor | null = null
  let title = 'Untitled story'
  let theme = 'midnight'
  let font = 'literary'
  let fontSize = 'normal'
  let spacing = 'comfortable'
  let wordCount = 0
  let isSaved = true
  let isSaving = false
  let saveError = ''
  let isReadyToPersist = false
  let contentVersion = 0
  let saveTimer: ReturnType<typeof setTimeout> | undefined
  let currentPath: string | null = null
  let showApiKeyModal = false
  let analysis: Analysis | null = null

  // Random preview rotation
  let visibleThemes: Theme[] = []
  let visibleFonts: Font[] = []
  let rotating = false
  let rotationTimer: ReturnType<typeof setInterval> | undefined

  function pickRandom<T>(arr: T[], count: number): T[] {
    const shuffled = [...arr].sort(() => Math.random() - 0.5)
    return shuffled.slice(0, count)
  }

  function refreshPreview() {
    rotating = true
    setTimeout(() => {
      visibleThemes = pickRandom(themes, 4)
      visibleFonts = pickRandom(fonts, 4)
      rotating = false
    }, 180)
  }

  // AI feature toggles — only one of word/paragraph autocomplete can be on at a time
  let hasApiKey = false
  let hasUnsplashKey = false
  let wordAutocompleteOn = false
  let paragraphAutocompleteOn = false
  let illustrationOn = false
  let autocompleteDebounceTimer: ReturnType<typeof setTimeout> | undefined
  let autocompleteLastCall = 0
  let suggestionsPopup: { words: string[]; visible: boolean } = { words: [], visible: false }

  const starterContent = `
    <h1>Untitled story</h1>
    <p class="lede">A place for your next impossible thing.</p>
    <p>The rain had already written its silver sentence across the glass when Mira opened the envelope.</p>
    <p>She read the first line twice, then smiled as if the night had finally remembered her name.</p>
  `

  function buildDraft(): Draft {
    return {
      exists: true,
      title: title,
      content: editor?.getHTML() ?? '',
      theme: theme,
      font: font,
      fontSize: fontSize,
      spacing: spacing,
      updatedAt: ''
    }
  }

  function applyDraft(draft: Draft) {
    if (!editor) return
    if (draft.content) editor.commands.setContent(draft.content)
    if (draft.title) title = draft.title
    if (themes.some((t) => t.id === draft.theme)) theme = draft.theme
    if (fonts.some((f) => f.id === draft.font)) font = draft.font
    if (draft.fontSize) fontSize = draft.fontSize
    if (draft.spacing) spacing = draft.spacing
  }

  function updateStats() {
    if (!editor) return
    const text = editor.getText().trim()
    wordCount = text ? text.split(/\s+/).length : 0
    if (isReadyToPersist) markChanged()
  }

  function markChanged() {
    contentVersion += 1
    isSaved = false
    saveError = ''
    if (!currentPath) {
      queueAutoSave()
    }
  }

  function queueAutoSave() {
    if (saveTimer) window.clearTimeout(saveTimer)
    saveTimer = window.setTimeout(() => autoSaveLocalDraft(), 650)
  }

  function autoSaveLocalDraft() {
    if (saveTimer) window.clearTimeout(saveTimer)
    void (async () => {
      if (!editor || !isReadyToPersist) return

      const versionBeingSaved = contentVersion
      isSaving = true
      saveError = ''
      logDebug('Auto-saving local draft', { version: versionBeingSaved })

      try {
        await SaveDraft(buildDraft())
        isSaved = versionBeingSaved === contentVersion
        if (!isSaved) queueAutoSave()
        logDebug('Local draft auto-saved', { version: versionBeingSaved })
      } catch (error) {
        logError('Local draft auto-save failed', error)
        saveError = 'Local save failed'
        isSaved = false
      } finally {
        isSaving = false
      }
    })()
  }

  function newFile() {
    if (!editor) return
    logInfo('Creating new file')
    title = 'Untitled story'
    editor.commands.setContent(starterContent)
    theme = 'midnight'
    font = 'literary'
    fontSize = 'normal'
    spacing = 'comfortable'
    updateWordCount()
    currentPath = null
    isSaved = true
    saveError = ''
    analysis = null
    dismissSuggestions()
  }

  async function openFile() {
    logInfo('Opening file...')
    try {
      const [draft, path] = await OpenFile()
      if (path && editor) {
        applyDraft(draft)
        updateWordCount()
        currentPath = path
        isSaved = true
        saveError = ''
        analysis = null
        dismissSuggestions()
        logInfo('File opened', { path })
      }
    } catch (error) {
      logError('File open failed', error)
      saveError = 'Could not open file'
    }
  }

  async function saveFile() {
    if (!editor) return
    if (currentPath) {
      isSaving = true
      saveError = ''
      logDebug('Saving file', { path: currentPath })
      try {
        await SaveFile(currentPath, buildDraft())
        isSaved = true
        logDebug('File saved', { path: currentPath })
      } catch (error) {
        logError('File save failed', error)
        saveError = 'Could not save file'
        isSaved = false
      } finally {
        isSaving = false
      }
    } else {
      await saveFileAs()
    }
  }

  async function saveFileAs() {
    if (!editor) return
    isSaving = true
    saveError = ''
    logDebug('Saving file as...')
    try {
      const path = await SaveFileAs(buildDraft())
      if (path) {
        currentPath = path
        isSaved = true
        logDebug('File saved as', { path })
      }
    } catch (error) {
      logError('File save as failed', error)
      saveError = 'Could not save file'
      isSaved = false
    } finally {
      isSaving = false
    }
  }

  function selectTheme(nextTheme: string) {
    if (theme === nextTheme) return
    theme = nextTheme
    markChanged()
  }

  function selectFont(nextFont: string) {
    if (font === nextFont) return
    font = nextFont
    markChanged()
  }

  function selectFontSize(nextFontSize: string) {
    if (fontSize === nextFontSize) return
    fontSize = nextFontSize
    markChanged()
  }

  function selectSpacing(nextSpacing: string) {
    if (spacing === nextSpacing) return
    spacing = nextSpacing
    markChanged()
  }

  function updateWordCount() {
    if (!editor) return
    const text = editor.getText().trim()
    wordCount = text ? text.split(/\s+/).length : 0
  }

  async function restoreDraft() {
    if (!editor) return

    logInfo('Restoring local draft')
    try {
      const draft = await LoadDraft()
      if (draft.exists) {
        applyDraft(draft)
      }
      updateWordCount()
      logInfo(draft.exists ? 'Local draft restored' : 'No local draft found')
    } catch (error) {
      logError('Local draft restore failed', error)
      saveError = 'Could not restore local draft'
    } finally {
      isReadyToPersist = true
      isSaved = !saveError
    }
  }

  function command(action: (instance: Editor) => void) {
    if (!editor) return
    action(editor)
    editor.commands.focus()
  }

  function setLink() {
    if (!editor) return
    const previousUrl = editor.getAttributes('link').href as string | undefined
    const url = window.prompt('Paste a link', previousUrl ?? '')
    if (url === null) return

    if (url.trim() === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
      return
    }

    editor.chain().focus().extendMarkRange('link').setLink({ href: url.trim() }).run()
  }

  async function refreshApiKeyStatus() {
    try {
      hasApiKey = await HasAnyAPIKey()
      hasUnsplashKey = hasApiKey && await HasUnsplashAPIKey()
      if (!hasUnsplashKey) illustrationOn = false
      logInfo('API key status refreshed', { hasApiKey, hasUnsplashKey })
    } catch {
      hasApiKey = false
      hasUnsplashKey = false
      illustrationOn = false
    }
  }

  // --- AI Feature Toggles ---

  function toggleWordAutocomplete() {
    wordAutocompleteOn = !wordAutocompleteOn
    if (wordAutocompleteOn) {
      paragraphAutocompleteOn = false
      dismissSuggestions()
    }
    logInfo('Word autocomplete', { enabled: wordAutocompleteOn })
  }

  function toggleParagraphAutocomplete() {
    paragraphAutocompleteOn = !paragraphAutocompleteOn
    if (paragraphAutocompleteOn) {
      wordAutocompleteOn = false
      dismissSuggestions()
    }
    logInfo('Paragraph autocomplete', { enabled: paragraphAutocompleteOn })
  }

  function toggleIllustration() {
    illustrationOn = !illustrationOn
    logInfo('Illustration', { enabled: illustrationOn })
  }

  function dismissSuggestions() {
    suggestionsPopup = { words: [], visible: false }
  }

  // --- Autocomplete Logic ---

  function onEditorKeyUp() {
    if (!editor || !wordAutocompleteOn) return

    const { state } = editor
    const { selection } = state
    const { $from } = selection

    if ($from.parent.type.name !== 'paragraph') {
      dismissSuggestions()
      return
    }

    const nodeText = $from.parent.textContent
    const cursorPos = $from.parentOffset
    const textBeforeCursor = nodeText.substring(0, cursorPos)
    const lastSpaceIndex = textBeforeCursor.lastIndexOf(' ')
    const partialWord = textBeforeCursor.substring(lastSpaceIndex + 1)

    if (partialWord.length < 3) {
      dismissSuggestions()
      return
    }

    const now = Date.now()
    if (now - autocompleteLastCall < 800) {
      return
    }

    if (autocompleteDebounceTimer) window.clearTimeout(autocompleteDebounceTimer)
    autocompleteDebounceTimer = window.setTimeout(async () => {
      const currentText = editor.getText()
      const textBefore = currentText.substring(0, editor.state.selection.$from.pos - 1)
      const lastSpace = textBefore.lastIndexOf(' ')
      const word = textBefore.substring(lastSpace + 1)
      if (word.length < 3) {
        dismissSuggestions()
        return
      }

      const paragraphText = $from.parent.textContent.substring(0, cursorPos)
      autocompleteLastCall = Date.now()
      try {
        const words = await CompleteWord(word, paragraphText)
        if (words && words.length > 0) {
          suggestionsPopup = { words, visible: true }
        } else {
          dismissSuggestions()
        }
      } catch (error) {
        logError('Word autocomplete failed', error)
        dismissSuggestions()
      }
    }, 300)
  }

  function acceptSuggestion(word: string) {
    if (!editor) return
    const { state } = editor
    const { selection } = state
    const { $from } = selection

    const nodeText = $from.parent.textContent
    const cursorPos = $from.parentOffset
    const textBeforeCursor = nodeText.substring(0, cursorPos)
    const lastSpaceIndex = textBeforeCursor.lastIndexOf(' ')
    const partialWord = textBeforeCursor.substring(lastSpaceIndex + 1)

    const from = $from.pos - partialWord.length
    const to = $from.pos

    editor.chain().focus().deleteRange({ from, to }).insertContent(word + ' ').run()
    dismissSuggestions()
  }

  function triggerParagraphAutocomplete() {
    if (!editor || !paragraphAutocompleteOn) return

    const { state } = editor
    const { selection } = state
    const { $from } = selection
    const paragraphText = $from.parent.textContent

    void (async () => {
      try {
        const continuation = await CompleteParagraph(paragraphText)
        if (continuation) {
          editor.chain().focus().insertContent(' ' + continuation).run()
        }
      } catch (error) {
        logError('Paragraph autocomplete failed', error)
      }
    })()
  }

  // --- Illustration Analysis ---

  let pendingIllustrationParagraph: string | null = null
  let pendingIllustrationPos: number = 0
  let illustrationDebounceTimer: ReturnType<typeof setTimeout> | undefined

  function onParagraphCompleted(detail: ParagraphCompletedDetail) {
    logDebug('Paragraph completed', { length: detail.paragraph.length })
    if (!illustrationOn) return
    pendingIllustrationParagraph = detail.paragraph
    pendingIllustrationPos = detail.afterPos
    if (illustrationDebounceTimer) window.clearTimeout(illustrationDebounceTimer)
    illustrationDebounceTimer = window.setTimeout(() => { void flushIllustration() }, 5000)
  }

  async function flushIllustration() {
    if (illustrationDebounceTimer) window.clearTimeout(illustrationDebounceTimer)
    const paragraph = pendingIllustrationParagraph
    if (!paragraph || !illustrationOn) return
    pendingIllustrationParagraph = null

    try {
      const eligible = await CanIllustrate(paragraph)
      if (!eligible) {
        logDebug('Illustration skipped (not eligible)', { length: paragraph.length })
        return
      }
    } catch (error) {
      logError('Could not check illustration eligibility', error)
      return
    }

    try {
      const illustration = await GetIllustration(paragraph)
      if (illustration) {
        analysis = { wordErrors: [], theme: '', font: '', illustration }
        logInfo('Illustration fetched', { illustration: illustration.substring(0, 60) })
      }
    } catch (error) {
      logError('Illustration fetch failed', error)
    }
  }

  function insertImageIntoEditor() {
    if (!editor || !analysis?.illustration) return
    const imageURL = analysis.illustration
    const pos = pendingIllustrationPos
    editor.chain().focus().setTextSelection(pos).setImage({ src: imageURL, alt: 'Scene illustration' }).run()
    logInfo('Illustration inserted into editor', { pos })
  }

  // --- Layout menu event handlers ---

  function onLayoutThemeEvent(themeID: string) {
    if (themes.some((t) => t.id === themeID)) {
      selectTheme(themeID)
      logInfo('Layout menu: theme', { themeID })
    }
  }

  function onLayoutFontEvent(fontID: string) {
    if (fonts.some((f) => f.id === fontID)) {
      selectFont(fontID)
      logInfo('Layout menu: font', { fontID })
    }
  }

  onMount(async () => {
    try {
      hasApiKey = await HasAnyAPIKey()
      hasUnsplashKey = hasApiKey && await HasUnsplashAPIKey()
      logInfo('API key status', { hasApiKey, hasUnsplashKey })
    } catch {
      hasApiKey = false
      hasUnsplashKey = false
    }

    editor = new Editor({
      element: editorElement,
      extensions: [
        StarterKit.configure({ heading: { levels: [1, 2, 3] }, link: false }),
        Placeholder.configure({ placeholder: 'Begin where the story starts…' }),
        Link.configure({ openOnClick: false, autolink: true }),
        Underline,
        Image,
        TextStyle,
        Color,
        ParagraphHandler.configure({ onParagraphCompleted }),
        WordError
      ],
      content: starterContent,
      editorProps: {
        attributes: { class: 'story-prose' },
        handleKeyUp: () => { onEditorKeyUp(); return false }
      },
      onUpdate: updateStats
    })

    logInfo('TipTap editor initialized')
    void restoreDraft()

    // Initialize preview rotation
    visibleThemes = pickRandom(themes, 4)
    visibleFonts = pickRandom(fonts, 4)
    rotationTimer = setInterval(refreshPreview, 5 * 60 * 1000)

    EventsOn('menu:file:new', newFile)
    EventsOn('menu:file:open', openFile)
    EventsOn('menu:file:save', saveFile)
    EventsOn('menu:file:save-as', saveFileAs)
    EventsOn('menu:settings:configure-api-key', () => { showApiKeyModal = true })

    // Layout menu events
    const themeIDs = themes.map((t) => t.id)
    for (const id of themeIDs) {
      EventsOn(`menu:layout:theme:${id}`, () => onLayoutThemeEvent(id))
    }
    const fontIDs = fonts.map((f) => f.id)
    for (const id of fontIDs) {
      EventsOn(`menu:layout:font:${id}`, () => onLayoutFontEvent(id))
    }

    const sizeValues = ['small', 'normal', 'large', 'huge']
    for (const size of sizeValues) {
      EventsOn(`menu:layout:font-size:${size}`, () => selectFontSize(size))
    }

    const spacingValues = ['tight', 'comfortable', 'relaxed']
    for (const spacingVal of spacingValues) {
      EventsOn(`menu:layout:spacing:${spacingVal}`, () => selectSpacing(spacingVal))
    }

    window.addEventListener('beforeunload', () => {
      if (!isSaved && !currentPath) { autoSaveLocalDraft() }
    })

    window.addEventListener('keydown', (event) => {
      if (event.ctrlKey && event.key === ' ') {
        event.preventDefault()
        triggerParagraphAutocomplete()
      }
    })

    return () => {
      if (saveTimer) window.clearTimeout(saveTimer)
      if (autocompleteDebounceTimer) window.clearTimeout(autocompleteDebounceTimer)
      editor?.destroy()
    }
  })

  $: statusText = saveError
    ? saveError
    : isSaving ? (currentPath ? 'Saving...' : 'Saving locally...')
    : isSaved ? (currentPath ? 'Saved' : 'Saved locally')
    : 'Unsaved changes'
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,400;0,500;0,600;0,700;1,500&family=DM+Sans:opsz,wght@9..40,400;9..40,500;9..40,600;9..40,700&family=Space+Mono:ital,wght@0,400;0,700;1,400&family=Playfair+Display:ital,wght@0,400;0,700;1,500&family=Inter:opsz,wght@14..32,400;14..32,600;14..32,700&family=Merriweather:wght@0,400;0,700;1,400&family=Monoton&family=Bebas+Neue&display=swap" rel="stylesheet" />
</svelte:head>

{#if showApiKeyModal}
  <ApiKeyModal on:close={() => (showApiKeyModal = false)} on:keyschanged={refreshApiKeyStatus} />
{/if}

<main class={`app-shell theme-${theme} font-${font} font-size-${fontSize} spacing-${spacing}`}>
  <header class="topbar">
    <div class="brand" aria-label="PandaWriter">
      <span class="brand-mark">P</span>
      <span>PandaWriter</span>
    </div>
    <div class="document-status">
      <span class:unsaved={!isSaved} class="status-dot"></span>
      {statusText}
    </div>
    <button class="publish-button" type="button" disabled title="Publishing is planned for a later POC phase">Publish</button>
  </header>

  {#if hasApiKey}
    <nav class="ai-toggle-bar" aria-label="AI features">
      <span class="toggle-label">AI</span>
      <button type="button" class="toggle-pill" class:active={wordAutocompleteOn} onclick={toggleWordAutocomplete} title="Suggest words as you type">
        <span class="pill-track"></span><span class="pill-text">Word</span>
      </button>
      <button type="button" class="toggle-pill" class:active={paragraphAutocompleteOn} onclick={toggleParagraphAutocomplete} title="Ctrl+Space to complete a sentence">
        <span class="pill-track"></span><span class="pill-text">Sentence</span>
      </button>
      {#if hasUnsplashKey}
        <button type="button" class="toggle-pill" class:active={illustrationOn} onclick={toggleIllustration} title="Analyze finished paragraphs for themes & illustrations">
          <span class="pill-track"></span><span class="pill-text">Illustration</span>
        </button>
      {:else}
        <button type="button" class="toggle-pill disabled-pill" disabled title="Configure an Unsplash access key to enable illustrations">
          <span class="pill-track"></span><span class="pill-text">Illustration</span>
        </button>
      {/if}
      {#if paragraphAutocompleteOn}
        <span class="toggle-hint">Ctrl+Space to autocomplete</span>
      {/if}
    </nav>
  {/if}

  <section class="workspace">
    <aside class="side-panel" aria-label="Writing appearance">
      <div class="panel-intro">
        <p class="eyebrow">Writing room</p>
        <h2>Set the scene.</h2>
        <p>Choose a mood before you begin.</p>
      </div>

      <section class="picker-section" aria-labelledby="background-label">
        <div class="section-heading"><h3 id="background-label">Background</h3><span>{themes.find((item) => item.id === theme)?.name} · {themes.length} total</span></div>
        <div class="theme-grid" class:rotating>
          {#each visibleThemes as item (item.id)}
            <button class:active={theme === item.id} class={`theme-card preview-${item.id}`} type="button" aria-pressed={theme === item.id} onclick={() => selectTheme(item.id)}>
              <span class="theme-swatch"></span>
              <span class="theme-copy"><strong>{item.name}</strong><small>{item.caption}</small></span>
            </button>
          {/each}
        </div>
      </section>

      <section class="picker-section" aria-labelledby="font-label">
        <div class="section-heading"><h3 id="font-label">Type mood</h3><span>{fonts.find((item) => item.id === font)?.name} · {fonts.length} total</span></div>
        <div class="font-list" class:rotating>
          {#each visibleFonts as item (item.id)}
            <button class:active={font === item.id} class={`font-choice font-preview-${item.id}`} type="button" aria-pressed={font === item.id} onclick={() => selectFont(item.id)}>
              <strong>{item.name}</strong><span>{item.sample}</span>
            </button>
          {/each}
        </div>
      </section>

      {#if analysis}
        <section class="picker-section" aria-labelledby="suggestions-label">
          <div class="section-heading"><h3 id="suggestions-label">Suggestions</h3></div>
          <div class="suggestions">
            {#if analysis.theme}
              <div class="suggestion"><p>Theme</p><button class="button" onclick={() => selectTheme(analysis.theme!)}>{analysis.theme}</button></div>
            {/if}
            {#if analysis.font}
              <div class="suggestion"><p>Font</p><button class="button" onclick={() => selectFont(analysis.font!)}>{analysis.font}</button></div>
            {/if}
            {#if analysis.illustration}
              <div class="suggestion">
                <p>Illustration</p>
                {#if analysis.illustration.startsWith('http')}
                  <img class="illustration-image" src={analysis.illustration} alt="Scene illustration" />
                {:else}
                  <p>{analysis.illustration}</p>
                {/if}
                <button class="insert-image-button" onclick={insertImageIntoEditor} title="Insert image into the editor">↓ Insert into page</button>
              </div>
            {/if}
          </div>
        </section>
      {/if}
    </aside>

    <section class="editor-area" aria-label="Story editor">
      <div class="editor-chrome">
        <div class="story-meta">
          <span>Chapter one</span>
          <span class="meta-divider">•</span>
          <span>{wordCount} words</span>
        </div>
        <div class="toolbar" aria-label="Text formatting">
          <button type="button" aria-label="Bold" class:active={editor?.isActive('bold')} onclick={() => command((instance) => instance.chain().focus().toggleBold().run())}><strong>B</strong></button>
          <button type="button" aria-label="Italic" class:active={editor?.isActive('italic')} onclick={() => command((instance) => instance.chain().focus().toggleItalic().run())}><em>I</em></button>
          <button type="button" aria-label="Underline" class:active={editor?.isActive('underline')} onclick={() => command((instance) => instance.chain().focus().toggleUnderline().run())}><u>U</u></button>
          <span class="toolbar-divider"></span>
          <button type="button" aria-label="Heading" class:active={editor?.isActive('heading', { level: 2 })} onclick={() => command((instance) => instance.chain().focus().toggleHeading({ level: 2 }).run())}>H</button>
          <button type="button" aria-label="Bullet list" class:active={editor?.isActive('bulletList')} onclick={() => command((instance) => instance.chain().focus().toggleBulletList().run())}>≡</button>
          <button type="button" aria-label="Add link" class:active={editor?.isActive('link')} onclick={setLink}>↗</button>
          <span class="toolbar-divider"></span>
          <button type="button" aria-label="Undo" onclick={() => command((instance) => instance.chain().focus().undo().run())}>↶</button>
          <button type="button" aria-label="Redo" onclick={() => command((instance) => instance.chain().focus().redo().run())}>↷</button>
          {#if paragraphAutocompleteOn}
            <span class="toolbar-divider"></span>
            <button type="button" aria-label="Paragraph autocomplete (Ctrl+Space)" onclick={triggerParagraphAutocomplete} title="Complete sentence (Ctrl+Space)">✦</button>
          {/if}
        </div>
      </div>

      <input
        class="title-input"
        type="text"
        bind:value={title}
        placeholder="Untitled story"
        aria-label="Story title"
        oninput={markChanged}
      />

      {#if wordAutocompleteOn && suggestionsPopup.visible}
        <div class="autocomplete-popup">
          {#each suggestionsPopup.words as word}
            <button type="button" class="completion-item" onclick={() => acceptSuggestion(word)}>{word}</button>
          {/each}
          <button type="button" class="completion-dismiss" onclick={dismissSuggestions}>× dismiss</button>
        </div>
      {/if}

      <article class="page" aria-label="Editable story">
        <div bind:this={editorElement}></div>
      </article>
    </section>
  </section>
</main>

<style>
  .suggestions { display: flex; flex-direction: column; gap: 1rem; }
  .suggestion { display: flex; flex-direction: column; gap: 0.5rem; }
  .suggestion p { margin: 0; }
  .illustration-image { width: 100%; border-radius: 8px; border: 1px solid var(--line); object-fit: cover; max-height: 180px; }
  .insert-image-button { padding: 6px 12px; border: 1px solid var(--accent); border-radius: 6px; background: transparent; color: var(--accent); cursor: pointer; font-size: 11px; font-weight: 600; transition: all 0.15s ease; }
  .insert-image-button:hover { background: var(--accent); color: var(--accent-ink); }

  .title-input {
    display: block;
    max-width: 825px;
    margin: 0 auto 10px;
    padding: 10px 16px;
    font-family: var(--heading-font);
    font-size: 28px;
    font-weight: 700;
    color: var(--ink);
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    outline: none;
    transition: border-color 0.2s ease;
  }
  .title-input:focus { border-bottom-color: var(--accent); }
  .title-input::placeholder { color: var(--muted); }

  .ai-toggle-bar { display: flex; align-items: center; gap: 10px; padding: 8px 32px; background: color-mix(in srgb, var(--surface) 50%, transparent); border-bottom: 1px solid var(--line); font-size: 12px; }
  .toggle-label { font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.1em; margin-right: 4px; font-size: 11px; }
  .toggle-pill { display: inline-flex; align-items: center; gap: 7px; padding: 5px 12px; border: 1px solid var(--line); border-radius: 999px; background: var(--control); color: var(--muted); cursor: pointer; transition: all 0.2s ease; font-size: 12px; }
  .toggle-pill:hover { border-color: var(--accent); color: var(--ink); }
  .toggle-pill.active { background: var(--accent); color: var(--accent-ink); border-color: var(--accent); }
  .disabled-pill { opacity: 0.35; cursor: not-allowed; }
  .pill-track { width: 10px; height: 10px; border-radius: 50%; background: currentColor; opacity: 0.3; transition: opacity 0.2s ease; flex-shrink: 0; }
  .toggle-pill.active .pill-track { opacity: 1; background: var(--accent-ink); }
  .pill-text { font-weight: 600; }
  .toggle-hint { margin-left: 12px; color: var(--muted); font-size: 11px; font-style: italic; }
  .autocomplete-popup { display: flex; align-items: center; gap: 6px; max-width: 825px; margin: 0 auto 8px; padding: 6px 10px; background: var(--surface); border: 1px solid var(--accent); border-radius: 10px; box-shadow: 0 6px 18px rgba(0,0,0,0.12); flex-wrap: wrap; }
  .completion-item { padding: 4px 12px; border: 1px solid var(--line); border-radius: 999px; background: var(--control); color: var(--ink); cursor: pointer; font-size: 13px; font-family: var(--story-font); transition: all 0.15s ease; }
  .completion-item:hover { background: var(--accent); color: var(--accent-ink); border-color: var(--accent); }
  .completion-dismiss { margin-left: auto; padding: 4px 8px; border: none; border-radius: 6px; background: transparent; color: var(--muted); cursor: pointer; font-size: 11px; transition: color 0.15s ease; }
  .completion-dismiss:hover { color: var(--ink); }

  .theme-grid.rotating, .font-list.rotating {
    opacity: 0.5;
    pointer-events: none;
    transition: opacity 0.15s ease;
  }
</style>
