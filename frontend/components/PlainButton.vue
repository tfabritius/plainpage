<script setup lang="ts">
const props = defineProps<{
  disabled?: boolean
  icon?: string
  form?: string
  variant?: 'link'
  label?: string
  loading?: boolean
  type?: 'submit'
  color?: 'success' | 'primary' | 'warning' | 'error'
  onClick?: (e: MouseEvent) => void | Promise<void>
}>()

const variant = computed(() => {
  if (props.variant === 'link') {
    return 'link'
  }
  if (props.color) {
    return 'solid'
  }
  return 'outline'
})
</script>

<template>
  <UButton
    :disabled="disabled"
    :loading="loading"
    :icon="icon"
    :form="form"
    :type="type"
    :color="color"
    :variant="variant"
    @click="onClick"
  >
    <span :class="icon ? 'hidden md:inline ml-1' : ''">{{ label }}</span>
  </UButton>
</template>
