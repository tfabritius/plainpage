<script setup lang="ts">
import type { Page } from '~/types/'

const props = defineProps<{
  modelValue: Page
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', payload: typeof props.modelValue): void
  (e: 'escape'): void
}>()

const page = computed({
  get: () => props.modelValue,
  set: (value) => { emit('update:modelValue', value) },
})

const { editorType, toggleEditor, isSourceEditor } = useEditorPreference()

// Visual editor ref and fullscreen state
const visualEditorRef = ref()
const showFullscreen = ref(false)

function undo() {
  visualEditorRef.value?.undo()
}

function redo() {
  visualEditorRef.value?.redo()
}

function toggleFullscreen() {
  showFullscreen.value = !showFullscreen.value
}

onKeyStroke('Escape', (e) => {
  // Only handle fullscreen for visual editor
  if (!isSourceEditor.value && showFullscreen.value) {
    e.preventDefault()
    e.stopPropagation()
    showFullscreen.value = false
    return
  }
  // For source editor, let MdEditor handle escape itself via @escape emit
  // Only emit escape from here if we're in visual editor and not in fullscreen
  if (!isSourceEditor.value) {
    emit('escape')
  }
})
</script>

<template>
  <div class="flex items-center mb-2">
    <span class="mr-2 text-sm">{{ $t('title') }}:</span>
    <UInput v-model="page.meta.title" class="grow" />

    <!-- Button group -->
    <div class="flex items-center gap-1 ml-2">
      <!-- Toggle button (existing) -->
      <ReactiveButton
        class="p-1.5"
        :icon="isSourceEditor ? 'tabler:eye' : 'tabler:code'"
        :aria-label="isSourceEditor ? $t('editor.switch-to-visual') : $t('editor.switch-to-source')"
        :label="isSourceEditor ? $t('editor.visual') : $t('editor.source')"
        @click="toggleEditor"
      />

      <!-- Undo/Redo/Fullscreen buttons (only for visual editor) -->
      <template v-if="!isSourceEditor">
        <UButton
          class="h-7 w-7 p-1 inline-flex items-center justify-center"
          :title="$t('editor.undo')"
          :disabled="!visualEditorRef?.canUndo()"
          @click="undo"
        >
          <Icon name="tabler:arrow-back-up" class="w-4 h-4" />
        </UButton>

        <UButton
          class="h-7 w-7 p-1 inline-flex items-center justify-center"
          :title="$t('editor.redo')"
          :disabled="!visualEditorRef?.canRedo()"
          @click="redo"
        >
          <Icon name="tabler:arrow-forward-up" class="w-4 h-4" />
        </UButton>

        <UButton
          class="h-7 w-7 p-1 inline-flex items-center justify-center"
          :title="showFullscreen ? $t('editor.minimize') : $t('editor.fullscreen')"
          @click="toggleFullscreen"
        >
          <Icon :name="showFullscreen ? 'tabler:minimize' : 'tabler:maximize'" class="w-4 h-4" />
        </UButton>
      </template>
    </div>
  </div>

  <MdEditor
    v-if="editorType === 'source'"
    v-model="page.content"
    class="grow min-h-0"
    @escape="emit('escape')"
  />

  <VisualEditor
    v-else
    ref="visualEditorRef"
    v-model="page.content"
    :show-fullscreen="showFullscreen"
    class="grow min-h-0"
    @escape="emit('escape')"
    @toggle-fullscreen="toggleFullscreen"
  />

  <Tags v-model="page.meta.tags" :editable="true" class="mt-2" />
</template>
