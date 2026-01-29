<script setup lang="ts">
import type { GetContentResponse } from '~/types/'
import { format } from 'date-fns'
import { FetchError } from 'ofetch'

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

const navTo = navigateTo
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
      <UIcon name="tabler:clock" class="mr-1" />
      {{ format(revDate, 'yyyy-MM-dd HH:mm') }}
    </template>

    <template #actions>
      <ReactiveButton
        v-if="!data.notFound"
        icon="tabler:arrow-back-up"
        :label="$t('current-version')"
        @click="navTo({ query: { rev: undefined } })"
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

        <ReactiveButton
          icon="tabler:arrow-back-up"
          :label="$t('current-version')"
          @click="navTo({ query: { rev: undefined } })"
        />
      </div>
    </div>
  </Layout>
</template>
