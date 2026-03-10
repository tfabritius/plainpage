<script setup lang="ts">
defineProps<{
  legend?: string
}>()
</script>

<template>
  <fieldset class="border border-default rounded-lg p-4">
    <legend v-if="legend" class="text-sm font-medium px-2 ">
      {{ legend }}
    </legend>
    <slot />
  </fieldset>
</template>
