<script setup lang="ts">
import type { SearchHit, SearchResponse } from '~/types/api'
import { useRouteQuery } from '@vueuse/router'

const { t } = useI18n()

// Helper to check if a tag was matched in search (appears in fragments)
function isMatchedTag(result: SearchHit, tag: string): boolean {
  const matchedTags = result.fragments['meta.tags']
  if (!matchedTags) {
    return false
  }
  // Strip HTML tags from fragments to compare with plain tag
  return matchedTags.some(fragment => fragment.replace(/<[^>]*>/g, '') === tag)
}

useHead(() => ({ title: t('search') }))

const q = useRouteQuery('q')
const pageQuery = useRouteQuery('page')

const query = ref('')
const loading = ref(false)
const results = ref<SearchHit[]>()
const currentPage = ref(1)
const hasMore = ref(false)
const limit = 20

function readQuery() {
  query.value = Array.isArray(q.value) ? (q.value[0] ?? '') : (q.value ?? '')
  const pageStr = Array.isArray(pageQuery.value) ? (pageQuery.value[0] ?? '1') : (pageQuery.value ?? '1')
  const parsedPage = Number.parseInt(pageStr, 10)
  currentPage.value = Number.isNaN(parsedPage) || parsedPage < 1 ? 1 : parsedPage
}

function updateQuery() {
  q.value = query.value || undefined
  pageQuery.value = currentPage.value > 1 ? String(currentPage.value) : undefined
}

async function onSearch(resetPage = true) {
  query.value = query.value.trim()

  if (resetPage) {
    currentPage.value = 1
  }

  updateQuery()

  if (query.value === '') {
    results.value = undefined
    hasMore.value = false
    return
  }

  loading.value = true
  const response = await apiFetch<SearchResponse>(`/search?q=${encodeURIComponent(query.value)}&page=${currentPage.value}&limit=${limit}`, { method: 'POST' })
  results.value = response.items
  hasMore.value = response.hasMore
  loading.value = false
}

function goToPage(page: number) {
  currentPage.value = page
  onSearch(false)
}

function previousPage() {
  if (currentPage.value > 1) {
    goToPage(currentPage.value - 1)
  }
}

function nextPage() {
  if (hasMore.value) {
    goToPage(currentPage.value + 1)
  }
}

onMounted(() => {
  readQuery()
  onSearch(false)
})

watch(q, () => {
  readQuery()
  onSearch(false)
})
</script>

<template>
  <Layout>
    <template #title>
      {{ $t('search') }}
    </template>

    <form class="flex gap-3" @submit.prevent="onSearch(true)">
      <UInput v-model="query" :placeholder="t('search')" class="w-full" size="lg" />
      <ReactiveButton color="primary" icon="tabler:search" variant="solid" :loading="loading" :label="$t('search')" type="submit" />
    </form>

    <div v-if="loading || results !== undefined">
      <h2 class="font-light text-xl my-4">
        {{ $t('_search.results') }}
      </h2>

      <div v-if="loading" class="space-y-4">
        <UCard v-for="i in 3" :key="i">
          <USkeleton class="h-7 w-1/3" />
          <USkeleton class="h-4 w-1/4 mt-2" />
          <USkeleton class="h-4 w-full mt-4" />
          <USkeleton class="h-4 w-3/4 mt-1" />
        </UCard>
      </div>

      <div v-else-if="results && results.length > 0" class="space-y-4">
        <NuxtLink v-for="(result, i) in results" :key="i" :to="`/${result.url}`" class="block">
          <UCard class="cursor-pointer transition-all duration-200 hover:ring-[var(--ui-primary)]/50 hover:bg-[var(--ui-bg-elevated)]/50">
            <div class="text-xl flex items-center">
              <UIcon :name="result.isFolder ? 'tabler:folder' : 'tabler:file-text'" class="mr-1" />
              <span v-if="'meta.title' in result.fragments" v-html="result.fragments['meta.title'][0]" />
              <span v-else :class="{ 'font-italic': !result.meta.title }">{{ result.meta.title || 'Untitled' }}</span>
            </div>
            <div class="text-sm font-mono text-[var(--ui-text-muted)]">
              <span v-if="'url' in result.fragments" v-html="result.fragments.url[0]" />
              <span v-else>{{ result.url }}</span>
            </div>

            <div v-if="'content' in result.fragments" class="text-[var(--ui-text-muted)] mt-2">
              <div v-for="(f, ii) in result.fragments.content" :key="ii" v-html="f" />
            </div>

            <div v-if="result.meta.tags?.length" class="flex gap-1 flex-wrap mt-2">
              <UBadge
                v-for="tag in result.meta.tags"
                :key="tag"
                :variant="isMatchedTag(result, tag) ? 'solid' : 'outline'"
                :color="isMatchedTag(result, tag) ? 'primary' : undefined"
              >
                {{ tag }}
              </UBadge>
            </div>
          </UCard>
        </NuxtLink>

        <!-- Pagination -->
        <div class="flex justify-center items-center gap-4 mt-6">
          <UButton
            icon="tabler:chevron-left"
            :disabled="currentPage <= 1"
            variant="outline"
            @click="previousPage"
          >
            {{ $t('_search.previous') }}
          </UButton>
          <span class="text-[var(--ui-text-muted)]">
            {{ $t('_search.page', { page: currentPage }) }}
          </span>
          <UButton
            icon="tabler:chevron-right"
            trailing
            :disabled="!hasMore"
            variant="outline"
            @click="nextPage"
          >
            {{ $t('_search.next') }}
          </UButton>
        </div>
      </div>
      <div v-else-if="results && results.length === 0">
        {{ $t('_search.no-results') }}
      </div>
    </div>
  </Layout>
</template>
