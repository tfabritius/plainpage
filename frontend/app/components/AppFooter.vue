<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'

const app = useAppStore()
const { version } = storeToRefs(app)
</script>

<template>
  <div class="py-2 text-sm text-gray-400 dark:text-gray-500 flex items-center justify-center">
    <ULink
      to="https://github.com/tfabritius/plainpage"
      class="text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-400 inline-flex items-center gap-1"
    >
      <UIcon name="lucide:github" class="size-4" />
      <span>PlainPage</span>
    </ULink>

    <template v-if="version">
      <span class="mx-1.5">·</span>
      <span>v{{ version }}</span>
    </template>
  </div>
</template>
