<script setup lang="ts">
import { minimalSetup } from 'codemirror'
import { EditorView, highlightActiveLine } from '@codemirror/view'
import { Compartment, EditorSelection } from '@codemirror/state'
import { markdown } from '@codemirror/lang-markdown'
import { oneDark } from '@codemirror/theme-one-dark'

import type { MdEditorGenerator } from '~/types'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', payload: typeof props.modelValue): void
  (e: 'scroll', payload: { firstVisibleLineNo: number }): void
}>()

const doc = computed({
  get: () => props.modelValue,
  set: (value) => { emit('update:modelValue', value) },
})

// Number of pixels of line that have to be visible to count as (first) visible line
const firstVisibleLineOffsetDetect = 5

// Number of pixels above (first) visible line when scrolling programmatically
const firstVisibleLineOffsetScrollTo = 1

// Codemirror EditorView instance ref
const editorView = shallowRef<EditorView>()

function handleReady(payload: { view: EditorView }) {
  editorView.value = payload.view
}

function getFirstVisibleLineNo(view: EditorView) {
  // Editor position on document
  const editorRect = view.dom.getBoundingClientRect()

  // First visible line block
  const firstVisibleLineBlock = view.lineBlockAtHeight(editorRect.top - view.documentTop + firstVisibleLineOffsetDetect)

  // Convert position in text (.from) to line
  const line = view.state.doc.lineAt(firstVisibleLineBlock.from)

  return line.number
}

// Is set to true while editor scrolles programmatically
const ignoreScrollEvent = ref(false)

function onScroll() {
  if (ignoreScrollEvent.value) {
    return
  }

  if (editorView.value) {
    const firstVisibleLineNo = getFirstVisibleLineNo(editorView.value)

    emit('scroll', { firstVisibleLineNo })
  }
}

const scrollTimeoutId = ref<number>()

function scrollToLineNo(lineNo: number) {
  if (editorView.value) {
    window.clearTimeout(scrollTimeoutId.value)
    ignoreScrollEvent.value = true

    // Convert line number to position in text
    const pos = editorView.value.state.doc.line(lineNo).from

    // Scroll view to make this character visible at the top
    editorView.value.dispatch({
      effects: EditorView.scrollIntoView(pos, { y: 'start', yMargin: firstVisibleLineOffsetScrollTo }),
    })

    scrollTimeoutId.value = window.setTimeout(() => {
      ignoreScrollEvent.value = false
    }, 50)
  }
}

function replaceLine(generator: MdEditorGenerator) {
  const view = editorView.value
  if (!view) {
    return
  }

  const cursorPosition = view.state.selection.ranges[0].head
  const lineBlock = view.lineBlockAt(cursorPosition)
  const lineText = view.state.doc.toString().substring(lineBlock.from, lineBlock.to)

  const result = generator(lineText)

  const changes = { from: lineBlock.from, to: lineBlock.to, insert: result.text }
  view.dispatch({
    changes,
  })

  // Set focus on editor
  view.focus()
}

function replaceSelection(generator: MdEditorGenerator) {
  const view = editorView.value
  if (!view) {
    return
  }

  const firstRange = view.state.selection.ranges[0]
  const selectedText = view.state.doc.toString().substring(firstRange.from, firstRange.to)

  const result = generator(selectedText)

  const changes = { from: firstRange.from, to: firstRange.to, insert: result.text }
  const selection = result.selection
    ? EditorSelection.create([
      EditorSelection.range(
        firstRange.from + result.selection.from,
        firstRange.from + result.selection.to,
      ),
    ])
    : undefined
  view.dispatch({
    changes,
    selection,
  })

  // Set focus to editor
  view.focus()
}

defineExpose({ scrollToLineNo, replaceSelection, replaceLine })

// Compartment to switch editor theme
const editorTheme = new Compartment()

const extensions = [
  minimalSetup,

  markdown(),
  highlightActiveLine(),

  // Don't scroll horizontally
  EditorView.lineWrapping,

  // Handle scroll event
  EditorView.domEventHandlers({ scroll: onScroll }),

  // Dark mode theme
  editorTheme.of(oneDark),
]

const isDark = useDark()

function updateTheme() {
  const view = editorView.value
  if (!view) {
    return
  }

  const theme = isDark.value ? oneDark : []

  view.dispatch({ effects: editorTheme.reconfigure(theme) })
}
onMounted(() => updateTheme())
watch(isDark, updateTheme)
</script>

<template>
  <Codemirror
    v-model="doc"
    placeholder="Empty"
    :style="{ height: '100%' }"
    :autofocus="true"
    :indent-with-tab="true"
    :tab-size="2"
    :extensions="extensions"
    @ready="handleReady"
  />
</template>

<style>
/* Remove dotted outline when focussed */
.cm-editor.cm-focused { outline: none }

.cm-scroller::-webkit-scrollbar {
  /* 6px thumb + 2px border on each side */
  width: 10px;
}
.cm-scroller::-webkit-scrollbar-track {
   /* Set the background of the track to transparent */
  background: transparent;
}

.cm-scroller::-webkit-scrollbar-thumb {
  background: #DDDEE0;
  border-radius: 6px;

  /* Transparent border */
  border: 2px solid transparent;

  /* Don't show background color under the border */
  background-clip: content-box;
}

.cm-scroller::-webkit-scrollbar-thumb:hover {
  background: #C7C9CC;

  border: 2px solid transparent;
  background-clip: content-box;
}
</style>
