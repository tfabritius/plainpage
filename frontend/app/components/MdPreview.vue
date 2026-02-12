<script setup lang="ts">
import type { Tokens } from 'marked'
import type { Segment } from '~/types/'
import dompurify from 'dompurify'
import { marked } from 'marked'
import slugify from 'slugify'

const props = defineProps<{
  segments: Segment[]
}>()

const emit = defineEmits<{
  (e: 'scroll', payload: { firstVisibleSegmentIdx: number }): void
}>()

/**
 * Creates a renderer for preview with heading IDs (no anchor links)
 */
function createPreviewRenderer(slugCounter: Map<string, number>) {
  const renderer = new marked.Renderer()

  // Custom heading renderer with IDs only
  renderer.heading = ({ tokens, depth }: Tokens.Heading) => {
    const text = tokens.map(t => ('text' in t ? t.text : '')).join('')
    const tag = `h${depth}`

    // Only process H1-H4 for IDs
    if (depth <= 4) {
      let slug = slugify(text, { lower: true, strict: true })

      const count = slugCounter.get(slug) || 0
      if (count > 0) {
        slug = `${slug}-${count}`
      }
      slugCounter.set(slug.replace(/-\d+$/, ''), count + 1)

      return `<${tag} id="${slug}">${text}</${tag}>`
    }

    return `<${tag}>${text}</${tag}>`
  }

  // External links
  renderer.link = ({ href, title, text }: Tokens.Link) =>
    `<a title="${title ?? ''}" href="${href}" target="_blank">${text}</a>`

  return renderer
}

function renderSegmentsToHtml(segments: Segment[]): string {
  // Reset slug counter for each render
  const slugCounter = new Map<string, number>()
  const renderer = createPreviewRenderer(slugCounter)

  return segments.map((segment) => {
    const tokens = segment.tokens
    const content = marked.parser(tokens, { gfm: true, renderer })
    const sanitizedContent = dompurify.sanitize(content, { ADD_ATTR: ['id'] })
    return `<div class="segment" data-segment="${segment.idx}">${sanitizedContent}</div>`
  }).join('')
}

const html = computed(() => renderSegmentsToHtml(props.segments))

const previewArea = useTemplateRef('previewArea')

useMutationObserver(previewArea, () => {
  updatePositionsOfPreviewSegments()
}, { childList: true })

onMounted(() => {
  updatePositionsOfPreviewSegments()
})

const segmentPositions = ref<{ top: number, height: number }[]>([])

function updatePositionsOfPreviewSegments() {
  segmentPositions.value = []

  for (const segmentNode of previewArea.value!.childNodes) {
    if (segmentNode instanceof HTMLDivElement) {
      const idx = Number(segmentNode.dataset.segment)

      segmentPositions.value[idx] = { top: segmentNode.offsetTop, height: segmentNode.offsetHeight }
    }
  }
}

function positionToSegmentIdx(pos: number) {
  for (let i = 0; i < segmentPositions.value.length; i++) {
    if (segmentPositions.value[i]!.top + segmentPositions.value[i]!.height >= pos) {
      return i
    }
  }
  return null
}

// Is set to true while preview area scrolles programmatically
const ignoreScrollEvent = ref(false)

function onScroll({ scrollTop }: { scrollTop: number }) {
  if (ignoreScrollEvent.value) {
    return
  }

  const firstVisibleSegmentIdx = positionToSegmentIdx(scrollTop)

  if (firstVisibleSegmentIdx === null) {
    throw new Error(`segment not found at ${scrollTop}`)
  } else {
    emit('scroll', { firstVisibleSegmentIdx })
  }
}

const previewScrollbar = useTemplateRef('previewScrollbar')

const scrollTimeoutId = ref<number>()

function scrollToSegmentIdx(idx: number) {
  const scrollTo = segmentPositions.value[idx]?.top
  if (scrollTo !== undefined) {
    window.clearTimeout(scrollTimeoutId.value)
    ignoreScrollEvent.value = true

    previewScrollbar.value?.setScrollTop(scrollTo)

    scrollTimeoutId.value = window.setTimeout(() => {
      ignoreScrollEvent.value = false
    }, 50)
  }
}

defineExpose({ scrollToSegmentIdx })
</script>

<template>
  <PlainScrollbar ref="previewScrollbar" @scroll="onScroll">
    <div ref="previewArea" class="mx-1 markdown" v-html="html" />
  </PlainScrollbar>
</template>
