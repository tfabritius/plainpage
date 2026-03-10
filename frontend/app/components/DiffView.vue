<script setup lang="ts">
import { diffLines } from 'diff'

const props = defineProps<{
  oldText: string
  newText: string
  oldLabel?: string
  newLabel?: string
}>()

const mode = defineModel<'side-by-side' | 'unified'>('mode', { default: 'side-by-side' })

const { t } = useI18n()

interface DiffLine {
  type: 'added' | 'removed' | 'unchanged'
  content: string
  oldLineNum?: number
  newLineNum?: number
}

const diffResult = computed(() => {
  const changes = diffLines(props.oldText, props.newText)
  const lines: DiffLine[] = []
  let oldLineNum = 1
  let newLineNum = 1

  for (const change of changes) {
    const changeLines = change.value.split('\n')
    // Remove last empty string if the change ends with newline
    if (changeLines.at(-1) === '') {
      changeLines.pop()
    }

    for (const line of changeLines) {
      if (change.added) {
        lines.push({
          type: 'added',
          content: line,
          newLineNum: newLineNum++,
        })
      } else if (change.removed) {
        lines.push({
          type: 'removed',
          content: line,
          oldLineNum: oldLineNum++,
        })
      } else {
        lines.push({
          type: 'unchanged',
          content: line,
          oldLineNum: oldLineNum++,
          newLineNum: newLineNum++,
        })
      }
    }
  }

  return lines
})

// For side-by-side view, we need to pair up lines
interface SideBySidePair {
  left: DiffLine | null
  right: DiffLine | null
}

const sideBySideLines = computed(() => {
  const pairs: SideBySidePair[] = []
  const lines = diffResult.value
  let i = 0

  while (i < lines.length) {
    const line = lines[i]
    if (!line) {
      i++
      continue
    }

    if (line.type === 'unchanged') {
      pairs.push({ left: line, right: line })
      i++
    } else if (line.type === 'removed') {
      // Look ahead for added lines to pair with
      const removedLines: DiffLine[] = []
      while (i < lines.length && lines[i]?.type === 'removed') {
        const removedLine = lines[i]
        if (removedLine) {
          removedLines.push(removedLine)
        }
        i++
      }
      const addedLines: DiffLine[] = []
      while (i < lines.length && lines[i]?.type === 'added') {
        const addedLine = lines[i]
        if (addedLine) {
          addedLines.push(addedLine)
        }
        i++
      }

      // Pair them up
      const maxLen = Math.max(removedLines.length, addedLines.length)
      for (let j = 0; j < maxLen; j++) {
        pairs.push({
          left: removedLines[j] ?? null,
          right: addedLines[j] ?? null,
        })
      }
    } else if (line.type === 'added') {
      // Added without preceding removed
      pairs.push({ left: null, right: line })
      i++
    }
  }

  return pairs
})

const stats = computed(() => {
  let additions = 0
  let deletions = 0
  for (const line of diffResult.value) {
    if (line.type === 'added') {
      additions++
    }
    if (line.type === 'removed') {
      deletions++
    }
  }
  return { additions, deletions }
})
</script>

