<script setup lang="ts">
import { marked } from 'marked'
import dompurify from 'dompurify'

const props = defineProps<{
  markdown: string
}>()

const renderer = new marked.Renderer()
renderer.link = (href: string, title: string, text: string) => `<a markdown-link title="${title ?? ''}" href="${href}">${text}</a>`

const html = computed(
  () => dompurify.sanitize(
    marked.parse(props.markdown, { gfm: true, renderer }),
  ),
)

onMounted(() => {
  document.querySelectorAll('a[markdown-link]').forEach((item) => {
    if (item instanceof HTMLAnchorElement) {
      item.onclick = (e) => {
        if (e.target instanceof HTMLAnchorElement) {
          e.preventDefault()
          navigateTo(e.target.getAttribute('href'), { external: true })
        }
      }
    }
  })
})
</script>

<template>
  <div v-html="html" />
</template>
