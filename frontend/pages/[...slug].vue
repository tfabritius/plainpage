<script setup lang="ts">
import { FetchError } from 'ofetch'
import { useRouteQuery } from '@vueuse/router'
import { Icon } from '#components'

import type { GetResponse, Page } from '~/types/'

const route = useRoute()

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref(deepClone(emptyPage))

const { data, error, refresh } = await useAsyncData(route.path, async () => {
  try {
    const relUrl = route.path === '/' ? '' : route.path
    const data = await $fetch<GetResponse>(`/_api/pages${relUrl}`)
    return {
      notFound: false,
      ...data,
    }
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 404) {
      editablePage.value.url = route.path

      const data = JSON.parse(err.response?._data) as GetResponse
      return { notFound: true, ...data }
    }
    throw err
  }
})

const page = computed(() => data.value?.page ?? null)
const notFound = computed(() => data.value?.notFound === true)
const allowCreate = computed(() => data.value?.allowCreate === true)
const folder = computed(() => data.value?.folder ?? null)

const pageTitle = computed(() => {
  if (folder.value && route.path === '/') {
    return 'Home'
  } else if (folder.value) {
    return data.value?.breadcrumbs.slice(-1)[0]?.name
  } else if (page.value) {
    return page.value.meta.title || 'Untitled'
  }
  return 'Not found'
})

useHead(() => ({ title: pageTitle.value }),
)

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === 'true' && (!!page.value || allowCreate.value) },
  set(value) {
    editQuery.value = value ? 'true' : null
  },
})

watch(editing, (editing) => {
  if (editing && page.value) {
    editablePage.value = deepClone(page.value)
  }
}, { immediate: true })

const ChevronIcon = h(Icon, { name: 'ci:chevron-right' })
const RevisionsIcon = h(Icon, { name: 'ic:baseline-restore' })
const DeleteIcon = h(Icon, { name: 'ci:trash-full' })
const ReloadIcon = h(Icon, { name: 'ci:arrows-reload-01' })

const createThisPage = () => {
  editablePage.value = deepClone(emptyPage)
  editablePage.value.url = route.path
  editing.value = true
}

const createThisFolder = async () => {
  const urlPath = route.path
  await $fetch(`/_api/pages/${urlPath}`, { method: 'PUT', body: { page: null } })

  ElMessage({
    message: 'Folder created',
    type: 'success',
  })
  refresh()
}

const createPage = async () => {
  let name
  try {
    const msgBox = await ElMessageBox.prompt('Please enter page name', 'New page', {
      confirmButtonText: 'OK',
      cancelButtonText: 'Cancel',
      inputPattern: /^[a-z0-9-][a-z0-9_-]*$/,
      inputErrorMessage: 'Invalid name (allowed: [a-z0-9_-])',
    })
    name = msgBox.value
  } catch (e) {
    return
  }

  const urlPath = route.path === '/' ? `/${name}` : `${route.path}/${name}`
  await navigateTo({ path: urlPath, query: { edit: 'true' } }, { replace: true })
}

