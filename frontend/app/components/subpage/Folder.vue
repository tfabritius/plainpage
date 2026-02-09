<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { Breadcrumb, Folder, PatchOperation } from '~/types'
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'
import { validUrlPartRegex } from '~/types/api'

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

const editFolderOpen = ref(false)
const editableTitle = ref('')
const editableName = ref('')

// Computed properties for folder name/path manipulation
const currentFolderName = computed(() => {
  const urlParts = props.urlPath.split('/')
  return urlParts[urlParts.length - 1] || ''
})
const parentPath = computed(() => {
  const urlParts = props.urlPath.split('/')
  urlParts.pop()
  return urlParts.join('/')
})
const isValidFolderName = computed(() => validUrlPartRegex.test(editableName.value))

function openEditFolderModal() {
  editableTitle.value = props.folder.meta.title
  editableName.value = currentFolderName.value
  editFolderOpen.value = true
}

// Move functionality
const moveModalOpen = ref(false)
async function onFolderMoved(newPath: string) {
  await navigateTo(`/${newPath}`)
}

async function saveEditedFolder() {
  // Validate folder name if we're renaming (non-root folder)
  if (props.urlPath !== '' && !isValidFolderName.value) {
    return
  }

  try {
    const newUrl = parentPath.value ? `${parentPath.value}/${editableName.value}` : editableName.value
    const nameChanged = props.urlPath !== '' && newUrl !== props.urlPath
    const titleChanged = editableTitle.value !== props.folder.meta.title

    // Build patch operations
    const operations: PatchOperation[] = []

    if (nameChanged) {
      operations.push({ op: 'replace', path: '/folder/url', value: newUrl })
    }

    if (titleChanged) {
      operations.push({ op: 'replace', path: '/folder/meta/title', value: editableTitle.value })
    }

    // Only make request if there are changes
    if (operations.length > 0) {
      await apiFetch(`/pages/${props.urlPath}`, {
        method: 'PATCH',
        body: operations,
      })

      toast.add({
        description: t('saved'),
        color: 'success',
      })
    }

    editFolderOpen.value = false

    // Navigate to new URL if folder was renamed, otherwise just reload
    if (nameChanged) {
      await navigateTo(`/${newUrl}`)
    } else if (titleChanged) {
      props.onReload()
    }
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
      icon: 'tabler:refresh',
      label: t('reload'),
      onSelect: async () => {
        await props.onReload()
        toast.add({ description: t('folder-reloaded'), color: 'success' })
      },
    },
  )

  if (props.urlPath !== '' && props.allowWrite) {
    items.push({
      icon: 'tabler:pencil',
      label: t('edit-folder'),
      onSelect: openEditFolderModal,
    })
  }

  if (props.urlPath !== '' && props.allowWrite && props.allowDelete) {
    items.push({
      icon: 'tabler:folder-symlink',
      label: t('move'),
      onSelect: () => { moveModalOpen.value = true },
    })
  }

  if (allowAdmin.value) {
    items.push(
      {
        icon: 'tabler:lock',
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
        icon: 'tabler:trash',
        label: t('delete'),
        onSelect: onDeleteFolder,
      },
    )
  }

  return items
})

onKeyStroke('Backspace', (e) => {
  if (props.urlPath !== '' && props.allowDelete && e.ctrlKey && !editFolderOpen.value) {
    e.preventDefault()
    onDeleteFolder()
  }
})
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <UIcon name="tabler:folder" class="mr-2" />
      {{ pageTitle }}
    </template>

    <template v-if="urlPath !== ''" #title:suffix>
      <UButton
        class="opacity-0 group-hover:opacity-100 duration-100"
        variant="link"
        color="neutral"
        @click="openEditFolderModal"
      >
        <UIcon name="tabler:edit" size="1.5em" />
      </UButton>
    </template>

    <template #actions>
      <NewContentModal v-if="allowWrite" type="page" :url-path="urlPath">
        <ReactiveButton icon="tabler:file-plus" :label="$t('create-page')" />
      </NewContentModal>
      <NewContentModal v-if="allowWrite" type="folder" :url-path="urlPath">
        <ReactiveButton icon="tabler:folder-plus" :label="$t('create-folder')" />
      </NewContentModal>

      <UDropdownMenu :items="menuItems">
        <ReactiveButton icon="tabler:dots-vertical" :label="$t('more')" />
      </UDropdownMenu>
    </template>

    <div v-if="subfolders.length > 0">
      <h2 class="font-light text-lg my-4">
        {{ $t('folders') }}
      </h2>
      <MultiColumnList
        :items="subfolders"
        :sort-and-group-by="item => item.title || item.name"
        :group-if-more-than="10"
      >
        <template #item="{ item }">
          <ULink :to="`/${item.url}`" class="inline-flex items-center gap-1">
            <UIcon name="tabler:folder" /> <span>{{ item.title || item.name }}</span>
          </ULink>
        </template>
      </MultiColumnList>
    </div>

    <div v-if="pages.length > 0">
      <h2 class="font-light text-lg my-4">
        {{ $t('pages') }}
      </h2>
      <MultiColumnList
        :items="pages"
        :sort-and-group-by="item => item.title || item.name"
        :group-if-more-than="10"
      >
        <template #item="{ item }">
          <ULink :to="`/${item.url}`" class="inline-flex items-center gap-1">
            <UIcon name="tabler:file-text" /> <span>{{ item.title || item.name }}</span>
          </ULink>
        </template>
      </MultiColumnList>
    </div>

    <PlainDialog ref="plainDialog" />

    <!-- Move Modal -->
    <MoveContentModal
      v-model:open="moveModalOpen"
      :current-path="urlPath"
      :is-folder="true"
      @moved="onFolderMoved"
    />

    <!-- Edit Folder Modal -->
    <UModal v-model:open="editFolderOpen">
      <template #title>
        {{ $t('edit-folder') }}
      </template>
      <template #body>
        <form id="editFolderForm" class="space-y-4" @submit.prevent="saveEditedFolder">
          <div>
            <label class="block text-sm font-medium mb-1">{{ $t('folder-title') }}</label>
            <UInput v-model="editableTitle" class="w-full" autofocus />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">{{ $t('folder-name') }}</label>
            <UInput
              v-model="editableName"
              class="w-full"
              :status="editableName && !isValidFolderName ? 'error' : undefined"
            />
            <p v-if="editableName && !isValidFolderName" class="text-sm text-red-500 mt-1">
              {{ $t('invalid-folder-name') }}
            </p>
          </div>
        </form>
      </template>
      <template #footer>
        <UButton :label="$t('cancel')" @click="editFolderOpen = false" />
        <UButton
          color="primary"
          variant="solid"
          :label="$t('ok')"
          type="submit"
          form="editFolderForm"
          :disabled="!isValidFolderName"
        />
      </template>
    </UModal>
  </Layout>
</template>
