<script setup lang="ts">
import { marked } from 'marked'
import type { MdEditorGenerator, Segment } from '~/types/'
import { MdCodeEditor, MdPreview } from '#components'

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
    const tokenLines = token.raw.split('\n').length - 1

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

const codeEditorRef = ref<InstanceType<typeof MdCodeEditor>>()
const previewRef = ref<InstanceType<typeof MdPreview>>()

function onEditorScroll({ firstVisibleLineNo }: { firstVisibleLineNo: number }) {
  const firstVisibleSegment = lineNoToSegment(firstVisibleLineNo)
  if (firstVisibleSegment === null) {
    throw new Error(`no segment found for line number ${firstVisibleLineNo}`)
  }

  previewRef.value?.scrollToSegmentIdx(firstVisibleSegment.idx)
}

function onPreviewScroll({ firstVisibleSegmentIdx }: { firstVisibleSegmentIdx: number }) {
  const segment = segments.value[firstVisibleSegmentIdx]
  codeEditorRef.value?.scrollToLineNo(segment.lineStart)
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

const showPreview = ref(true)
const showFullscreen = ref(false)

function onToolbarClick(action: string) {
  const editor = codeEditorRef.value
  if (!editor) {
    return
  }

  switch (action) {
    case 'bold':
      editor.replaceSelection(createWrapUnwrapGenerator('**', '**'))
      break
    case 'italic':
      editor.replaceSelection(createWrapUnwrapGenerator('*', '*'))
      break
    case 'underline':
      editor.replaceSelection(createWrapUnwrapGenerator('_', '_'))
      break
    case 'strikethrough':
      editor.replaceSelection(createWrapUnwrapGenerator('~~', '~~'))
      break
    case 'superscript':
      editor.replaceSelection(createWrapUnwrapGenerator('<sup>', '</sup>'))
      break
    case 'subscript':
      editor.replaceSelection(createWrapUnwrapGenerator('<sub>', '</sub>'))
      break

    case 'heading-1':
      editor.replaceLine(createWrapUnwrapGenerator('# ', ''))
      break
    case 'heading-2':
      editor.replaceLine(createWrapUnwrapGenerator('## ', ''))
      break
    case 'heading-3':
      editor.replaceLine(createWrapUnwrapGenerator('### ', ''))
      break
    case 'heading-4':
      editor.replaceLine(createWrapUnwrapGenerator('#### ', ''))
      break
    case 'heading-5':
      editor.replaceLine(createWrapUnwrapGenerator('##### ', ''))
      break
    case 'heading-6':
      editor.replaceLine(createWrapUnwrapGenerator('###### ', ''))
      break
    case 'list-unordered':
      editor.replaceLine(createWrapUnwrapGenerator('- ', ''))
      break
    case 'list-ordered':
      editor.replaceLine(createWrapUnwrapGenerator('1. ', ''))
      break
    case 'checkbox':
      editor.replaceLine(createWrapUnwrapGenerator('- [ ] ', ''))
      break
    case 'quote':
      editor.replaceLine(createWrapUnwrapGenerator('> ', ''))
      break

    case 'link':
      editor.replaceSelection(createWrapUnwrapGenerator('[', ']()'))
      break
    case 'code-inline':
      editor.replaceSelection(createWrapUnwrapGenerator('`', '`'))
      break
    case 'code-block':
      editor.replaceLine(createWrapUnwrapGenerator('```\n', '\n```'))
      break
    case 'table':
      editor.replaceLine(createWrapUnwrapGenerator('|   |   |\n' + '|---|---|\n' + '|   |   |\n' + '|   |   |\n', ''))
      break

    case 'fullscreen':
      showFullscreen.value = !showFullscreen.value
      break
    case 'preview':
      showPreview.value = !showPreview.value
      break

    default:
      throw new Error(`Unknown toolbar action: ${action}`)
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
