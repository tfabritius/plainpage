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

const { data, error, refresh } = await useAsyncData(`/attic/${props.urlPath}?rev=${props.revision}`, async () => {
  try {
    const data = await apiFetch<GetContentResponse>(`/attic/${props.urlPath}?rev=${props.revision}`)
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

const revDate = computed(() => new Date(Number(props.revision) * 1000))
</script>

<template>
  <SubpageNetworkError
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
      <PlainButton
        v-if="!data.notFound"
        icon="ic:baseline-update"
        :label="$t('current-version')"
        @click="navigateTo({ query: { rev: undefined } })"
      />
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

        <PlainButton
          icon="ic:baseline-update"
          :label="$t('current-version')"
          @click="navigateTo({ query: { rev: undefined } })"
        />
      </div>
    </div>
  </Layout>
</template>
