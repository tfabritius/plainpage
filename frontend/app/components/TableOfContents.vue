<script setup lang="ts">
import type { TocItem } from '~/composables/useMarkdown'

defineProps<{
  items: TocItem[]
}>()

function scrollToHeading(id: string) {
  const element = document.getElementById(id)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth' })
    // Update URL hash without triggering scroll
    history.pushState(null, '', `#${id}`)
  }
}
</script>

<template>
  <nav v-if="items.length > 0" class="text-sm leading-relaxed">
    <ul class="list-none p-0 m-0">
      <li
        v-for="item in items"
        :key="item.id"
        :class="{
          'pl-3': item.level === 2,
          'pl-6': item.level === 3,
          'pl-9': item.level === 4,
        }"
      >
        <a
          :href="`#${item.id}`"
          class="block py-1 md:py-0.5 hover:text-[var(--ui-primary)] transition-colors"
          :class="{
            'font-semibold text-[var(--ui-text)]': item.level === 1,
            'text-[var(--ui-text-muted)]': item.level > 1,
          }"
          @click.prevent="scrollToHeading(item.id)"
        >
          {{ item.text }}
        </a>
      </li>
    </ul>
  </nav>
</template>
