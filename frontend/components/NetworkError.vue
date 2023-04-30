<script setup lang="ts">
const _props = defineProps<{
  msg?: string
  onReload: Function
}>()
</script>

<template>
  <div class="absolute inset-0 bg-white flex">
    <div class="m-auto text-center text-gray-500">
      <Icon name="ci:cloud-off" size="5em" />

      <div class="m-2">
        {{ msg || 'There was an error.' }}
      </div>

      <ElButton @click="onReload()">
        <Icon name="ci:arrows-reload-01" class="mr-1" />
        Try again
      </ElButton>
    </div>
  </div>
</template>
