import type { marked } from 'marked'

export * from './api'

export interface Segment {
  idx: number
  lineStart: number
  lineEnd: number

  tokens: marked.Token[]
}

export type MdEditorGenerator = (text: string) => {
  text: string
  selection?: {
    from: number
    to: number
  }
}
