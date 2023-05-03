<script setup lang="ts">
import { useRouteQuery } from '@vueuse/router'
import { storeToRefs } from 'pinia'
import type { Breadcrumb, Page } from '~/types'
import { useAppStore } from '~/store/app'
import { Icon } from '#components'

const props = defineProps<{
  page: Page
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
  onReload: Function
}>()

const route = useRoute()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const page = computed(() => props.page)
const breadcrumbs = computed(() => props.breadcrumbs)

const allowWrite = computed(() => props.allowWrite)
const allowDelete = computed(() => props.allowDelete)

const pageTitle = computed(() => {
  return page.value.meta.title || 'Untitled'
})

useHead(() => ({ title: pageTitle.value }))

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref(deepClone(emptyPage))

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === 'true' && allowWrite.value },
  set(value) {
    editQuery.value = value ? 'true' : null
  },
})

watch(editing, (editing) => {
  if (editing) {
    editablePage.value = deepClone(page.value)
  }
}, { immediate: true })

const PermissionsIcon = h(Icon, { name: 'ci:shield' })
const RevisionsIcon = h(Icon, { name: 'ic:baseline-restore' })
const DeleteIcon = h(Icon, { name: 'ci:trash-full' })
const ReloadIcon = h(Icon, { name: 'ci:arrows-reload-01' })

const onEditPage = () => {
  editing.value = true
}

const onSavePage = async () => {
  try {
    await apiFetch(`/pages${editablePage.value.url}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    ElMessage({
      message: 'Saved',
      type: 'success',
    })

    await props.onReload()
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const deleteConfirmOpen = ref(false)
const onDeletePage = async () => {
  if (deleteConfirmOpen.value) {
    // Prevent multiple dialogs at the same time
    return
  }

  deleteConfirmOpen.value = true
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
    deleteConfirmOpen.value = false
    return
  }
  deleteConfirmOpen.value = false

  try {
    await apiFetch(`/pages${route.path}`, { method: 'DELETE' })

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
    await props.onReload()
    ElMessage({ message: 'Page reloaded', type: 'success' })
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

onKeyStroke('e', (e) => {
  if (!editing.value && allowWrite.value) {
    e.preventDefault()
    onEditPage()
  }
})

onKeyStroke('Delete', (e) => {
  if (!editing.value && allowDelete.value) {
    e.preventDefault()
    onDeletePage()
  }
})

onKeyStroke('Escape', async (_event: KeyboardEvent) => {
  if (editing.value) {
    onCancelEdit()
  }
}, { eventName: 'keyup' })
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <span v-if="page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">
        {{ pageTitle }}
      </span>
    </template>

    <template #actions>
      <div v-if="!editing">
        <ElButton v-if="allowWrite" class="m-1" @click="onEditPage">
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
              <ElDropdownItem :icon="RevisionsIcon" command="rev">
                Revisions
              </ElDropdownItem>
              <ElDropdownItem v-if="allowAdmin" :icon="PermissionsIcon" command="acl">
                Permissions
              </ElDropdownItem>
              <ElDropdownItem v-if="allowDelete" :icon="DeleteIcon" command="delete">
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

    <PageView v-if="!editing" :page="page" />
    <PageEditor v-else v-model="editablePage" />
  </Layout>
</template>
