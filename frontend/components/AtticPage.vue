<script setup lang="ts">
import { FetchError } from 'ofetch'
import { format } from 'date-fns'
import { Icon } from '#components'

import type { GetContentResponse } from '~/types/'

const props = defineProps<{
  urlPath: string
  revision: string
}>()

const { t } = useI18n()

const urlPath = computed(() => props.urlPath)
const revision = computed(() => props.revision)

const { data, error, refresh } = await useAsyncData(`/attic${urlPath.value}?rev=${revision.value}`, async () => {
  try {
    const data = await apiFetch<GetContentResponse>(`/attic${urlPath.value}?rev=${revision.value}`)
    return {
      notFound: false,
      page: data.page,
      breadcrumbs: data.breadcrumbs,
    }
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 404) {
      return {
        notFound: true,
        page: null,
        breadcrumbs: [],
      }
    }
    throw err
  }
})

const pageTitle = computed(() => {
  if (data.value?.page) {
    return data.value.page.meta.title || t('untitled')
  }
  return t('not-found')
})

useHead(() => ({ title: pageTitle.value }))

const revDate = computed(() => new Date(Number(revision.value) * 1000))
</script>

<template>
  <NetworkError
    v-if="error || !data"
    :msg="error?.message || ''"
    :on-reload="refresh"
  />
  <Layout v-else :breadcrumbs="data.breadcrumbs">
    <template #title>
      <span v-if="data.page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">{{ pageTitle }}</span>
    </template>

    <template v-if="!data?.notFound" #subtitle>
      <Icon name="ci:clock" class="mr-1" />
      {{ format(revDate, 'yyyy-MM-dd HH:mm') }}
    </template>

    <template #actions>
      <ElButton v-if="!data.notFound" class="m-1" @click="navigateTo({ query: { rev: undefined } })">
        <Icon name="ic:baseline-update" /> <span class="hidden md:inline ml-1">{{ $t('current-version') }}</span>
      </ElButton>
    </template>

    <div v-if="data.page">
      <PageView :page="data.page" />
    </div>
    <div v-else>
      <div class="text-center">
        <span class="text-3xl">😟</span>
        <div class="font-medium m-4">
          {{ $t('revision-doesnt-exist') }}
        </div>

        <ElButton @click="navigateTo({ query: { rev: undefined } })">
          <Icon name="ic:baseline-update" /> <span class="hidden md:inline ml-1">{{ $t('current-version') }}</span>
        </ElButton>
      </div>
    </div>
  </Layout>
</template>