const createFolder = async () => {
  let name
  try {
    const msgBox = await ElMessageBox.prompt('Please enter folder name', 'New folder', {
      confirmButtonText: 'OK',
      cancelButtonText: 'Cancel',
      inputPattern: /^[a-z0-9-][a-z0-9_-]*$/,
      inputErrorMessage: 'Invalid name (allowed: [a-z0-9_-])',
    })
    name = msgBox.value
  } catch (e) {
    return
  }

  try {
    const urlPath = route.path === '/' ? `/${name}` : `${route.path}/${name}`
    await $fetch(`/_api/pages/${urlPath}`, { method: 'PUT', body: { page: null } })

    ElMessage({
      message: 'Folder created',
      type: 'success',
    })
    await navigateTo(urlPath)
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const onEditPage = () => {
  editing.value = true
}

const onSavePage = async () => {
  try {
    await $fetch(`/_api/pages${editablePage.value.url}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    ElMessage({
      message: 'Saved',
      type: 'success',
    })

    refresh()
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const onDeletePageOrFolder = async () => {
  try {
    await ElMessageBox.confirm(
      `Are you sure to delete this ${page.value ? 'page' : 'folder'}?`,
      {
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        type: 'warning',
      })
  } catch {
    // do nothing
    return
  }

  try {
    await $fetch(`/_api/pages${route.path}`, { method: 'DELETE' })

    ElMessage({
      message: `${page.value ? 'Page' : 'Folder'} deleted`,
      type: 'success',
    })

    await navigateTo(route.path.substring(0, route.path.lastIndexOf('/') + 1))
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const handleDropdownMenuCommand = (command: string | number | object) => {
  if (command === 'reload') {
    refresh()
  } else if (command === 'delete') {
    onDeletePageOrFolder()
  } else if (command === 'rev') {
    ElMessage('Not implemented yet')
  } else {
    throw new Error(`Unhandled command ${command}`)
  }
}

const cancelEditConfirmOpen = ref(false)

const onCancelEdit = async () => {
  if (cancelEditConfirmOpen.value) {
    ElMessageBox.close()
    cancelEditConfirmOpen.value = false
    return
  }

  if (!deepEqual(page.value, editablePage.value)) {
    try {
      cancelEditConfirmOpen.value = true
      await ElMessageBox.confirm('Discard changes to this page?', {
        confirmButtonText: 'OK',
        cancelButtonText: 'Cancel',
        type: 'warning',
        closeOnPressEscape: false,
      })
    } catch {
      cancelEditConfirmOpen.value = false
      return
    }
    cancelEditConfirmOpen.value = false
  }

  editing.value = false
}

onKeyStroke('Escape', async (_event: KeyboardEvent) => {
  if (editing.value) {
    onCancelEdit()
  }
}, { eventName: 'keyup' })
</script>

<template>
  <NetworkError v-if="!folder && !page && !notFound" :msg="error?.message" @refresh="refresh" />
  <Layout v-else>
    <template #breadcrumbs>
      <ElBreadcrumb v-if="!notFound" :separator-icon="ChevronIcon">
        <ElBreadcrumbItem :to="{ path: '/' }">
          <Icon name="ic:outline-home" />
        </ElBreadcrumbItem>

        <ElBreadcrumbItem v-for="crumb in data?.breadcrumbs" :key="crumb.url" :to="{ path: crumb.url }">
          {{ crumb.name }}
        </ElBreadcrumbItem>
      </ElBreadcrumb>
    </template>

    <template #title>
      <Icon v-if="folder" name="ci:folder" class="mr-1" />
      <span v-if="folder || page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">
        {{ pageTitle }}
      </span>
    </template>

    <template #actions>
      <div v-if="!editing">
        <ElButton v-if="page" class="m-1" @click="onEditPage">
          <Icon name="ci:edit" /> <span class="hidden md:inline ml-1">Edit</span>
        </ElButton>

        <ElButton v-if="folder" class="m-1" @click="createPage">
          <Icon name="ci:file-add" /> <span class="hidden md:inline ml-1">Add page</span>
        </ElButton>
        <span />
        <ElButton v-if="folder" class="m-1" @click="createFolder">
          <Icon name="ci:folder-add" /> <span class="hidden md:inline ml-1">Add folder</span>
        </ElButton>

        <ElDropdown trigger="click" class="m-1" @command="handleDropdownMenuCommand">
          <ElButton>
            <Icon name="ci:more-vertical" /> <span class="hidden md:inline ml-1">More</span>
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="ReloadIcon" command="reload">
                Reload
              </ElDropdownItem>
              <ElDropdownItem v-if="page" :icon="RevisionsIcon" command="rev">
                Revisions
              </ElDropdownItem>
              <ElDropdownItem v-if="(page || folder) && route.path !== '/'" :icon="DeleteIcon" command="delete">
                Delete
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>

      <div v-if="editing">
        <ElButton class="ml-2" @click="onCancelEdit">
          <Icon name="ci:close-md" /> <span class="hidden md:inline ml-1">Cancel</span>
        </ElButton>
        <ElButton type="success" @click="onSavePage">
          <Icon name="ci:save" /> <span class="hidden md:inline ml-1">Save</span>
        </ElButton>
      </div>
    </template>

    <Folder v-if="folder" :folder="folder" />
    <div v-else-if="page">
      <PageView v-if="!editing" :page="page" />
      <PageEditor v-else v-model="editablePage" />
    </div>
    <div v-else>
      <div v-if="!editing" class="text-center">
        <span class="text-3xl">😟</span>
        <div class="font-medium m-4">
          This page doesn't exist!
        </div>

        <div v-if="allowCreate">
          <ElButton @click="createThisPage">
            <Icon name="ci:file-add" /> <span class="hidden md:inline ml-1">Create page</span>
          </ElButton>
          <ElButton @click="createThisFolder">
            <Icon name="ci:folder-add" /> <span class="hidden md:inline ml-1">Create folder</span>
          </ElButton>
        </div>

        <ElButton v-else @click="navigateTo('/')">
          <Icon name="ic:outline-home" /> <span class="hidden md:inline ml-1">Back home</span>
        </ElButton>
      </div>

      <div v-else>
        <PageEditor v-model="editablePage" />
      </div>
    </div>
  </Layout>
</template>
