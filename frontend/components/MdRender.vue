<script setup lang="ts">
import type { Tokens } from 'marked'
import dompurify from 'dompurify'
import { marked } from 'marked'

const props = defineProps<{
  markdown: string
}>()

const renderer = new marked.Renderer()
renderer.link = ({ href, title, text }: Tokens.Link) => `<a class="markdown-link" title="${title ?? ''}" href="${href}">${text}</a>`

const html = ref('')

watchEffect(async () => {
  const parsedMarkdown = await marked.parse(props.markdown, {
    gfm: true,
    renderer,
  })
  html.value = dompurify.sanitize(parsedMarkdown)
})

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
