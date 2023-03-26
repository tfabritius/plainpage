<script setup lang="ts">
import type { Page } from '~/types/'

const props = defineProps<{
  page: Page
}>()

const page = computed(() => props.page)
</script>

<template>
  <MdRender :markdown="page.content" />

  <Tags :model-value="page.meta.tags" :editable="false" class="mt-2" />
</template>
