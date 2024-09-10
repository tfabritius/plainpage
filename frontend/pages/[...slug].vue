<script setup lang="ts">
import { FetchError } from 'ofetch'
import { storeToRefs } from 'pinia'

import { useAuthStore } from '~/store/auth'
import type { GetContentResponse } from '~/types/'

const { t } = useI18n()

const route = useRoute()
const urlPath = computed(() => route.path.replace(/^\//, ''))

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

const { data, error, refresh } = await useAsyncData(`/pages/${urlPath.value}:${loggedIn}`, async () => {
  try {
    const data = await apiFetch<GetContentResponse>(`/pages/${urlPath.value}`)
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
      const data = JSON.parse(err.response?._data) as GetContentResponse
      return { notFound: true, accessDenied: false, ...data }
    }
    throw err
  }
}, {
  watch: [loggedIn],
})

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

const _deepClone = deepClone
</script>

<template>
  <div class="flex flex-col">
    <SubpageNetworkError
      v-if="!folder && !page && !notFound && !accessDenied"
      :msg="error?.message"
      :on-reload="refresh"
    />
    <SubpageAccessDenied
      v-else-if="data?.accessDenied"
    />
    <SubpageAtticList
      v-else-if="revQuery === null"
      :title="pageTitle"
      :url-path="urlPath"
    />
    <SubpageAtticPage
      v-else-if="revQuery !== undefined"
      :url-path="urlPath"
      :revision="revQuery"
    />
    <SubpageContentPermissions
      v-else-if="folder && aclQuery"
      :is-folder="true"
      :url-path="urlPath"
      :meta="_deepClone(folder.meta)"
      :title="urlPath === '' ? $t('home') : data?.breadcrumbs.slice(-1)[0]?.name"
      :breadcrumbs="data?.breadcrumbs ?? []"
      @refresh="refresh"
    />
    <SubpageFolder
      v-else-if="folder"
      :allow-write="data?.allowWrite ?? false"
      :allow-delete="data?.allowDelete ?? false"
      :breadcrumbs="data?.breadcrumbs ?? []"
      :folder="folder"
      :url-path="urlPath"
      :on-reload="refresh"
    />
    <SubpageContentPermissions
      v-else-if="page && aclQuery"
      :is-folder="false"
      :url-path="urlPath"
      :meta="_deepClone(page.meta)"
      :title="page.meta.title"
      :breadcrumbs="data?.breadcrumbs ?? []"
      @refresh="refresh"
    />
    <SubpagePage
      v-else-if="page"
      :page="page"
      :breadcrumbs="data?.breadcrumbs ?? []"
      :allow-write="data?.allowWrite ?? false"
      :allow-delete="data?.allowDelete ?? false"
      :on-reload="refresh"
    />
    <SubpageNotFound
      v-else
      :url-path="urlPath"
      :breadcrumbs="data?.breadcrumbs ?? []"
      :allow-create="data?.allowWrite ?? false"
      @refresh="refresh"
    />
  </div>
</template>
