<script setup lang="ts">
const props = defineProps<{
  icon: string
}>()
</script>

<template>
  <button
    class="inline-block hover:text-blue-500 bg-inherit p-1 text-center cursor-pointer border-0 h-6 w-6"
  >
    <UIcon :name="props.icon" />
  </button>
</template>

<style scoped>
button {
  font-size: 1rem;
}
</style>
