<script setup lang="ts">
const props = defineProps<{
  icon: string
  tooltip?: string
}>()
</script>

<template>
  <UTooltip v-if="props.tooltip" :text="props.tooltip">
    <button
      :aria-label="props.tooltip"
      class="inline-flex items-center justify-center hover:text-primary hover:bg-gray-100 dark:hover:bg-gray-800 p-1.5 text-center cursor-pointer border-0 h-7 w-7 rounded-md"
    >
      <UIcon :name="props.icon" aria-hidden="true" />
    </button>
  </UTooltip>
  <button
    v-else
    :aria-label="props.tooltip"
    class="inline-flex items-center justify-center hover:text-primary hover:bg-gray-100 dark:hover:bg-gray-800 p-1.5 text-center cursor-pointer border-0 h-7 w-7 rounded-md"
  >
    <UIcon :name="props.icon" aria-hidden="true" />
  </button>
</template>
