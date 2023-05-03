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
</script>

<template>
  <div class="flex items-baseline mb-2">
    <span class="mr-2 text-sm">Title:</span> <ElInput v-model="page.meta.title" />
  </div>

  <MdEditor
    v-model="page.content"
    height="400px"
    @escape="emit('escape')"
  />

  <Tags v-model="page.meta.tags" :editable="true" class="mt-2" />
</template>
