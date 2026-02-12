<script setup lang="ts">
import type { Segment } from '~/types/'
import { parseMarkdownSegments } from '~/composables/useMarkdown'

const props = defineProps<{
  segments: Segment[]
}>()

const emit = defineEmits<{
  (e: 'scroll', payload: { firstVisibleSegmentIdx: number }): void
}>()

const html = computed(() => parseMarkdownSegments(props.segments, {
  collectToc: false,
  externalLinksNewTab: true,
}))

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
