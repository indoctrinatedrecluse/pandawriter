export interface WordError {
  incorrect: string
  correct: string
}

export interface Analysis {
  wordErrors: WordError[]
  theme: string
  font: string
  illustration: string
}

export interface Draft {
  exists: boolean
  content: string
  theme: string
  font: string
  updatedAt: string
}

export interface Theme {
  id: string
  name: string
  caption: string
}

export interface Font {
  id: string
  name: string
  sample: string
}