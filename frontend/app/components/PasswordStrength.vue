<script setup lang="ts">
import { computed } from 'vue'
import zxcvbn from 'zxcvbn-typescript'

const props = defineProps<{
  password?: string
}>()

const score = computed(() => {
  const password = props.password ?? ''

  if (password.length === 0) {
    return 0
  }

  return zxcvbn(password).score + 1
})

const color = computed(() => {
  switch (score.value) {
    case 5:
      return 'success'
    case 4:
      return 'success'
    case 3:
      return 'warning'
    case 2:
      return 'warning'
    case 1:
      return 'error'
    default:
      return 'error'
  }
})
</script>

<template>
  <UProgress
    :color="color"
    :model-value="score"
    :max="5"
    size="sm"
  />
</template>
