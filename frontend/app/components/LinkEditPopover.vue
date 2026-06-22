<script setup lang="ts">
import type { Editor } from '@tiptap/vue-3'
// Import for type augmentation
import '@tiptap/extension-link'

const props = defineProps<{
  editor: Editor
}>()

const open = ref(false)
const url = ref('')

const active = computed(() => props.editor.isActive('link'))
const disabled = computed(() => {
  if (!props.editor.isEditable) {
    return true
  }
  const { selection } = props.editor.state
  return selection.empty && !props.editor.isActive('link')
})

// Detect if URL looks like an external domain without protocol
const looksLikeExternalUrl = computed(() => {
  if (!url.value) {
    return false
  }
  // Match patterns like: www.example.com, google.de, example.co.uk
  const domainPattern = /^(?:www\.)?[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}(?:\.[a-zA-Z]{2,})?(?:\/.*)?$/
  return domainPattern.test(url.value) && !url.value.startsWith('http://') && !url.value.startsWith('https://')
})

watch(() => props.editor, (editor, _, onCleanup) => {
  if (!editor) {
    return
  }

  const updateUrl = () => {
    const { href } = editor.getAttributes('link')
    url.value = href || ''
  }

  updateUrl()
  editor.on('selectionUpdate', updateUrl)

  onCleanup(() => {
    editor.off('selectionUpdate', updateUrl)
  })
}, { immediate: true })

function setLink() {
  if (!url.value) {
    return
  }

  const { selection } = props.editor.state
  const isEmpty = selection.empty
  const hasCode = props.editor.isActive('code')

  let chain = props.editor.chain().focus()

  if (hasCode && !isEmpty) {
    chain = chain.extendMarkRange('code').setLink({ href: url.value })
  } else {
    chain = chain.extendMarkRange('link').setLink({ href: url.value })

    if (isEmpty && !props.editor.isActive('link')) {
      chain = chain.insertContent({ type: 'text', text: url.value })
    }
  }

  chain.run()
  open.value = false
}

function removeLink() {
  props.editor
    .chain()
    .focus()
    .extendMarkRange('link')
    .unsetLink()
    .setMeta('preventAutolink', true)
    .run()

  url.value = ''
  open.value = false
}

function openLink() {
  if (!url.value) {
    return
  }
  window.open(url.value, '_blank', 'noopener,noreferrer')
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault()
    setLink()
  }
}
</script>

<template>
  <UPopover
    v-model:open="open"
    :ui="{ content: 'p-0.5' }"
  >
    <UTooltip :text="$t('editor.link')">
      <UButton
        icon="tabler:link"
        color="neutral"
        active-color="primary"
        variant="ghost"
        active-variant="soft"
        size="sm"
        :active="active"
        :disabled="disabled"
      />
    </UTooltip>

    <template #content>
      <div class="flex flex-col gap-1 p-2 min-w-80">
        <UInput
          v-model="url"
          autofocus
          name="url"
          type="url"
          variant="none"
          :placeholder="$t('url')"
          @keydown="handleKeyDown"
        >
          <div class="flex items-center mr-0.5">
            <UButton
              icon="tabler:check"
              variant="ghost"
              size="sm"
              :disabled="!url && !active"
              :title="active ? $t('save') : $t('ok')"
              @click="setLink"
            />

            <USeparator
              orientation="vertical"
              class="h-6 mx-1"
            />

            <UButton
              icon="tabler:external-link"
              color="neutral"
              variant="ghost"
              size="sm"
              :disabled="!url && !active"
              :title="$t('editor.link-open')"
              @click="openLink"
            />

            <UButton
              icon="tabler:trash"
              color="neutral"
              variant="ghost"
              size="sm"
              :disabled="!url && !active"
              :title="$t('editor.link-remove')"
              @click="removeLink"
            />
          </div>
        </UInput>

        <div v-if="looksLikeExternalUrl" class="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1 px-2">
          <Icon name="tabler:bulb" class="w-3 h-3" />
          <span>{{ $t('editor.link-hint-protocol') }}</span>
        </div>
      </div>
    </template>
  </UPopover>
</template>
