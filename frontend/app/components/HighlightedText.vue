<script setup lang="ts">
/**
 * Safely renders text with <mark> highlights, preventing XSS.
 * Parses <mark> tags and renders all other content as escaped text.
 */

interface HighlightPart {
  text: string
  highlight: boolean
}

const props = defineProps<{
  fragment: string
}>()

/**
 * Splits fragment by <mark> tags into safe parts.
 * All text outside <mark> is rendered as escaped text.
 */
function parseHighlightedFragment(fragment: string): HighlightPart[] {
  const parts: HighlightPart[] = []

  // Split by <mark> and </mark> tags, keeping track of position
  // eslint-disable-next-line e18e/prefer-static-regex
  const tokens = fragment.split(/(<mark>|<\/mark>)/)
  let inHighlight = false

  for (const token of tokens) {
    if (token === '<mark>') {
      inHighlight = true
      continue
    }
    if (token === '</mark>') {
      inHighlight = false
      continue
    }
    if (token) {
      parts.push({ text: token, highlight: inHighlight })
    }
  }

  return parts
}

const parts = computed(() => parseHighlightedFragment(props.fragment))
</script>

<template>
  <span>
    <span v-for="(part, i) in parts" :key="i">
      <mark v-if="part.highlight">{{ part.text }}</mark>
      <span v-else>{{ part.text }}</span>
    </span>
  </span>
</template>
