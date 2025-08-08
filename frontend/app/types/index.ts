import type { Token } from 'marked'

export * from './api'

export interface Segment {
  idx: number
  lineStart: number
  lineEnd: number

  tokens: Token[]
}

export type MdEditorGenerator = (text: string) => {
  text: string
  selection?: {
    from: number
    to: number
  }
}
