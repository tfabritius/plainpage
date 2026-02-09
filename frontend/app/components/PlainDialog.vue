<script setup lang="ts">
import { useConfirmDialog } from '@vueuse/core'

const { t } = useI18n()

const {
  isRevealed,
  reveal,
  confirm: confirmDialog,
} = useConfirmDialog()

const parameters = ref<{
  message: string
  confirmButtonText?: string
  confirmButtonColor?: 'success' | 'warning' | 'error'
  cancelButtonText?: string
}>({ message: '' })

async function confirm(message: string, params: {
  confirmButtonText?: string
  confirmButtonColor?: 'success' | 'warning' | 'error'
  cancelButtonText?: string
} = {}): Promise<boolean> {
  parameters.value = {
    message,
    ...params,
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
      <template #body>
        <p>{{ parameters.message }}</p>
      </template>

      <template #footer>
        <UButton :label="parameters.cancelButtonText || t('cancel')" @click="confirmDialog(false)" />
        <UButton :color="parameters.confirmButtonColor || 'primary'" variant="solid" :label="parameters.confirmButtonText || t('ok')" @click="confirmDialog(true)" />
      </template>
    </UModal>
  </Teleport>
</template>
