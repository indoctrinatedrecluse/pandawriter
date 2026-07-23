import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'

export type ParagraphCompletedDetail = {
  paragraph: string
  /** The document position just after the completed paragraph (where Enter was pressed) */
  afterPos: number
}

export type ParagraphHandlerOptions = {
  onParagraphCompleted: (detail: ParagraphCompletedDetail) => void
}

export const ParagraphHandler = Extension.create<ParagraphHandlerOptions>({
  name: 'paragraphHandler',

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: new PluginKey('paragraphHandler'),
        props: {
          handleKeyDown: (view, event) => {
            if (event.key !== 'Enter') {
              return false
            }

            const { state } = view
            const { selection } = state
            const { $from } = selection

            const parent = $from.parent
            if (parent.type.name !== 'paragraph') {
              return false
            }

            // Check if the cursor is at the end of the paragraph
            if ($from.parentOffset < parent.content.size) {
              return false
            }

            const paragraphText = parent.textContent.trim()
            if (paragraphText) {
              const afterPos = $from.pos
              this.options.onParagraphCompleted({ paragraph: paragraphText, afterPos })
            }

            return false
          }
        }
      })
    ]
  }
})
