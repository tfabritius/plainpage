<script setup lang="ts">
import type { GetContentResponse } from '~/types/'
import { FetchError } from 'ofetch'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '~/store/auth'

const { t } = useI18n()

const route = useRoute()
// eslint-disable-next-line e18e/prefer-static-regex
const urlPath = computed(() => route.path.replace(/^\//, ''))

// revQuery is a string if ?rev=123 is passed,
// null if ?rev is passed (no value),
// undefined if rev isn't used as query parameter,
// is null if ?rev=123&rev=456 is passed (multiple values) - not used
const revQuery = useRouteQuery('rev', undefined, {
  transform: (data: string | string[] | null | undefined) => Array.isArray(data) ? null : data,
})

const diffQuery = useRouteQuery('diff', undefined, {
  transform: (data: string | string[] | null | undefined) => Array.isArray(data) ? null : data,
})

const aclQuery = useRouteQuery('acl', undefined, { transform: data => data !== undefined })

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
    <SubpageAtticDiff
      v-else-if="revQuery && diffQuery"
      :url-path="urlPath"
      :revision1="revQuery"
      :revision2="diffQuery"
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
      :meta="deepClone(folder.meta)"
      :title="urlPath === '' ? $t('home') : (folder.meta.title || data?.breadcrumbs.slice(-1)[0]?.name)"
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
      :meta="deepClone(page.meta)"
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
