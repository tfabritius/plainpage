<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import { storeToRefs } from 'pinia'

import { useAppStore } from '~/store/app'
import type { Breadcrumb, Folder, PutRequest } from '~/types'

const props = defineProps<{
  urlPath: string
  folder: Folder
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
  onReload: () => void
}>()

const subfolders = computed(() => props.folder.content.filter(e => e.isFolder))
const pages = computed(() => props.folder.content.filter(e => !e.isFolder))

const { t } = useI18n()
const toast = useToast()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const pageTitle = computed(() => {
  if (props.urlPath === '') {
    return t('home')
  }
  return props.folder.meta.title || props.breadcrumbs.slice(-1)[0]?.name
})

useHead(() => ({ title: pageTitle.value }))

const editTitelOpen = ref(false)
const editableTitle = ref('')
async function saveEditedTitle() {
  try {
    const body = { folder: { meta: { title: editableTitle.value, tags: null }, content: [] } } satisfies PutRequest
    await apiFetch(`/pages/${props.urlPath}`, { method: 'PUT', body })

    editTitelOpen.value = false
    props.onReload()
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

const plainDialog = useTemplateRef('plainDialog')

const deleteConfirmOpen = ref(false)
async function onDeleteFolder() {
  if (deleteConfirmOpen.value) {
    // Prevent multiple dialogs at the same time
    return
  }

  deleteConfirmOpen.value = true
  if (!await plainDialog.value?.confirm(
    t('are-you-sure-to-delete-this-folder'),
    {
      confirmButtonText: t('delete'),
      confirmButtonColor: 'warning',
    },
  )) {
    // do nothing
    deleteConfirmOpen.value = false
    return
  }
  deleteConfirmOpen.value = false

  try {
    await apiFetch(`/pages/${props.urlPath}`, { method: 'DELETE' })

    toast.add({
      description: t('folder-deleted'),
      color: 'success',
    })

    await navigateTo('./')
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

const menuItems = computed(() => {
  const items: DropdownMenuItem[] = []

  items.push(
    {
      icon: 'ci:arrows-reload-01',
      label: t('reload'),
      onSelect: async () => {
        await props.onReload()
        toast.add({ description: t('folder-reloaded'), color: 'success' })
      },
    },
  )

  if (allowAdmin.value) {
    items.push(
      {
        icon: 'ci:shield',
        label: t('permissions'),
        onSelect: async () => {
          await navigateTo({ query: { acl: null } })
        },
      },
    )
  }

  if (props.urlPath !== '' && props.allowDelete) {
    items.push(
      {
        icon: 'ci:trash-full',
        label: t('delete'),
        onSelect: onDeleteFolder,
      },
    )
  }

  return items
})

onKeyStroke('Backspace', (e) => {
  if (props.urlPath !== '' && props.allowDelete && e.ctrlKey) {
    e.preventDefault()
    onDeleteFolder()
  }
})
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <UIcon name="ci:folder" class="mr-1" />
      {{ pageTitle }}
    </template>

    <template v-if="urlPath !== ''" #title:suffix>
      <UModal v-model:open="editTitelOpen">
        <UButton
          class="opacity-0 group-hover:opacity-100 duration-100"
          variant="link"
          color="neutral"
          @click="editableTitle = props.folder.meta.title"
        >
          <UIcon name="ci:edit" size="1.5em" />
        </UButton>
        <template #title>
          {{ t('folder-title') }}
        </template>
        <template #body>
          <form id="editTitleForm" @submit.prevent="saveEditedTitle">
            <UInput v-model="editableTitle" class="w-full" autofocus />
          </form>
        </template>
        <template #footer>
          <PlainButton :label="t('cancel')" @click="editTitelOpen = false" />
          <PlainButton color="primary" :label="t('ok')" type="submit" form="editTitleForm" />
        </template>
      </UModal>
    </template>

    <template #actions>
      <div>
        <NewContentModal v-if="allowWrite" type="page" :url-path="urlPath">
          <PlainButton icon="ci:file-add" :label="$t('create-page')" />
        </NewContentModal>
        <NewContentModal v-if="allowWrite" type="folder" :url-path="urlPath">
          <PlainButton icon="ci:folder-add" :label="$t('create-folder')" class="ml-3" />
        </NewContentModal>

        <UDropdownMenu :items="menuItems">
          <PlainButton icon="ci:more-vertical" :label="$t('more')" class="ml-3" />
        </UDropdownMenu>
      </div>
    </template>

    <div v-if="subfolders.length > 0">
      <h2 class="font-light text-xl my-4">
        {{ $t('folders') }}
      </h2>
      <MultiColumnList
        :items="subfolders"
        :sort-and-group-by="item => item.title || item.name"
        :group-if-more-than="10"
      >
        <template #item="{ item }">
          <ULink :to="`/${item.url}`">
            <UIcon name="ci:folder" class="align-middle" /> <span class="align-middle">{{ item.title || item.name }}</span>
          </ULink>
        </template>
      </MultiColumnList>
    </div>

    <div v-if="pages.length > 0">
      <h2 class="font-light text-xl my-4">
        {{ $t('pages') }}
      </h2>
      <MultiColumnList
        :items="pages"
        :sort-and-group-by="item => item.title || item.name"
        :group-if-more-than="10"
      >
        <template #item="{ item }">
          <ULink :to="`/${item.url}`">
            {{ item.title || item.name }}
          </ULink>
        </template>
      </MultiColumnList>
    </div>

    <PlainDialog ref="plainDialog" />
  </Layout>
</template>
