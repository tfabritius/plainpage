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
    <span class="mr-2 text-sm">{{ $t('title') }}:</span> <UInput v-model="page.meta.title" class="w-full" />
  </div>

  <MdEditor
    v-model="page.content"
    class="grow min-h-0"
    @escape="emit('escape')"
  />

  <Tags v-model="page.meta.tags" :editable="true" class="mt-2" />
</template>
