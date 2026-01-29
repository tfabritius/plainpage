<script setup lang="ts">
import type { MdEditorGenerator, Segment } from '~/types/'
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core'
import { marked } from 'marked'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', payload: typeof props.modelValue): void
  (e: 'escape'): void
}>()

const markdown = computed({
  get: () => props.modelValue,
  set: (value) => { emit('update:modelValue', value) },
})

const segments = ref<Segment[]>([])

function lineNoToSegment(lineNo: number): Segment | null {
  for (const s of segments.value) {
    if (s.lineStart >= lineNo) {
      return s
    }
  }
  return null
}

// Splits markdown into segments, adds tokens of type space to previous segment
function splitMdToSegments(markdown: string): Segment[] {
  const tokens = marked.lexer(markdown)

  const segments: Segment[] = []

  let currentLine = 1
  let currentSegment: Segment = {
    idx: 0,
    lineStart: 0,
    lineEnd: 0,
    tokens: [],
  }
  let idx = 0

  for (const token of tokens) {
    const tokenLines = token.raw ? token.raw.split('\n').length - 1 : 0

    if (token.type !== 'space') {
      // Start a new segment

      // Add latest segment (unless it's empty because it is the first token)
      if (currentSegment.tokens.length > 0) {
        segments.push(currentSegment)
        idx++
      }

      // .. and start a new segment
      currentSegment = {
        idx,
        lineStart: currentLine,
        lineEnd: currentLine,
        tokens: [],
      }
    }

    // Add token to segment
    currentSegment.tokens.push(token)
    currentLine += tokenLines
    currentSegment.lineEnd = currentLine
  }

  // Handle the last segment
  if (currentSegment.tokens.length > 0) {
    segments.push(currentSegment)
  }

  return segments
}

watch(markdown, (newVal) => {
  segments.value = splitMdToSegments(newVal)
})
segments.value = splitMdToSegments(markdown.value)

const codeEditor = useTemplateRef('codeEditorRef')
const preview = useTemplateRef('previewRef')

function onEditorScroll({ firstVisibleLineNo }: { firstVisibleLineNo: number }) {
  const firstVisibleSegment = lineNoToSegment(firstVisibleLineNo)
  if (firstVisibleSegment === null) {
    throw new Error(`no segment found for line number ${firstVisibleLineNo}`)
  }

  preview.value?.scrollToSegmentIdx(firstVisibleSegment.idx)
}

function onPreviewScroll({ firstVisibleSegmentIdx }: { firstVisibleSegmentIdx: number }) {
  const segment = segments.value[firstVisibleSegmentIdx]!
  codeEditor.value?.scrollToLineNo(segment.lineStart)
}

function createWrapUnwrapGenerator(enclosingStart: string, enclosingEnd: string) {
  const generator: MdEditorGenerator = (oldText: string) => {
    let text = ''
    if (oldText.startsWith(enclosingStart) && oldText.endsWith(enclosingEnd)) {
      // Unwrap
      if (enclosingEnd.length > 0) {
        text = oldText.slice(enclosingStart.length, -enclosingEnd.length)
      } else {
        text = oldText.slice(enclosingStart.length)
      }
    } else {
      // Wrap
      text = enclosingStart + oldText + enclosingEnd
    }

    return {
      text,
      selection: oldText === '' ? { from: enclosingStart.length, to: enclosingStart.length } : undefined,
    }
  }

  return generator
}

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMdOrLarger = breakpoints.greaterOrEqual('md')
const showPreview = ref(isMdOrLarger.value)
const showFullscreen = ref(false)

interface ToolbarAction {
  type: 'selection' | 'line'
  start: string
  end: string
}

const toolbarActions: Record<string, ToolbarAction> = {
  // Selection-based formatting
  bold: { type: 'selection', start: '**', end: '**' },
  italic: { type: 'selection', start: '*', end: '*' },
  underline: { type: 'selection', start: '_', end: '_' },
  strikethrough: { type: 'selection', start: '~~', end: '~~' },
  superscript: { type: 'selection', start: '<sup>', end: '</sup>' },
  subscript: { type: 'selection', start: '<sub>', end: '</sub>' },
  link: { type: 'selection', start: '[', end: ']()' },
  'code-inline': { type: 'selection', start: '`', end: '`' },

  // Line-based formatting
  'heading-1': { type: 'line', start: '# ', end: '' },
  'heading-2': { type: 'line', start: '## ', end: '' },
  'heading-3': { type: 'line', start: '### ', end: '' },
  'heading-4': { type: 'line', start: '#### ', end: '' },
  'heading-5': { type: 'line', start: '##### ', end: '' },
  'heading-6': { type: 'line', start: '###### ', end: '' },
  'list-unordered': { type: 'line', start: '- ', end: '' },
  'list-ordered': { type: 'line', start: '1. ', end: '' },
  checkbox: { type: 'line', start: '- [ ] ', end: '' },
  quote: { type: 'line', start: '> ', end: '' },
  'code-block': { type: 'line', start: '```\n', end: '\n```' },
  table: { type: 'line', start: '|   |   |\n|---|---|\n|   |   |\n|   |   |\n', end: '' },
}

function onToolbarClick(action: string) {
  // Handle toggle actions (they don't need editor)
  if (action === 'fullscreen') {
    showFullscreen.value = !showFullscreen.value
    return
  }
  if (action === 'preview') {
    showPreview.value = !showPreview.value
    return
  }

  const editor = codeEditor.value
  if (!editor) {
    return
  }

  const config = toolbarActions[action]
  if (!config) {
    throw new Error(`Unknown toolbar action: ${action}`)
  }

  const generator = createWrapUnwrapGenerator(config.start, config.end)

  if (config.type === 'selection') {
    editor.replaceSelection(generator)
  } else {
    editor.replaceLine(generator)
  }
}

const containerClasses = computed(() => {
  if (showFullscreen.value) {
    return 'fixed inset-0 z-50 flex flex-col bg-white dark:bg-black h-full'
  }
  return 'border border-gray-300 border-solid'
})

onKeyStroke('Escape', async (_e) => {
  if (showFullscreen.value) {
    showFullscreen.value = false
  } else {
    emit('escape')
  }
})
</script>

<template>
  <div
    class="flex flex-col"
    :class="containerClasses"
  >
    <MdEditorToolbar
      :show-fullscreen="showFullscreen"
      class="border-b border-b-gray-300 border-b-solid "
      :show-preview="showPreview"
      @click="onToolbarClick"
    />
    <div class="grow flex overflow-auto">
      <div
        :class="showPreview ? 'w-1/2' : 'w-full'"
      >
        <MdCodeEditor
          ref="codeEditorRef"
          v-model="markdown"
          @scroll="onEditorScroll"
        />
      </div>
      <div
        class="w-1/2 overflow-auto border-l border-l-gray-300 border-l-solid"
        :class="showPreview ? 'w-1/2' : 'hidden'"
      >
        <MdPreview
          v-show="showPreview"
          ref="previewRef"
          :segments="segments"
          @scroll="onPreviewScroll"
        />
      </div>
    </div>
  </div>
</template>
