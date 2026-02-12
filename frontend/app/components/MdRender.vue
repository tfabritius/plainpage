<script setup lang="ts">
import type { TocItem } from '~/composables/useMarkdown'
import { parseMarkdown } from '~/composables/useMarkdown'

const props = defineProps<{
  markdown: string
}>()

const emit = defineEmits<{
  (e: 'toc', toc: TocItem[]): void
}>()

// tabler:link icon
// https://icon-sets.iconify.design/tabler/link/
const ANCHOR_ICON_SVG = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 15l6-6m-4-3l.463-.536a5 5 0 0 1 7.071 7.072L18 13m-5 5l-.397.534a5.07 5.07 0 0 1-7.127 0a4.97 4.97 0 0 1 0-7.071L6 11"/></svg>'

const html = ref('')

watchEffect(async () => {
  const result = await parseMarkdown(props.markdown, {
    collectToc: true,
    addAnchorLinks: true,
  })
  html.value = result.html
  emit('toc', result.toc)
})

onMounted(() => {
  setupMarkdownLinks()
  setupHeadingAnchors()
})

onUpdated(() => {
  setupMarkdownLinks()
  setupHeadingAnchors()
})

function setupHeadingAnchors() {
  // Inject icon into all heading anchors
  document.querySelectorAll('.markdown .heading-anchor').forEach((anchor) => {
    if (anchor.innerHTML === '') {
      anchor.innerHTML = ANCHOR_ICON_SVG
    }
  })
}

function setupMarkdownLinks() {
  document.querySelectorAll('.markdown a:not(.heading-anchor)').forEach((item) => {
    if (item instanceof HTMLAnchorElement) {
      const url = item.getAttribute('href')
      // Skip anchor links (internal hash links)
      if (url?.startsWith('#')) {
        item.onclick = (e) => {
          e.preventDefault()
          const targetId = url.slice(1)
          const targetElement = document.getElementById(targetId)
          if (targetElement) {
            targetElement.scrollIntoView({ behavior: 'smooth' })
            // Update URL hash without scrolling
            history.pushState(null, '', url)
          }
        }
      } else {
        item.onclick = (e) => {
          if (e.target instanceof HTMLAnchorElement) {
            e.preventDefault()
            navigateTo(url, { external: url?.startsWith('http:') || url?.startsWith('https:') })
          }
        }
      }
    }
  })
}
</script>

<template>
  <div class="markdown" v-html="html" />
</template>
