<script setup lang="ts">
import type { Page } from '~/types/'

const props = defineProps<{
  page: Page
}>()
</script>

<template>
  <MdRender :markdown="props.page.content" />

  <Tags :model-value="props.page.meta.tags" :editable="false" class="mt-2" />
</template>
