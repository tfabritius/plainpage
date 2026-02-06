<script setup lang="ts">
import type { Breadcrumb, Page } from '~/types'
import { useRouteQuery } from '@vueuse/router'

const props = defineProps<{
  urlPath: string
  breadcrumbs: Breadcrumb[]
  allowCreate: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

const { t } = useI18n()
const toast = useToast()

const emptyPage: Page = {
  url: '',
  content: '',
  meta: {
    title: '',
    tags: [],
    modifiedAt: '0001-01-01T00:00:00Z',
    modifiedByUsername: '',
    modifiedByDisplayName: '',
  },
}
const editablePage = ref<Page>(deepClone(emptyPage))

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === null && props.allowCreate },
  set(value) {
    editQuery.value = value ? null : undefined
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

  toast.add({
    description: t('folder-created'),
    color: 'success',
  })
  emit('refresh')
}

async function onSavePage() {
  try {
    await apiFetch(`/pages/${props.urlPath}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    toast.add({
      description: t('saved'),
      color: 'success',
    })
    emit('refresh')
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

const plainDialog = useTemplateRef('plainDialog')

async function onCancelEdit() {
  if (!deepEqual(emptyPage, editablePage.value)) {
    if (!await plainDialog.value?.confirm(
      t('discard-changes-to-this-page'),
      {
        confirmButtonColor: 'warning',
      },
    )) {
      return
    }
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

    <template v-if="editing" #actions>
      <ReactiveButton icon="tabler:x" :label="$t('cancel')" @click="onCancelEdit" />
      <ReactiveButton icon="tabler:device-floppy" :label="$t('save')" color="success" @click="onSavePage" />
    </template>

    <div v-if="!editing" class="text-center">
      <span class="text-3xl">😟</span>
      <div class="font-medium m-4">
        {{ $t('this-page-doesnt-exist') }}
      </div>

      <div v-if="allowCreate">
        <ReactiveButton icon="tabler:file-plus" :label="$t('create-page')" @click="createThisPage" />
        <ReactiveButton icon="tabler:folder-plus" :label="$t('create-folder')" class="ml-3" @click="createThisFolder" />
      </div>

      <ReactiveButton v-else icon="tabler:home" :label="$t('back-home')" @click="navTo('/')" />
    </div>

    <PageEditor v-else v-model="editablePage" />

    <PlainDialog ref="plainDialog" />
  </Layout>
</template>
