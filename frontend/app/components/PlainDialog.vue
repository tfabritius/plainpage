<script setup lang="ts">
import { useConfirmDialog } from '@vueuse/core'

const { t } = useI18n()

const {
  isRevealed,
  reveal,
  confirm: confirmDialog,
} = useConfirmDialog()

interface ConfirmOptions {
  title?: string
  confirmButtonText?: string
  confirmButtonColor?: 'success' | 'warning' | 'error'
  cancelButtonText?: string
}

interface ConfirmState extends ConfirmOptions {
  message: string
}

const messageRef = useTemplateRef('message')

// Focus the message element to prevent buttons from receiving initial focus
watch(isRevealed, async (revealed) => {
  if (revealed) {
    await nextTick()
    messageRef.value?.focus()
  }
})

const parameters = ref<ConfirmState>({ message: '' })

async function confirm(message: string, options: ConfirmOptions = {}): Promise<boolean> {
  parameters.value = {
    message,
    ...options,
  }

  const { data, isCanceled } = await reveal()

  if (isCanceled) {
    return false
  }
  return data
}

defineExpose({ confirm })
</script>

<template>
  <Teleport to="body">
    <UModal
      :open="isRevealed"
      :close="false"
      :dismissible="true"
      @update:open="(v: Boolean) => v || confirmDialog(false)"
    >
      <template #title>
        {{ parameters.title || t('confirm') }}
      </template>

      <template #description>
        <span ref="message" tabindex="-1" class="outline-none">
          {{ parameters.message }}
        </span>
      </template>

      <template #footer>
        <UButton :label="parameters.cancelButtonText || t('cancel')" @click="confirmDialog(false)" />
        <UButton :color="parameters.confirmButtonColor || 'primary'" variant="solid" :label="parameters.confirmButtonText || t('ok')" @click="confirmDialog(true)" />
      </template>
    </UModal>
  </Teleport>
</template>
