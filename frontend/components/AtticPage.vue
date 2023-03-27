<script setup lang="ts">
import { FetchError } from 'ofetch'
import { format } from 'date-fns'
import { Icon } from '#components'

import type { Breadcrumb, GetResponse } from '~/types/'

const props = defineProps<{
  urlPath: string
  revision: string
  breadcrumbs: Breadcrumb[]
}>()

const urlPath = computed(() => props.urlPath)
const revision = computed(() => props.revision)

const { data, error, refresh } = await useAsyncData(`/_api/attic${urlPath.value}?rev=${revision.value}`, async () => {
  try {
    const data = await $fetch<GetResponse>(`/_api/attic${urlPath.value}?rev=${revision.value}`)
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
    return data.value.page.meta.title || 'Untitled'
  }
  return 'Not found'
})

useHead(() => ({ title: pageTitle.value }))

const revDate = computed(() => new Date(Number(revision.value) * 1000))

const ChevronIcon = h(Icon, { name: 'ci:chevron-right' })
</script>

<template>
  <NetworkError v-if="error || !data" :msg="error?.message || ''" @refresh="refresh" />
  <Layout v-else>
    <template #breadcrumbs>
      <ElBreadcrumb :separator-icon="ChevronIcon">
        <ElBreadcrumbItem :to="{ path: '/' }">
          <Icon name="ic:outline-home" />
        </ElBreadcrumbItem>

        <ElBreadcrumbItem v-for="crumb in breadcrumbs" :key="crumb.url" :to="{ path: crumb.url }">
          {{ crumb.name }}
        </ElBreadcrumbItem>
      </ElBreadcrumb>
    </template>

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
        <Icon name="ic:baseline-update" /> <span class="hidden md:inline ml-1">Current version</span>
      </ElButton>
    </template>

    <div v-if="data.page">
      <PageView :page="data.page" />
    </div>
    <div v-else>
      <div class="text-center">
        <span class="text-3xl">😟</span>
        <div class="font-medium m-4">
          This revision doesn't exist!
        </div>

        <ElButton @click="navigateTo({ query: { rev: undefined } })">
          <Icon name="ic:baseline-update" /> <span class="hidden md:inline ml-1">Current version</span>
        </ElButton>
      </div>
    </div>
  </Layout>
</template>
