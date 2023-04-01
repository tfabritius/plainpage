<script setup lang="ts">
import { FetchError } from 'ofetch'
import { useRouteQuery } from '@vueuse/router'
import { Icon } from '#components'

import type { GetPageResponse, Page } from '~/types/'

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

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref(deepClone(emptyPage))

const { data, error, refresh } = await useAsyncData(route.path, async () => {
  try {
    const relUrl = route.path === '/' ? '' : route.path
    const data = await $fetch<GetPageResponse>(`/_api/pages${relUrl}`)
    return {
      notFound: false,
      ...data,
    }
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 404) {
      editablePage.value.url = route.path

      const data = JSON.parse(err.response?._data) as GetPageResponse
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
  if (page.value) {
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

const PermissionsIcon = h(Icon, { name: 'ci:shield' })
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

const onDeletePage = async () => {
  try {
    await ElMessageBox.confirm(
      'Are you sure to delete this page?',
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
      message: 'Page deleted',
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

const handleDropdownMenuCommand = async (command: string | number | object) => {
  if (command === 'reload') {
    refresh()
  } else if (command === 'delete') {
    onDeletePage()
  } else if (command === 'rev') {
    await navigateTo({ query: { rev: null } })
  } else if (command === 'acl') {
    await navigateTo({ query: { acl: null } })
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
  <AtticList v-else-if="revQuery === null" :title="pageTitle" :url-path="urlPath" />
  <AtticPage v-else-if="revQuery !== undefined" :url-path="urlPath" :revision="revQuery" />
  <PageFolderPermissions
    v-else-if="folder && aclQuery"
    :is-folder="true"
    :url-path="urlPath"
    :meta="deepClone(folder.meta)"
    :title="urlPath === '' ? 'Home' : data?.breadcrumbs.slice(-1)[0]?.name"
    :breadcrumbs="data?.breadcrumbs ?? []"
    @refresh="refresh"
  />
  <Folder v-else-if="folder" :breadcrumbs="data?.breadcrumbs ?? []" :folder="folder" :url-path="urlPath" />
  <PageFolderPermissions v-else-if="page && aclQuery" :is-folder="false" :url-path="urlPath" :meta="deepClone(page.meta)" :title="page.meta.title" :breadcrumbs="data?.breadcrumbs ?? []" @refresh="refresh" />
  <Layout v-else :breadcrumbs="data?.breadcrumbs ?? []">
    <template #title>
      <span v-if="page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">
        {{ pageTitle }}
      </span>
    </template>

    <template #actions>
      <div v-if="!editing">
        <ElButton v-if="page" class="m-1" @click="onEditPage">
          <Icon name="ci:edit" /> <span class="hidden md:inline ml-1">Edit</span>
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
              <ElDropdownItem v-if="page" :icon="PermissionsIcon" command="acl">
                Permissions
              </ElDropdownItem>
              <ElDropdownItem v-if="page" :icon="DeleteIcon" command="delete">
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

    <div v-if="page">
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
