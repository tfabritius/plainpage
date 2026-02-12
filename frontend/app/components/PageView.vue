<script setup lang="ts">
import type { TocItem } from '~/composables/useMarkdown'
import type { Page } from '~/types/'
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core'

const props = defineProps<{
  page: Page
}>()

const { t } = useI18n()

const toc = ref<TocItem[]>([])

// Default to expanded on desktop (md+), collapsed on mobile
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMdOrLarger = breakpoints.greaterOrEqual('md')
const tocOpen = ref(isMdOrLarger.value)

function onToc(items: TocItem[]) {
  toc.value = items
}
</script>

<template>
  <div>
    <!-- Table of Contents (collapsible, floating on desktop) -->
    <div
      v-if="toc.length > 0"
      class="mb-4 border border-[var(--ui-border)] rounded-lg bg-[var(--ui-bg-muted)] md:float-right md:ml-4 md:mb-2 md:max-w-[250px] relative z-10"
    >
      <UCollapsible v-model:open="tocOpen">
        <button class="flex items-center gap-2 w-full p-3 text-left text-sm hover:bg-[var(--ui-bg)] transition-colors rounded-lg">
          <UIcon
            name="tabler:list"
            class="size-4"
          />
          <span>{{ t('table-of-contents') }}</span>
          <UIcon
            name="tabler:chevron-down"
            class="size-4 ml-auto transition-transform"
            :class="{ 'rotate-180': tocOpen }"
          />
        </button>

        <template #content>
          <div class="p-3 pt-0">
            <TableOfContents :items="toc" />
          </div>
        </template>
      </UCollapsible>
    </div>

    <MdRender :markdown="props.page.content" @toc="onToc" />

    <Tags :model-value="props.page.meta.tags" :editable="false" class="mt-2 clear-both" />
  </div>
</template>
