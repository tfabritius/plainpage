<script setup lang="ts">
const props = defineProps<{
  icon: string
}>()
</script>

<template>
  <button
    class="inline-flex items-center justify-center hover:text-[var(--ui-primary)] hover:bg-gray-100 dark:hover:bg-gray-800 p-1.5 text-center cursor-pointer border-0 h-7 w-7 rounded-md"
  >
    <UIcon :name="props.icon" />
  </button>
</template>
