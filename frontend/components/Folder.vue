<script setup lang="ts">
import { storeToRefs } from 'pinia'
import type { Breadcrumb, Folder } from '~/types'
import { Icon } from '#components'
import { useAppStore } from '~/store/app'

const props = defineProps<{
  urlPath: string
  folder: Folder
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const urlPath = computed(() => props.urlPath)
const folder = computed(() => props.folder)
const breadcrumbs = computed(() => props.breadcrumbs)

const pageTitle = computed(() => {
  if (urlPath.value === '') {
    return 'Home'
  }
  return breadcrumbs.value.slice(-1)[0]?.name
})

useHead(() => ({ title: pageTitle.value }))

const PermissionsIcon = h(Icon, { name: 'ci:shield' })
const DeleteIcon = h(Icon, { name: 'ci:trash-full' })
const ReloadIcon = h(Icon, { name: 'ci:arrows-reload-01' })

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

  await navigateTo({ path: `${urlPath.value}/${name}`, query: { edit: 'true' } })
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
    await apiFetch(`/pages${urlPath.value}/${name}`, { method: 'PUT', body: { page: null } })

    ElMessage({
      message: 'Folder created',
      type: 'success',
    })
    await navigateTo(`${urlPath.value}/${name}`)
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const onDeleteFolder = async () => {
  try {
    await ElMessageBox.confirm(
      'Are you sure to delete this folder?',
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
    await apiFetch(`/pages${urlPath.value}`, { method: 'DELETE' })

    ElMessage({
      message: 'Folder deleted',
      type: 'success',
    })

    await navigateTo('./')
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const handleDropdownMenuCommand = async (command: string | number | object) => {
  if (command === 'reload') {
    emit('refresh')
  } else if (command === 'acl') {
    await navigateTo({ query: { acl: null } })
  } else if (command === 'delete') {
    onDeleteFolder()
  } else {
    throw new Error(`Unhandled command ${command}`)
  }
}
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <Icon name="ci:folder" class="mr-1" />
      {{ pageTitle }}
    </template>

    <template #actions>
      <div>
        <ElButton v-if="allowWrite" class="m-1" @click="createPage">
          <Icon name="ci:file-add" /> <span class="hidden md:inline ml-1">Add page</span>
        </ElButton>
        <span />
        <ElButton v-if="allowWrite" class="m-1" @click="createFolder">
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
              <ElDropdownItem v-if="allowAdmin" :icon="PermissionsIcon" command="acl">
                Permissions
              </ElDropdownItem>
              <ElDropdownItem v-if="urlPath !== '' && allowDelete" :icon="DeleteIcon" command="delete">
                Delete
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>
    </template>

    <div>
      <h2 v-if="folder.content.some(e => e.isFolder)" class="font-light text-xl">
        Folders
      </h2>
      <div v-for="entry of folder.content.filter(e => e.isFolder)" :key="entry.name">
        <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
          <ElLink :href="href" @click="navigate">
            <Icon name="ci:folder" class="mr-1" /> {{ entry.name }}
          </ElLink>
        </NuxtLink>
      </div>

      <h2
        v-if="folder.content.some(e => !e.isFolder)" class="font-light text-xl"
      >
        Pages
      </h2>
      <div v-for="entry of folder.content.filter(e => !e.isFolder)" :key="entry.name">
        <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
          <ElLink :href="href" @click="navigate">
            {{ entry.name }}
          </ElLink>
        </NuxtLink>
      </div>
    </div>
  </Layout>
</template>
