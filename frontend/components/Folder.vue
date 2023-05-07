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
  onReload: Function
}>()

const { t } = useI18n()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const urlPath = computed(() => props.urlPath)
const folder = computed(() => props.folder)
const breadcrumbs = computed(() => props.breadcrumbs)

const pageTitle = computed(() => {
  if (urlPath.value === '') {
    return t('home')
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
    const msgBox = await ElMessageBox.prompt(t('enter-page-name'), t('create-page'), {
      confirmButtonText: t('ok'),
      cancelButtonText: t('cancel'),
      inputPattern: /^[a-z0-9-][a-z0-9_-]*$/,
      inputErrorMessage: t('invalid-page-name'),
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
    const msgBox = await ElMessageBox.prompt(t('enter-folder-name'), t('create-folder'), {
      confirmButtonText: t('ok'),
      cancelButtonText: t('cancel'),
      inputPattern: /^[a-z0-9-][a-z0-9_-]*$/,
      inputErrorMessage: t('invalid-folder-name'),
    })
    name = msgBox.value
  } catch (e) {
    return
  }

  try {
    await apiFetch(`/pages${urlPath.value}/${name}`, { method: 'PUT', body: { page: null } })

    ElMessage({
      message: t('folder-created'),
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

const deleteConfirmOpen = ref(false)
const onDeleteFolder = async () => {
  if (deleteConfirmOpen.value) {
    // Prevent multiple dialogs at the same time
    return
  }

  deleteConfirmOpen.value = true
  try {
    await ElMessageBox.confirm(
      t('are-you-sure-to-delete-this-folder'),
      {
        confirmButtonText: t('delete'),
        cancelButtonText: t('cancel'),
        type: 'warning',
      })
  } catch {
    // do nothing
    deleteConfirmOpen.value = false
    return
  }
  deleteConfirmOpen.value = false

  try {
    await apiFetch(`/pages${urlPath.value}`, { method: 'DELETE' })

    ElMessage({
      message: t('folder-deleted'),
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
    await props.onReload()
    ElMessage({ message: t('folder-reloaded'), type: 'success' })
  } else if (command === 'acl') {
    await navigateTo({ query: { acl: null } })
  } else if (command === 'delete') {
    onDeleteFolder()
  } else {
    throw new Error(`Unhandled command ${command}`)
  }
}

onKeyStroke('Delete', (e) => {
  if (urlPath.value !== '' && props.allowDelete) {
    e.preventDefault()
    onDeleteFolder()
  }
})
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
          <Icon name="ci:file-add" /> <span class="hidden md:inline ml-1">{{ $t('create-page') }}</span>
        </ElButton>
        <span />
        <ElButton v-if="allowWrite" class="m-1" @click="createFolder">
          <Icon name="ci:folder-add" /> <span class="hidden md:inline ml-1">{{ $t('create-folder') }}</span>
        </ElButton>

        <ElDropdown trigger="click" class="m-1" @command="handleDropdownMenuCommand">
          <ElButton>
            <Icon name="ci:more-vertical" /> <span class="hidden md:inline ml-1">{{ $t('more') }}</span>
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="ReloadIcon" command="reload">
                {{ $t('reload') }}
              </ElDropdownItem>
              <ElDropdownItem v-if="allowAdmin" :icon="PermissionsIcon" command="acl">
                {{ $t('permissions') }}
              </ElDropdownItem>
              <ElDropdownItem v-if="urlPath !== '' && allowDelete" :icon="DeleteIcon" command="delete">
                {{ $t('delete') }}
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>
    </template>

    <div>
      <h2
        v-if="folder.content.some(e => e.isFolder)"
        class="font-light text-xl"
      >
        {{ $t('folders') }}
      </h2>
      <div v-for="entry of folder.content.filter(e => e.isFolder)" :key="entry.name">
        <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
          <ElLink :href="href" @click="navigate">
            <Icon name="ci:folder" class="mr-1" /> {{ entry.name }}
          </ElLink>
        </NuxtLink>
      </div>

      <h2
        v-if="folder.content.some(e => !e.isFolder)"
        class="font-light text-xl"
      >
        {{ $t('pages') }}
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
