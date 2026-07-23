<script lang="ts">
  import { onMount } from 'svelte'
  import { Editor } from '@tiptap/core'
  import StarterKit from '@tiptap/starter-kit'
  import Placeholder from '@tiptap/extension-placeholder'
  import Link from '@tiptap/extension-link'
  import Underline from '@tiptap/extension-underline'
  import { TextStyle } from '@tiptap/extension-text-style'
  import Color from '@tiptap/extension-color'
  import { LoadDraft, SaveDraft, OpenFile, SaveFile, SaveFileAs } from '../wailsjs/go/main/App'
  import { logDebug, logError, logInfo } from './logger'
  import { EventsOn } from '../wailsjs/runtime/runtime'

  type Theme = {
    id: string
    name: string
    caption: string
  }

  type Font = {
    id: string
    name: string
    sample: string
  }

  type Draft = {
    exists: boolean
    content: string
    theme: string
    font: string
    updatedAt: string
  }

  const themes: Theme[] = [
    { id: 'midnight', name: 'Midnight ink', caption: 'Quiet and cinematic' },
    { id: 'parchment', name: 'Soft parchment', caption: 'Warm and literary' },
    { id: 'blossom', name: 'Electric bloom', caption: 'Playful and bright' },
    { id: 'studio', name: 'Studio blue', caption: 'Focused and clean' }
  ]

  const fonts: Font[] = [
    { id: 'literary', name: 'Literary', sample: 'Cormorant Garamond' },
    { id: 'editorial', name: 'Editorial', sample: 'DM Sans' },
    { id: 'typewriter', name: 'Typewriter', sample: 'Space Mono' }
  ]

  let editorElement: HTMLDivElement
  let editor: Editor | null = null
  let theme = 'midnight'
  let font = 'literary'
  let wordCount = 0
  let isSaved = true
  let isSaving = false
  let saveError = ''
  let isReadyToPersist = false
  let contentVersion = 0
  let saveTimer: ReturnType<typeof setTimeout> | undefined
  let currentPath: string | null = null

  const starterContent = `
    <h1>Untitled story</h1>
    <p class="lede">A place for your next impossible thing.</p>
    <p>The rain had already written its silver sentence across the glass when Mira opened the envelope.</p>
    <p>She read the first line twice, then smiled as if the night had finally remembered her name.</p>
  `

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
        await SaveDraft({
          exists: true,
          content: editor.getHTML(),
          theme,
          font,
          updatedAt: ''
        })
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

  async function openFile() {
    logInfo('Opening file...')
    try {
      const [draft, path] = await OpenFile()
      if (path && editor) {
        editor.commands.setContent(draft.content)
        if (themes.some((item) => item.id === draft.theme)) theme = draft.theme
        if (fonts.some((item) => item.id === draft.font)) font = draft.font
        updateWordCount()
        currentPath = path
        isSaved = true
        saveError = ''
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
        await SaveFile(currentPath, {
          exists: true,
          content: editor.getHTML(),
          theme,
          font,
          updatedAt: new Date().toISOString()
        })
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
      const path = await SaveFileAs({
        exists: true,
        content: editor.getHTML(),
        theme,
        font,
        updatedAt: new Date().toISOString()
      })
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
        if (draft.content) editor.commands.setContent(draft.content)
        if (themes.some((item) => item.id === draft.theme)) theme = draft.theme
        if (fonts.some((item) => item.id === draft.font)) font = draft.font
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

  onMount(() => {
    editor = new Editor({
      element: editorElement,
      extensions: [
        StarterKit.configure({
          heading: { levels: [1, 2, 3] },
          link: false
        }),
        Placeholder.configure({ placeholder: 'Begin where the story starts…' }),
        Link.configure({ openOnClick: false, autolink: true }),
        Underline,
        TextStyle,
        Color
      ],
      content: starterContent,
      editorProps: {
        attributes: {
          class: 'story-prose'
        }
      },
      onUpdate: updateStats
    })

    logInfo('TipTap editor initialized')
    void restoreDraft()

    EventsOn('menu:file:open', openFile)
    EventsOn('menu:file:save', saveFile)
    EventsOn('menu:file:save-as', saveFileAs)

    window.addEventListener('beforeunload', (event) => {
      if (!isSaved && !currentPath) {
        autoSaveLocalDraft()
      }
    })

    return () => {
      if (saveTimer) window.clearTimeout(saveTimer)
      editor?.destroy()
    }
  })

  $: statusText = saveError
    ? saveError
    : isSaving
      ? currentPath ? 'Saving...' : 'Saving locally...'
      : isSaved
        ? currentPath ? 'Saved' : 'Saved locally'
        : 'Unsaved changes'
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,400;0,500;0,600;0,700;1,500&family=DM+Sans:opsz,wght@9..40,400;9..40,500;9..40,600;9..40,700&family=Space+Mono:ital,wght@0,400;0,700;1,400&display=swap" rel="stylesheet" />
</svelte:head>

<main class={`app-shell theme-${theme} font-${font}`}>
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

  <section class="workspace">
    <aside class="side-panel" aria-label="Writing appearance">
      <div class="panel-intro">
        <p class="eyebrow">Writing room</p>
        <h2>Set the scene.</h2>
        <p>Choose a mood before you begin.</p>
      </div>

      <section class="picker-section" aria-labelledby="background-label">
        <div class="section-heading">
          <h3 id="background-label">Background</h3>
          <span>{themes.find((item) => item.id === theme)?.name}</span>
        </div>
        <div class="theme-grid">
          {#each themes as item}
            <button
              class:active={theme === item.id}
              class={`theme-card preview-${item.id}`}
              type="button"
              aria-pressed={theme === item.id}
              onclick={() => selectTheme(item.id)}
            >
              <span class="theme-swatch"></span>
              <span class="theme-copy"><strong>{item.name}</strong><small>{item.caption}</small></span>
            </button>
          {/each}
        </div>
      </section>

      <section class="picker-section" aria-labelledby="font-label">
        <div class="section-heading">
          <h3 id="font-label">Type mood</h3>
          <span>{fonts.find((item) => item.id === font)?.name}</span>
        </div>
        <div class="font-list">
          {#each fonts as item}
            <button
              class:active={font === item.id}
              class={`font-choice font-preview-${item.id}`}
              type="button"
              aria-pressed={font === item.id}
              onclick={() => selectFont(item.id)}
            >
              <strong>{item.name}</strong>
              <span>{item.sample}</span>
            </button>
          {/each}
        </div>
      </section>
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
        </div>
      </div>
      <article class="page" aria-label="Editable story">
        <div bind:this={editorElement}></div>
      </article>
    </section>
  </section>
</main>