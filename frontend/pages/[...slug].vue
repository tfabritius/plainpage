<script setup lang="ts">
import { FetchError } from 'ofetch'
import { storeToRefs } from 'pinia'

import type { GetContentResponse, Page } from '~/types/'
import { useAuthStore } from '~/store/auth'

const { t } = useI18n()

const route = useRoute()
const urlPath = computed(() => route.path === '/' ? '' : route.path)

const revQuery = computed(() => {
  const data = route.query.rev
  if (data === undefined || data === null) {
    return data
  }
  if (Array.isArray(data)) {
    return null
  }
  return data
})

const aclQuery = computed(() => {
  if (route.query.acl === undefined) {
    return false
  }
  return true
})

const auth = useAuthStore()
const { loggedIn } = storeToRefs(auth)
const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref(deepClone(emptyPage))

const { data, error, refresh } = await useAsyncData(`/pages${urlPath.value}:${loggedIn}`, async () => {
  try {
    const data = await apiFetch<GetContentResponse>(`/pages${urlPath.value}`)
    return {
      notFound: false,
      accessDenied: false,
      ...data,
    }
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 403) {
      return {
        accessDenied: true,
        notFound: false,
        page: null,
        folder: null,
        breadcrumbs: [],
        allowWrite: false,
        allowDelete: false,
      }
    }
    if (err instanceof FetchError && err.statusCode === 404) {
      editablePage.value.url = route.path

      const data = JSON.parse(err.response?._data) as GetContentResponse
      return { notFound: true, accessDenied: false, ...data }
    }
    throw err
  }
},
{ watch: [loggedIn] })

const accessDenied = computed(() => data.value?.accessDenied ?? false)
const page = computed(() => data.value?.page ?? null)
const notFound = computed(() => data.value?.notFound === true)
const folder = computed(() => data.value?.folder ?? null)

const pageTitle = computed(() => {
  if (page.value) {
    return page.value.meta.title || t('untitled')
  }
  return t('not-found')
})
</script>

<template>
  <NetworkError
    v-if="!folder && !page && !notFound && !accessDenied"
    :msg="error?.message"
    :on-reload="refresh"
  />
  <AccessDenied
    v-else-if="data?.accessDenied"
  />
  <AtticList
    v-else-if="revQuery === null"
    :title="pageTitle"
    :url-path="urlPath"
  />
  <AtticPage
    v-else-if="revQuery !== undefined"
    :url-path="urlPath"
    :revision="revQuery"
  />
  <ContentPermissions
    v-else-if="folder && aclQuery"
    :is-folder="true"
    :url-path="urlPath"
    :meta="deepClone(folder.meta)"
    :title="urlPath === '' ? $t('home') : data?.breadcrumbs.slice(-1)[0]?.name"
    :breadcrumbs="data?.breadcrumbs ?? []"
    @refresh="refresh"
  />
  <Folder
    v-else-if="folder"
    :allow-write="data?.allowWrite ?? false"
    :allow-delete="data?.allowDelete ?? false"
    :breadcrumbs="data?.breadcrumbs ?? []"
    :folder="folder"
    :url-path="urlPath"
    :on-reload="refresh"
  />
  <ContentPermissions
    v-else-if="page && aclQuery"
    :is-folder="false"
    :url-path="urlPath"
    :meta="deepClone(page.meta)"
    :title="page.meta.title"
    :breadcrumbs="data?.breadcrumbs ?? []"
    @refresh="refresh"
  />
  <PPage
    v-else-if="page"
    :page="page"
    :breadcrumbs="data?.breadcrumbs ?? []"
    :allow-write="data?.allowWrite ?? false"
    :allow-delete="data?.allowDelete ?? false"
    :on-reload="refresh"
  />
  <NotFound
    v-else
    :url-path="urlPath"
    :breadcrumbs="data?.breadcrumbs ?? []"
    :allow-create="data?.allowWrite ?? false"
    @refresh="refresh"
  />
</template>
