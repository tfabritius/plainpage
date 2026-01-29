<script setup lang="ts">
import type { BreadcrumbItem } from '@nuxt/ui'
import type { Folder, FolderEntry, GetContentResponse } from '~/types'

const props = defineProps<{
  /** Current URL path of the content being moved */
  currentPath: string
  /** Whether moving a folder (true) or page (false) */
  isFolder: boolean
}>()

const emit = defineEmits<{
  moved: [newPath: string]
}>()

const open = defineModel<boolean>('open', { default: false })

const { t } = useI18n()
const toast = useToast()

// State for folder navigation
const currentFolderPath = ref('')
const folderData = ref<Folder | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const allowWriteInCurrentFolder = ref(false)

// Compute the item name from the current path
const itemName = computed(() => {
  const parts = props.currentPath.split('/')
  return parts[parts.length - 1] || ''
})

// Compute the current parent path
const currentParentPath = computed(() => {
  const parts = props.currentPath.split('/')
  parts.pop()
  return parts.join('/')
})

// Compute the destination path if moving to current folder
const destinationPath = computed(() => {
  if (currentFolderPath.value === '') {
    return itemName.value
  }
  return `${currentFolderPath.value}/${itemName.value}`
})

// Check if destination is same as current location
const isSameLocation = computed(() => {
  return currentFolderPath.value === currentParentPath.value
})

// Check if trying to move folder into itself or its descendants
const isMovingIntoSelf = computed(() => {
  if (!props.isFolder) {
    return false
  }
  // Cannot move to the folder itself
  if (currentFolderPath.value === props.currentPath) {
    return true
  }
  // Cannot move to a descendant
  if (currentFolderPath.value.startsWith(`${props.currentPath}/`)) {
    return true
  }
  return false
})

// Check if move is allowed
const canMove = computed(() => {
  return allowWriteInCurrentFolder.value && !isSameLocation.value && !isMovingIntoSelf.value
})

// Build breadcrumbs for the folder picker
const pickerBreadcrumbItems = computed(() => {
  const items: BreadcrumbItem[] = []

  // Root folder item
  const isAtRoot = currentFolderPath.value === ''
  items.push({
    icon: 'ic:outline-home',
    class: isAtRoot ? 'text-[var(--ui-text-muted)] pointer-events-none' : 'cursor-pointer',
    onClick: isAtRoot ? undefined : () => loadFolder(''),
  })

  if (currentFolderPath.value) {
    const parts = currentFolderPath.value.split('/')
    let path = ''
    for (const [i, part] of parts.entries()) {
      path = path ? `${path}/${part}` : part
      const isLast = i === parts.length - 1
      const itemPath = path // Capture path for closure
      items.push({
        label: part,
        class: isLast ? 'text-[var(--ui-text-muted)] pointer-events-none' : 'cursor-pointer',
        onClick: isLast ? undefined : () => loadFolder(itemPath),
      })
    }
  }

  return items
})

// Compute parent folder path for "go up" entry
const parentFolderPath = computed(() => {
  if (currentFolderPath.value === '') {
    return null // Already at root
  }
  const parts = currentFolderPath.value.split('/')
  parts.pop()
  return parts.join('/')
})

// Get only folders from current folder data
const subfolders = computed(() => {
  if (!folderData.value) {
    return []
  }
  return folderData.value.content.filter((e: FolderEntry) => e.isFolder)
})

// Load folder contents
async function loadFolder(path: string) {
  loading.value = true
  error.value = null

  try {
    const response = await apiFetch<GetContentResponse>(`/pages/${path}`)
    folderData.value = response.folder
    allowWriteInCurrentFolder.value = response.allowWrite
    currentFolderPath.value = path
  } catch (err) {
    error.value = String(err)
    folderData.value = null
  } finally {
    loading.value = false
  }
}

// Navigate into a folder
function navigateToFolder(path: string) {
  loadFolder(path)
}

