<script setup lang="ts">
import { ScrollAreaCorner, ScrollAreaRoot, ScrollAreaScrollbar, ScrollAreaThumb, ScrollAreaViewport } from 'radix-vue'

const emit = defineEmits<{
  (e: 'scroll', position: { scrollTop: number }): void
}>()

const scrollAreaViewport = useTemplateRef('scrollAreaViewport')

onMounted(() => {
  scrollAreaViewport.value?.viewportElement?.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  scrollAreaViewport.value?.viewportElement?.removeEventListener('scroll', handleScroll)
})

function handleScroll() {
  const scrollTop = scrollAreaViewport.value?.viewportElement?.scrollTop || 0
  emit('scroll', { scrollTop })
}

function setScrollTop(position: number): void {
  scrollAreaViewport.value?.viewportElement?.scrollTo({ top: position })
}

defineExpose({ setScrollTop })
</script>

<template>
  <ScrollAreaRoot
    class="w-full h-full overflow-hidden relative"
    type="auto"
  >
    <ScrollAreaViewport
      ref="scrollAreaViewport"
      class="w-full h-full"
    >
      <slot />
    </ScrollAreaViewport>
    <ScrollAreaScrollbar
      orientation="vertical"
      class="flex select-none touch-none p-0.5 w-2.5"
    >
      <ScrollAreaThumb
        class="flex-1 bg-neutral-200 hover:bg-neutral-300 transition-colors duration-[160ms] ease-out rounded-[10px] relative"
      />
    </ScrollAreaScrollbar>
    <ScrollAreaCorner />
  </ScrollAreaRoot>
</template>
