import { Mark, markInputRule, markPasteRule, mergeAttributes } from '@tiptap/core'

export interface WordErrorOptions {
  HTMLAttributes: Record<string, any>
}

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    wordError: {
      /**
       * Set a word error mark
       */
      setWordError: (attributes?: { incorrect?: string }) => ReturnType
      /**
       * Toggle a word error mark
       */
      toggleWordError: (attributes?: { incorrect?: string }) => ReturnType
      /**
       * Unset a word error mark
       */
      unsetWordError: () => ReturnType
    }
  }
}

export const WordError = Mark.create<WordErrorOptions>({
  name: 'wordError',

  addOptions() {
    return {
      HTMLAttributes: {}
    }
  },

  addAttributes() {
    return {
      incorrect: {
        default: null,
        parseHTML: (element) => element.getAttribute('data-incorrect'),
        renderHTML: (attributes) => {
          if (!attributes.incorrect) {
            return {}
          }

          return {
            'data-incorrect': attributes.incorrect
          }
        }
      }
    }
  },

  parseHTML() {
    return [
      {
        tag: 'span[data-word-error]'
      }
    ]
  },

  renderHTML({ HTMLAttributes }) {
    return ['span', mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, { 'data-word-error': '' }), 0]
  },

  addCommands() {
    return {
      setWordError: (attributes) => ({ commands }) => {
        return commands.setMark(this.name, attributes)
      },
      toggleWordError: (attributes) => ({ commands }) => {
        return commands.toggleMark(this.name, attributes)
      },
      unsetWordError: () => ({ commands }) => {
        return commands.unsetMark(this.name)
      }
    }
  }
})