// Perform the move operation
async function performMove() {
  if (!canMove.value) {
    return
  }

  try {
    const patchPath = props.isFolder ? '/folder/url' : '/page/url'
    await apiFetch(`/pages/${props.currentPath}`, {
      method: 'PATCH',
      body: [{ op: 'replace', path: patchPath, value: destinationPath.value }],
    })

    toast.add({
      description: props.isFolder ? t('folder-moved') : t('page-moved'),
      color: 'success',
    })

    open.value = false
    emit('moved', destinationPath.value)
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

// Load initial folder when modal opens
watch(open, (isOpen) => {
  if (isOpen) {
    // Start at the current parent folder
    loadFolder(currentParentPath.value)
  }
})
</script>

<template>
  <UModal v-model:open="open">
    <template #title>
      {{ isFolder ? $t('move-folder') : $t('move-page') }}
    </template>

    <template #body>
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ $t('select-destination-folder') }}
        </p>

        <!-- Breadcrumbs -->
        <UBreadcrumb
          :items="pickerBreadcrumbItems"
          :ui="{ link: 'hover:text-[var(--ui-primary)]' }"
        />

        <!-- Loading state -->
        <div v-if="loading" class="flex items-center justify-center py-8">
          <UIcon name="ci:arrows-reload-01" class="animate-spin text-2xl" />
        </div>

        <!-- Error state -->
        <div v-else-if="error" class="text-red-500 py-4">
          {{ error }}
        </div>

        <!-- Folder list -->
        <div v-else class="border rounded-md dark:border-gray-700 max-h-64 overflow-y-auto">
          <!-- Parent folder entry (go up) -->
          <button
            v-if="parentFolderPath !== null"
            type="button"
            class="w-full flex items-center gap-2 px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-800 text-left border-b dark:border-gray-700 cursor-pointer"
            @click="navigateToFolder(parentFolderPath)"
          >
            <UIcon name="ci:folder-upload" class="text-lg flex-shrink-0" />
            <span>..</span>
          </button>

          <!-- Empty folder -->
          <div v-if="subfolders.length === 0" class="p-4 text-center text-gray-500">
            {{ $t('no-subfolders') }}
          </div>

          <!-- Folder entries -->
          <button
            v-for="folder in subfolders"
            :key="folder.url"
            type="button"
            class="w-full flex items-center gap-2 px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-800 text-left border-b last:border-b-0 dark:border-gray-700 cursor-pointer"
            :class="{
              'opacity-50': folder.url === props.currentPath,
            }"
            :disabled="folder.url === props.currentPath"
            @click="navigateToFolder(folder.url)"
          >
            <UIcon name="ci:folder" class="text-lg flex-shrink-0" />
            <span class="truncate">{{ folder.title || folder.name }}</span>
            <UIcon name="ci:chevron-right" class="ml-auto text-gray-400 flex-shrink-0" />
          </button>
        </div>

        <!-- Destination preview -->
        <div class="text-sm">
          <span class="text-gray-500">{{ $t('destination') }}:</span>
          <code class="ml-2 px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded">
            {{ `/${destinationPath}` }}
          </code>
        </div>

        <!-- Warning messages -->
        <div v-if="isSameLocation" class="text-amber-600 dark:text-amber-400 text-sm flex items-center gap-2">
          <UIcon name="ci:info" />
          {{ $t('cannot-move-to-same-location') }}
        </div>
        <div v-else-if="isMovingIntoSelf" class="text-amber-600 dark:text-amber-400 text-sm flex items-center gap-2">
          <UIcon name="ci:info" />
          {{ $t('cannot-move-folder-into-itself') }}
        </div>
        <div v-else-if="!allowWriteInCurrentFolder" class="text-amber-600 dark:text-amber-400 text-sm flex items-center gap-2">
          <UIcon name="ci:info" />
          {{ $t('no-write-permission') }}
        </div>
      </div>
    </template>

    <template #footer>
      <UButton :label="$t('cancel')" @click="open = false" />
      <UButton
        color="primary"
        variant="solid"
        :label="$t('move-here')"
        :disabled="!canMove"
        @click="performMove"
      />
    </template>
  </UModal>
</template>
