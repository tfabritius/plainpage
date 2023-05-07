<script setup lang="ts">
import { marked } from 'marked'
import dompurify from 'dompurify'
import { ElScrollbar } from 'element-plus'

import type { Segment } from '~/types/'

const props = defineProps<{
  segments: Segment[]
}>()

const emit = defineEmits<{
  (e: 'scroll', payload: { firstVisibleSegmentIdx: number }): void
}>()

const segments = computed(() => props.segments)

const renderer = new marked.Renderer()
renderer.link = (href: string, title: string, text: string) => `<a title="${title ?? ''}" href="${href}" target="_blank">${text}</a>`

const renderSegmentsToHtml = (segments: Segment[]): string => {
  return segments.map((segment) => {
    const tokens = segment.tokens
    const content = marked.parser(tokens, { gfm: true, renderer })
    const sanitizedContent = dompurify.sanitize(content)
    return `<div class="segment" data-segment="${segment.idx}">${sanitizedContent}</div>`
  }).join('')
}

const html = computed(() => renderSegmentsToHtml(segments.value))

const previewArea = ref<InstanceType<typeof HTMLDivElement>>()

useMutationObserver(previewArea, () => {
  updatePositionsOfPreviewSegments()
}, { childList: true })

onMounted(() => {
  updatePositionsOfPreviewSegments()
})

const segmentPositions = ref<{ top: number; height: number }[]>([])

function updatePositionsOfPreviewSegments() {
  segmentPositions.value = []

  for (const segmentNode of previewArea.value!.childNodes) {
    if (segmentNode instanceof HTMLDivElement) {
      const idx = Number(segmentNode.dataset.segment)

      segmentPositions.value[idx] = { top: segmentNode.offsetTop, height: segmentNode.offsetHeight }
    }
  }
}

const positionToSegmentIdx = (pos: number) => {
  for (let i = 0; i < segmentPositions.value.length; i++) {
    if (segmentPositions.value[i].top + segmentPositions.value[i].height >= pos) {
      return i
    }
  }
  return null
}

// Is set to true while preview area scrolles programmatically
const ignoreScrollEvent = ref(false)

const onScroll = ({ scrollTop }: { scrollTop: number }) => {
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

const previewScrollbar = ref<InstanceType<typeof ElScrollbar>>()

const scrollTimeoutId = ref<number>()

const scrollToSegmentIdx = (idx: number) => {
  const scrollTo = segmentPositions.value[idx].top
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
  <ElScrollbar ref="previewScrollbar" height="100%" :always="true" @scroll="onScroll">
    <div ref="previewArea" v-html="html" />
  </ElScrollbar>
</template>
