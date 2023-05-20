<script setup lang="ts">
import { useRouteQuery } from '@vueuse/router'
import type { Breadcrumb, Page } from '~/types'

const props = defineProps<{
  urlPath: string
  breadcrumbs: Breadcrumb[]
  allowCreate: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

const { t } = useI18n()

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

useHead(() => ({ title: t('not-found') }))

function createThisPage() {
  editablePage.value = deepClone(emptyPage)
  editing.value = true
}

async function createThisFolder() {
  await apiFetch(`/pages/${urlPath.value}`, { method: 'PUT', body: { page: null } })

  ElMessage({
    message: 'Folder created',
    type: 'success',
  })
  emit('refresh')
}

async function onSavePage() {
  try {
    await apiFetch(`/pages/${urlPath.value}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    ElMessage({
      message: t('saved'),
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

async function onCancelEdit() {
  if (cancelEditConfirmOpen.value) {
    ElMessageBox.close()
    cancelEditConfirmOpen.value = false
    return
  }

  if (!deepEqual(emptyPage, editablePage.value)) {
    try {
      cancelEditConfirmOpen.value = true
      await ElMessageBox.confirm(t('discard-changes-to-this-page'), {
        confirmButtonText: t('ok'),
        cancelButtonText: t('cancel'),
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
      <span class="italic">{{ $t('not-found') }}</span>
    </template>

    <template #actions>
      <div v-if="editing">
        <ElButton class="ml-2" @click="onCancelEdit">
          <Icon name="ci:close-md" /> <span class="hidden md:inline ml-1">{{ $t('cancel') }}</span>
        </ElButton>
        <ElButton type="success" @click="onSavePage">
          <Icon name="ci:save" /> <span class="hidden md:inline ml-1">{{ $t('save') }}</span>
        </ElButton>
      </div>
    </template>

    <div v-if="!editing" class="text-center">
      <span class="text-3xl">😟</span>
      <div class="font-medium m-4">
        {{ $t('this-page-doesnt-exist') }}
      </div>

      <div v-if="allowCreate">
        <ElButton @click="createThisPage">
          <Icon name="ci:file-add" /> <span class="hidden md:inline ml-1">{{ $t('create-page') }}</span>
        </ElButton>
        <ElButton @click="createThisFolder">
          <Icon name="ci:folder-add" /> <span class="hidden md:inline ml-1">{{ $t('create-folder') }}</span>
        </ElButton>
      </div>

      <ElButton v-else @click="navigateTo('/')">
        <Icon name="ic:outline-home" /> <span class="hidden md:inline ml-1">{{ $t('back-home') }}</span>
      </ElButton>
    </div>

    <div v-else>
      <PageEditor v-model="editablePage" />
    </div>
  </Layout>
</template>
