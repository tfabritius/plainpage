<script setup lang="ts">
import type { GetAtticListResponse } from '~/types'
import { UseTimeAgo } from '@vueuse/components'
import { format } from 'date-fns'
import { timeAgoMessages } from '~/composables/timeAgoMessages'

const props = defineProps<{
  urlPath: string
  title: string
}>()

const { t } = useI18n()

useHead(() => ({ title: `${t('revisions')}: ${props.title}` }))

const { data } = await useAsyncData(`/attic/${props.urlPath}`, async () => {
  const data = await apiFetch<GetAtticListResponse>(`/attic/${props.urlPath}`)

  const entries = data.entries.map(e => ({ ...e, date: new Date(e.rev * 1000) }))
    .sort((a, b) => b.rev - a.rev)

  return {
    entries,
    breadcrumbs: data.breadcrumbs,
  }
})

// Selection state for diff comparison
// Pre-select first item (current version)
const selectedRevisions = ref<number[]>([])

// Initialize with first revision selected when data loads
watch(() => data.value?.entries, (entries) => {
  if (entries && entries.length > 0 && selectedRevisions.value.length === 0) {
    const firstEntry = entries[0]
    if (firstEntry) {
      selectedRevisions.value = [firstEntry.rev]
    }
  }
}, { immediate: true })

function toggleSelection(rev: number) {
  const idx = selectedRevisions.value.indexOf(rev)
  if (idx >= 0) {
    // Remove from selection
    selectedRevisions.value = selectedRevisions.value.filter(r => r !== rev)
  } else {
    // Add to selection (max 2)
    if (selectedRevisions.value.length >= 2) {
      // Remove the first (oldest) selection and add new one
      selectedRevisions.value = [selectedRevisions.value[1]!, rev]
    } else {
      selectedRevisions.value = [...selectedRevisions.value, rev]
    }
  }
}

function isSelected(rev: number): boolean {
  return selectedRevisions.value.includes(rev)
}

const canCompare = computed(() => selectedRevisions.value.length === 2)

const selectedArray = computed(() => selectedRevisions.value.toSorted((a, b) => a - b))

function navigateToCompare() {
  if (!canCompare.value) {
    return
  }
  const [rev1, rev2] = selectedArray.value
  navigateTo({ query: { rev: String(rev1), diff: String(rev2) } })
}

const navTo = navigateTo
</script>

<template>
  <Layout :breadcrumbs="data?.breadcrumbs">
    <template #title>
      {{ $t('old-revisions-of') }}: {{ title }}
    </template>

    <template #actions>
      <ReactiveButton
        v-if="canCompare"
        icon="tabler:git-compare"
        :label="$t('diff.compare')"
        color="primary"
        @click="navigateToCompare"
      />
      <ReactiveButton
        icon="tabler:arrow-back-up"
        :label="$t('current-version')"
        @click="navTo({ query: { rev: undefined } })"
      />
    </template>

    <p v-if="data?.entries && data.entries.length > 1" class="text-sm text-[var(--ui-text-muted)] mb-4">
      {{ $t('diff.select-two-revisions') }}
    </p>

    <div class="space-y-1">
      <div
        v-for="(el, idx) in data?.entries"
        :key="el.rev"
        class="flex items-center rounded-lg"
        :class="{ 'bg-[var(--ui-bg-muted)]/50': isSelected(el.rev) }"
      >
        <!-- Checkbox area with its own hover -->
        <div
          class="flex items-center justify-center w-12 py-2 pl-3 rounded-l-lg hover:bg-[var(--ui-bg-muted)] self-stretch"
          @click="toggleSelection(el.rev)"
        >
          <UCheckbox
            :model-value="isSelected(el.rev)"
            @click.stop
            @update:model-value="toggleSelection(el.rev)"
          />
        </div>

        <!-- Link with its own hover -->
        <ULink
          :to="`?rev=${el.rev}`"
          :active="false"
          class="flex-1 py-2 pl-2 pr-3 rounded-r-lg hover:bg-[var(--ui-bg-muted)]"
        >
          {{ format(el.date, 'yyyy-MM-dd HH:mm:ss') }}
          <UseTimeAgo v-slot="{ timeAgo }" :time="el.date" :messages="timeAgoMessages()">
            ({{ timeAgo }})
          </UseTimeAgo>
          <span v-if="idx === 0">
            <UIcon class="ml-2 align-middle" name="tabler:eye" />
            <span class="align-middle">
              ({{ $t('current-version') }})
            </span>
          </span>
        </ULink>
      </div>
    </div>
  </Layout>
</template>
