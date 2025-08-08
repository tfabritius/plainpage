<script setup lang="ts">
import type { SearchHit } from '~/types/api'
import { useRouteQuery } from '@vueuse/router'

const { t } = useI18n()

useHead(() => ({ title: t('search') }))

const q = useRouteQuery('q')

const query = ref('')
const loading = ref(false)
const results = ref<SearchHit[]>()

function readQuery() {
  query.value = Array.isArray(q.value) ? (q.value[0] ?? '') : (q.value ?? '')
}

function updateQuery() {
  q.value = query.value || undefined
}

async function onSearch() {
  query.value = query.value.trim()

  updateQuery()

  if (query.value === '') {
    results.value = undefined
    return
  }

  loading.value = true
  results.value = await apiFetch<SearchHit[]>(`/search?q=${encodeURIComponent(query.value)}`, { method: 'POST' })
  loading.value = false
}

onMounted(() => {
  readQuery()
  onSearch()
})

watch(q, () => {
  readQuery()
  onSearch()
})
</script>

<template>
  <Layout>
    <template #title>
      {{ $t('search') }}
    </template>

    <form class="flex" @submit.prevent="onSearch">
      <UInput v-model="query" :placeholder="t('search')" class="w-full" />
      <ReactiveButton color="primary" icon="ci:search" class="ml-2" :loading="loading" :label="$t('search')" type="submit" />
    </form>

    <div v-if="results !== undefined">
      <h2 class="font-light text-xl my-4">
        {{ $t('_search.results') }}
      </h2>

      <div v-if="results.length > 0">
        <div v-for="(result, i) in results" :key="i" class="mb-4">
          <ULink :to="`/${result.url}`">
            <span class="text-xl flex items-center">
              <UIcon :name="result.isFolder ? 'ci:folder' : 'ci:file-blank'" class="mr-1" />
              <span v-if="'meta.title' in result.fragments" v-html="result.fragments['meta.title'][0]" />
              <span v-else :class="{ 'font-italic': !result.meta.title }">{{ result.meta.title || 'Untitled' }}</span>
            </span>
            <span class="text-sm font-mono ml-2">
              <span v-if="'url' in result.fragments" v-html="result.fragments.url[0]" />
              <span v-else>{{ result.url }}</span>
            </span>
          </ULink>
          <br>

          <div v-if="'content' in result.fragments" class="text-gray-400 dark:text-gray-500">
            <div v-for="(f, ii) in result.fragments.content" :key="ii" v-html="f" />
          </div>

          <!-- eslint-disable vue/no-v-text-v-html-on-component -->
          <UBadge
            v-for="tag in result.fragments['meta.tags']"
            :key="tag"
            class="mr-1"
            v-html="tag"
          />
        </div>
      </div>
      <div v-else>
        {{ $t('_search.no-results') }}
      </div>
    </div>
  </Layout>
</template>
