<script setup lang="ts">
import type { EditorCustomHandlers } from '@nuxt/ui'
import type { Editor } from '@tiptap/core'
import { Subscript } from '@tiptap/extension-subscript'
import { Superscript } from '@tiptap/extension-superscript'
import { TableKit } from '@tiptap/extension-table'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import { CellSelection } from '@tiptap/pm/tables'
import LinkEditPopover from './LinkEditPopover.vue'

const props = defineProps<{
  modelValue: string
  showFullscreen?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', payload: typeof props.modelValue): void
  (e: 'escape'): void
  (e: 'toggleFullscreen'): void
}>()

const uEditorRef = ref()
const editorRef = computed(() => uEditorRef.value?.editor ?? null)

function onMarkdownUpdated(markdown: string) {
  emit('update:modelValue', markdown)
}

function handleUndo() {
  editorRef.value?.chain().focus().undo().run()
}

function handleRedo() {
  editorRef.value?.chain().focus().redo().run()
}

function canUndo() {
  return editorRef.value?.can().undo() ?? false
}

function canRedo() {
  return editorRef.value?.can().redo() ?? false
}

const containerClasses = computed(() => {
  if (props.showFullscreen) {
    return 'fixed inset-0 z-50 flex flex-col bg-white dark:bg-black h-full'
  }
  return 'border border-default rounded-lg shadow-sm overflow-hidden'
})

defineExpose({
  undo: handleUndo,
  redo: handleRedo,
  canUndo,
  canRedo,
})

const tiptapExtensions = [
  TaskList,
  TaskItem.configure({
    nested: true,
  }),
  Superscript,
  Subscript,
  TableKit,
]

const customHandlers = {
  table: {
    canExecute: (editor: Editor) => editor.can().insertTable({ rows: 3, cols: 3, withHeaderRow: true }),
    execute: (editor: Editor) => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }),
    isActive: (editor: Editor) => editor.isActive('table'),
    isDisabled: undefined,
  },
} satisfies EditorCustomHandlers

const { slashmenuItems, bubbleToolbarItems, getTableToolbarItems } = useEditorToolbar(customHandlers)

const appendToBody = import.meta.client ? () => document.body : undefined
</script>

<template>
  <div
    class="flex flex-col"
    :class="containerClasses"
  >
    <!-- Fullscreen toolbar (only shown in fullscreen mode) -->
    <div v-if="showFullscreen" class="flex items-center justify-end gap-1 p-1 border-b border-b-default bg-gray-50 dark:bg-gray-900">
      <UButton
        class="h-7 w-7 p-1 inline-flex items-center justify-center"
        :title="$t('editor.undo')"
        :disabled="!editorRef?.can().undo()"
        @click="handleUndo"
      >
        <Icon name="tabler:arrow-back-up" class="w-4 h-4" />
      </UButton>

      <UButton
        class="h-7 w-7 p-1 inline-flex items-center justify-center"
        :title="$t('editor.redo')"
        :disabled="!editorRef?.can().redo()"
        @click="handleRedo"
      >
        <Icon name="tabler:arrow-forward-up" class="w-4 h-4" />
      </UButton>

      <UButton
        class="h-7 w-7 p-1 inline-flex items-center justify-center"
        :title="$t('editor.minimize')"
        @click="emit('toggleFullscreen')"
      >
        <Icon name="tabler:minimize" class="w-4 h-4" />
      </UButton>
    </div>

    <div class="grow overflow-auto">
      <UEditor
        ref="uEditorRef"
        v-slot="{ editor }"
        content-type="markdown"
        :model-value="modelValue"
        :starter-kit="{
          link: {
            openOnClick: false,
          },
        }"
        :handlers="customHandlers"
        :extensions="tiptapExtensions"
        :placeholder="$t('editor.placeholder')"
        @update:model-value="onMarkdownUpdated"
      >
        <UEditorDragHandle :editor="editor" />

        <UEditorToolbar
          :editor="editor"
          :items="bubbleToolbarItems"
          layout="bubble"
          :should-show="({ editor, view, state }: any) => {
            if (state.selection instanceof CellSelection) {
              return false
            }
            const { selection } = state
            // Show when there's a selection OR when cursor is within a link
            return view.hasFocus() && (!selection.empty || editor.isActive('link'))
          }"
        >
          <template #link>
            <LinkEditPopover :editor="editor" />
          </template>
        </UEditorToolbar>

        <UEditorToolbar
          :editor="editor"
          :items="getTableToolbarItems(editor)"
          layout="bubble"
          :should-show="({ editor, view }: any) => {
            return editor.state.selection instanceof CellSelection && view.hasFocus()
          }"
        />

        <UEditorSuggestionMenu :editor="editor" :items="slashmenuItems" :append-to="appendToBody" />
      </UEditor>
    </div>
  </div>
</template>
