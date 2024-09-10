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

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref<Page>(deepClone(emptyPage))

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === 'true' && props.allowCreate },
  set(value) {
    editQuery.value = value ? 'true' : null
  },
})

useHead(() => ({ title: t('not-found') }))

onMounted(() => {
  // Take new page's title from state if available.
  // Creating new page via dialog will set this value.
  if (window.history.state.title) {
    editablePage.value.meta.title = window.history.state.title
  }
})

function createThisPage() {
  editablePage.value = deepClone(emptyPage)
  editing.value = true
}

async function createThisFolder() {
  await apiFetch(`/pages/${props.urlPath}`, { method: 'PUT', body: { page: null } })

  ElMessage({
    message: t('folder-created'),
    type: 'success',
  })
  emit('refresh')
}

async function onSavePage() {
  try {
    await apiFetch(`/pages/${props.urlPath}`, { method: 'PUT', body: { page: editablePage.value } })
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

const navTo = navigateTo
</script>

<template>
  <Layout
    :breadcrumbs="breadcrumbs"
    :use-full-height="editing"
  >
    <template #title>
      <span class="italic">{{ $t('not-found') }}</span>
    </template>

    <template #actions>
      <div v-if="editing">
        <PlainButton icon="ci:close-md" :label="$t('cancel')" @click="onCancelEdit" />
        <PlainButton icon="ci:save" :label="$t('save')" type="success" @click="onSavePage" />
      </div>
    </template>

    <div v-if="!editing" class="text-center">
      <span class="text-3xl">😟</span>
      <div class="font-medium m-4">
        {{ $t('this-page-doesnt-exist') }}
      </div>

      <div v-if="allowCreate">
        <PlainButton icon="ci:file-add" :label="$t('create-page')" @click="createThisPage" />
        <PlainButton icon="ci:folder-add" :label="$t('create-folder')" @click="createThisFolder" />
      </div>

      <PlainButton v-else icon="ic:outline-home" :label="$t('back-home')" @click="navTo('/')" />
    </div>

    <PageEditor v-else v-model="editablePage" />
  </Layout>
</template>
