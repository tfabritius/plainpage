<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'

const app = useAppStore()
const { version } = storeToRefs(app)
</script>

<template>
  <div class="pt-3 text-sm text-dimmed flex items-center justify-center">
    <ULink
      to="https://github.com/tfabritius/plainpage"
      class="text-dimmed inline-flex items-center gap-1"
    >
      <UIcon name="tabler:brand-github" class="size-4" />
      <span>PlainPage</span>
    </ULink>

    <template v-if="version">
      <span class="mx-1.5">·</span>
      <span>v{{ version }}</span>
    </template>
  </div>
</template>
