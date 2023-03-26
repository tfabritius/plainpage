<script setup lang="ts">
const props = defineProps<{
  icon: string
}>()
</script>

<template>
  <span class="inline-block hover:text-blue-500 hover:bg-gray-100 w-5 h-5 p-1 text-center cursor-pointer">
    <Icon :name="props.icon" />
  </span>
</template>
