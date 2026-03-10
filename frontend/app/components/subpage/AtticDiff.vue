<script setup lang="ts">
import type { GetContentResponse } from '~/types/'
import { format } from 'date-fns'
import { FetchError } from 'ofetch'

const props = defineProps<{
  urlPath: string
  revision1: string
  revision2: string
}>()

const { t } = useI18n()

const diffMode = ref<'side-by-side' | 'unified'>('side-by-side')

const { data, error, refresh } = await useAsyncData(
  `/attic/${props.urlPath}?diff=${props.revision1}&${props.revision2}`,
  async () => {
    try {
      // Sort revisions so we always fetch old first, new second
      const [oldRev, newRev] = [props.revision1, props.revision2]
        .map(Number)
        .sort((a, b) => a - b)
        .map(String)

      const [oldData, newData] = await Promise.all([
        apiFetch<GetContentResponse>(`/attic/${props.urlPath}?rev=${oldRev}`),
        apiFetch<GetContentResponse>(`/attic/${props.urlPath}?rev=${newRev}`),
      ])

      return {
        notFound: false,
        oldPage: oldData.page,
        newPage: newData.page,
        oldRev,
        newRev,
        breadcrumbs: oldData.breadcrumbs,
      }
    } catch (err) {
      if (err instanceof FetchError && err.statusCode === 404) {
        return {
          notFound: true,
          oldPage: null,
          newPage: null,
          oldRev: '',
          newRev: '',
          breadcrumbs: [],
        }
      }
      throw err
    }
  },
)

const pageTitle = computed(() => {
  if (data.value?.oldPage || data.value?.newPage) {
    const page = data.value.newPage || data.value.oldPage
    return page?.meta.title || t('untitled')
  }
  return t('not-found')
})

useHead(() => ({ title: `${t('diff.compare')}: ${pageTitle.value}` }))

const oldLabel = computed(() =>
  data.value?.oldRev ? format(new Date(Number(data.value.oldRev) * 1000), 'yyyy-MM-dd HH:mm:ss') : '',
)
const newLabel = computed(() =>
  data.value?.newRev ? format(new Date(Number(data.value.newRev) * 1000), 'yyyy-MM-dd HH:mm:ss') : '',
)

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
      <span v-if="pageTitle !== t('untitled')">{{ pageTitle }}</span>
      <span v-else class="italic">{{ pageTitle }}</span>
    </template>

    <template #subtitle>
      <UIcon name="tabler:git-compare" class="mr-1" />
      {{ $t('diff.comparing-revisions') }}
    </template>

    <template #actions>
      <ReactiveButton
        icon="tabler:history"
        :label="$t('revisions')"
        @click="navTo({ query: { rev: null, diff: undefined } })"
      />
      <ReactiveButton
        v-if="!data.notFound"
        icon="tabler:arrow-back-up"
        :label="$t('current-version')"
        @click="navTo({ query: { rev: undefined } })"
      />
    </template>

    <div v-if="data.notFound" class="text-center">
      <span class="text-3xl">😟</span>
      <div class="font-medium m-4">
        {{ $t('revision-doesnt-exist') }}
      </div>

      <ReactiveButton
        icon="tabler:arrow-back-up"
        :label="$t('revisions')"
        @click="navTo({ query: { rev: null, diff: undefined } })"
      />
    </div>

    <div v-else-if="data.oldPage && data.newPage">
      <!-- Version info cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
        <!-- Old version info -->
        <div class="p-3 rounded-lg bg-muted">
          <div class="font-medium text-sm mb-1 flex items-center justify-between">
            <span>{{ oldLabel }}</span>
            <UButton
              variant="link"
              size="xs"
              icon="tabler:external-link"
              :title="$t('diff.view-revision')"
              @click="navTo({ query: { rev: data.oldRev, diff: undefined } })"
            />
          </div>
          <div v-if="data.oldPage.meta.title" class="text-sm mb-1">
            <UIcon name="tabler:file-text" class="mr-1 align-middle text-muted" />
            <span class="align-middle">{{ data.oldPage.meta.title }}</span>
          </div>
          <div v-if="data.oldPage.meta.modifiedByDisplayName" class="text-sm text-muted">
            <UIcon name="tabler:user" class="mr-1 align-middle" />
            <span class="align-middle">{{ data.oldPage.meta.modifiedByDisplayName }}</span>
          </div>
          <div v-if="data.oldPage.meta.tags?.length" class="mt-2">
            <Tags :model-value="data.oldPage.meta.tags" />
          </div>
        </div>

        <!-- New version info -->
        <div class="p-3 rounded-lg bg-muted">
          <div class="font-medium text-sm mb-1 flex items-center justify-between">
            <span>{{ newLabel }}</span>
            <UButton
              variant="link"
              size="xs"
              icon="tabler:external-link"
              :title="$t('diff.view-revision')"
              @click="navTo({ query: { rev: data.newRev, diff: undefined } })"
            />
          </div>
          <div v-if="data.newPage.meta.title" class="text-sm mb-1">
            <UIcon name="tabler:file-text" class="mr-1 align-middle text-muted" />
            <span class="align-middle">{{ data.newPage.meta.title }}</span>
          </div>
          <div v-if="data.newPage.meta.modifiedByDisplayName" class="text-sm text-muted">
            <UIcon name="tabler:user" class="mr-1 align-middle" />
            <span class="align-middle">{{ data.newPage.meta.modifiedByDisplayName }}</span>
          </div>
          <div v-if="data.newPage.meta.tags?.length" class="mt-2">
            <Tags :model-value="data.newPage.meta.tags" />
          </div>
        </div>
      </div>

      <DiffView
        v-model:mode="diffMode"
        :old-text="data.oldPage.content"
        :new-text="data.newPage.content"
        :old-label="oldLabel"
        :new-label="newLabel"
      />
    </div>

    <div v-else class="text-center">
      <span class="text-3xl">😟</span>
      <div class="font-medium m-4">
        {{ $t('diff.one-revision-missing') }}
      </div>

      <ReactiveButton
        icon="tabler:arrow-back-up"
        :label="$t('revisions')"
        @click="navTo({ query: { rev: null, diff: undefined } })"
      />
    </div>
  </Layout>
</template>
