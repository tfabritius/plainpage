<script setup lang="ts">
defineProps<{
  disabled?: boolean
  icon?: string
  label?: string
  loading?: boolean
  size?: 'small' | 'large'
  text?: boolean
  type?: 'success' | 'primary' | 'danger'
  onClick?: (e: MouseEvent) => any
}>()
</script>

<template>
  <ElButton
    :disabled="disabled"
    :loading="loading"
    :size="size"
    :text="text"
    :type="type"
    @click="onClick"
  >
    <Icon v-if="icon" :name="icon" />
    <span v-if="icon && label" class="hidden md:inline ml-1">{{ label }}</span>
    <span v-else-if="label">{{ label }}</span>
  </ElButton>
</template>
