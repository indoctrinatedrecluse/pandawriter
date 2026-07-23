import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'

export type ParagraphHandlerOptions = {
  onParagraphCompleted: (paragraph: string) => void
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
              this.options.onParagraphCompleted(paragraphText)
            }

            return false
          }
        }
      })
    ]
  }
})
