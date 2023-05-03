<script setup lang="ts">
import { useRouteQuery } from '@vueuse/router'
import type { Breadcrumb, Page } from '~/types'

const props = defineProps<{
  urlPath: string
  breadcrumbs: Breadcrumb[]
  allowCreate: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

const urlPath = computed(() => props.urlPath)
const allowCreate = computed(() => props.allowCreate)
const breadcrumbs = computed(() => props.breadcrumbs)

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref<Page>(deepClone(emptyPage))

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === 'true' && allowCreate.value },
  set(value) {
    editQuery.value = value ? 'true' : null
  },
})

useHead({ title: 'Not found' })

const createThisPage = () => {
  editablePage.value = deepClone(emptyPage)
  editing.value = true
}

const createThisFolder = async () => {
  await apiFetch(`/pages/${urlPath.value}`, { method: 'PUT', body: { page: null } })

  ElMessage({
    message: 'Folder created',
    type: 'success',
  })
  emit('refresh')
}

const onSavePage = async () => {
  try {
    await apiFetch(`/pages/${urlPath.value}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    ElMessage({
      message: 'Saved',
      type: 'success',
    })
    emit('refresh')
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const cancelEditConfirmOpen = ref(false)

const onCancelEdit = async () => {
  if (cancelEditConfirmOpen.value) {
    ElMessageBox.close()
    cancelEditConfirmOpen.value = false
    return
  }

  if (!deepEqual(emptyPage, editablePage.value)) {
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

onKeyStroke('s', (e) => {
  if (editing.value && e.ctrlKey) {
    e.preventDefault()
    onSavePage()
  }
})
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <span class="italic">Not found
      </span>
    </template>

    <template #actions>
      <div v-if="editing">
        <ElButton class="ml-2" @click="onCancelEdit">
          <Icon name="ci:close-md" /> <span class="hidden md:inline ml-1">Cancel</span>
        </ElButton>
        <ElButton type="success" @click="onSavePage">
          <Icon name="ci:save" /> <span class="hidden md:inline ml-1">Save</span>
        </ElButton>
      </div>
    </template>

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
  </Layout>
</template>
