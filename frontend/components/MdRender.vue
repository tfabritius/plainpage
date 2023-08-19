<script setup lang="ts">
import { marked } from 'marked'
import dompurify from 'dompurify'

const props = defineProps<{
  markdown: string
}>()

const renderer = new marked.Renderer()
renderer.link = (href: string, title: string, text: string) =>
  `<a class="markdown-link" title="${title ?? ''}" href="${href}">${text}</a>`

const html = computed(
  () => dompurify.sanitize(
    marked.parse(props.markdown, {
      gfm: true,
      renderer,
    }),
  ),
)

onMounted(() => {
  document.querySelectorAll('a.markdown-link').forEach((item) => {
    if (item instanceof HTMLAnchorElement) {
      const url = item.getAttribute('href')
      item.onclick = (e) => {
        if (e.target instanceof HTMLAnchorElement) {
          e.preventDefault()
          navigateTo(url, { external: url?.startsWith('http:') || url?.startsWith('https:') })
        }
      }
    }
  })
})
</script>

<template>
  <div class="markdown" v-html="html" />
</template>