<template>
  <div class="diff-view">
    <!-- Toolbar -->
    <div class="flex items-center justify-between mb-4 gap-4 flex-wrap">
      <div class="flex items-center gap-4 text-sm">
        <span class="text-green-600 dark:text-green-400">
          +{{ stats.additions }} {{ t('diff.additions') }}
        </span>
        <span class="text-red-600 dark:text-red-400">
          -{{ stats.deletions }} {{ t('diff.deletions') }}
        </span>
      </div>
      <div class="flex items-center gap-2">
        <UFieldGroup>
          <UButton
            :variant="mode === 'side-by-side' ? 'subtle' : 'outline'"
            size="sm"
            icon="tabler:columns-2"
            @click="mode = 'side-by-side'"
          >
            {{ t('diff.side-by-side') }}
          </UButton>
          <UButton
            :variant="mode === 'unified' ? 'subtle' : 'outline'"
            size="sm"
            icon="tabler:align-left"
            @click="mode = 'unified'"
          >
            {{ t('diff.unified') }}
          </UButton>
        </UFieldGroup>
      </div>
    </div>

    <!-- Side-by-side view -->
    <div v-if="mode === 'side-by-side'" class="diff-container border border-default rounded-lg overflow-hidden">
      <!-- Headers -->
      <div class="grid grid-cols-2 bg-muted border-b border-default">
        <div class="px-3 py-2 font-medium text-sm border-r border-default">
          {{ oldLabel || t('diff.old-version') }}
        </div>
        <div class="px-3 py-2 font-medium text-sm">
          {{ newLabel || t('diff.new-version') }}
        </div>
      </div>

      <!-- Content -->
      <div class="overflow-x-auto">
        <div class="grid grid-cols-2 font-mono text-sm">
          <template v-for="(pair, idx) in sideBySideLines" :key="idx">
            <!-- Left (old) -->
            <div
              class="flex border-b border-r border-default"
              :class="{
                'bg-red-100 dark:bg-red-900/40': pair.left?.type === 'removed',
                'bg-default': pair.left?.type === 'unchanged' || !pair.left,
              }"
            >
              <span
                class="w-12 shrink-0 px-2 py-0.5 text-right text-muted select-none border-r border-default"
              >
                {{ pair.left?.oldLineNum ?? '' }}
              </span>
              <span class="w-6 shrink-0 px-1 py-0.5 text-center select-none">
                <template v-if="pair.left?.type === 'removed'">-</template>
              </span>
              <pre class="flex-1 px-2 py-0.5 whitespace-pre-wrap break-all">{{ pair.left?.content ?? '' }}</pre>
            </div>

            <!-- Right (new) -->
            <div
              class="flex border-b border-default"
              :class="{
                'bg-green-100 dark:bg-green-900/40': pair.right?.type === 'added',
                'bg-default': pair.right?.type === 'unchanged' || !pair.right,
              }"
            >
              <span
                class="w-12 shrink-0 px-2 py-0.5 text-right text-muted select-none border-r border-default"
              >
                {{ pair.right?.newLineNum ?? '' }}
              </span>
              <span class="w-6 shrink-0 px-1 py-0.5 text-center select-none">
                <template v-if="pair.right?.type === 'added'">+</template>
              </span>
              <pre class="flex-1 px-2 py-0.5 whitespace-pre-wrap break-all">{{ pair.right?.content ?? '' }}</pre>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Unified view -->
    <div v-else class="diff-container border border-default rounded-lg overflow-hidden">
      <!-- Header -->
      <div class="bg-muted border-b border-default px-3 py-2">
        <div class="font-medium text-sm">
          {{ oldLabel || t('diff.old-version') }} → {{ newLabel || t('diff.new-version') }}
        </div>
      </div>

      <!-- Content -->
      <div class="overflow-x-auto">
        <div class="font-mono text-sm">
          <div
            v-for="(line, idx) in diffResult"
            :key="idx"
            class="flex border-b border-default last:border-b-0"
            :class="{
              'bg-red-100 dark:bg-red-900/40': line.type === 'removed',
              'bg-green-100 dark:bg-green-900/40': line.type === 'added',
              'bg-default': line.type === 'unchanged',
            }"
          >
            <span
              class="w-12 shrink-0 px-2 py-0.5 text-right text-muted select-none border-r border-default"
            >
              {{ line.oldLineNum ?? '' }}
            </span>
            <span
              class="w-12 shrink-0 px-2 py-0.5 text-right text-muted select-none border-r border-default"
            >
              {{ line.newLineNum ?? '' }}
            </span>
            <span
              class="w-6 shrink-0 px-1 py-0.5 text-center select-none"
              :class="{
                'text-red-600 dark:text-red-400': line.type === 'removed',
                'text-green-600 dark:text-green-400': line.type === 'added',
              }"
            >
              <template v-if="line.type === 'removed'">-</template>
              <template v-else-if="line.type === 'added'">+</template>
            </span>
            <pre class="flex-1 px-2 py-0.5 whitespace-pre-wrap break-all">{{ line.content }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
