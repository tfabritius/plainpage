<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { Breadcrumb, Page } from '~/types'
import { useRouteQuery } from '@vueuse/router'
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'

const props = defineProps<{
  page: Page
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
  onReload: () => void
}>()

// Regex for valid page names (matches backend validation)
const validUrlPartRegex = /^[a-z0-9-][a-z0-9_-]*$/

const { t } = useI18n()
const toast = useToast()

const route = useRoute()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const pageTitle = computed(() => props.page.meta.title || t('untitled'))

useHead(() => ({ title: pageTitle.value }))

const plainDialog = useTemplateRef('plainDialog')

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref(deepClone(emptyPage))

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === null && props.allowWrite },
  set(value) {
    editQuery.value = value ? null : undefined
  },
})

// Rename functionality
const renameModalOpen = ref(false)
const newPageName = ref('')
const currentPageName = computed(() => {
  const urlParts = props.page.url.split('/')
  return urlParts[urlParts.length - 1] || ''
})
const parentPath = computed(() => {
  const urlParts = props.page.url.split('/')
  urlParts.pop()
  return urlParts.join('/')
})

function openRenameModal() {
  newPageName.value = currentPageName.value
  renameModalOpen.value = true
}

const isValidPageName = computed(() => validUrlPartRegex.test(newPageName.value))

async function onRenamePage() {
  if (!isValidPageName.value) {
    return
  }

  const newUrl = parentPath.value ? `${parentPath.value}/${newPageName.value}` : newPageName.value

  if (newUrl === props.page.url) {
    renameModalOpen.value = false
    return
  }

  try {
    await apiFetch(`/pages/${props.page.url}`, {
      method: 'PATCH',
      body: [{ op: 'replace', path: '/page/url', value: newUrl }],
    })

    renameModalOpen.value = false

    toast.add({
      description: t('page-renamed'),
      color: 'success',
    })

    await navigateTo(`/${newUrl}`)
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

watch(editing, (editing) => {
  if (editing) {
    editablePage.value = deepClone(props.page)
  }
}, { immediate: true })

function onEditPage() {
  editing.value = true
}

async function onSavePage() {
  try {
    await apiFetch(`/pages/${editablePage.value.url}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    toast.add({
      description: t('saved'),
      color: 'success',
    })

    await props.onReload()
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

const deleteConfirmOpen = ref(false)
async function onDeletePage() {
  if (deleteConfirmOpen.value) {
    // Prevent multiple dialogs at the same time
    return
  }

  deleteConfirmOpen.value = true

  if (!await plainDialog.value?.confirm(
    t('are-you-sure-to-delete-this-page'),
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
    await apiFetch(`/pages${route.path}`, { method: 'DELETE' })

    toast.add({
      description: t('page-deleted'),
      color: 'success',
    })

    await navigateTo(route.path.substring(0, route.path.lastIndexOf('/') + 1))
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

const menuItems = computed(() => {
  const items: DropdownMenuItem[] = []

  items.push({
    icon: 'ci:arrows-reload-01',
    label: t('reload'),
    onSelect: async () => {
      await props.onReload()
      toast.add({ description: t('page-reloaded'), color: 'success' })
    },
  })

  items.push({
    icon: 'ic:baseline-restore',
    label: t('revisions'),
    onSelect: async () => {
      await navigateTo({ query: { rev: null } })
    },
  })

  if (props.allowWrite && props.allowDelete) {
    items.push({
      icon: 'ci:edit-pencil-line-01',
      label: t('rename'),
      onSelect: openRenameModal,
    })
  }

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

  if (props.allowDelete) {
    items.push(
      {
        icon: 'ci:trash-full',
        label: t('delete'),
        onSelect: onDeletePage,
      },
    )
  }

  return items
})

async function onCancelEdit() {
  if (!deepEqual(props.page, editablePage.value)) {
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

onKeyStroke('e', (e) => {
  if (!editing.value && props.allowWrite && !renameModalOpen.value) {
    e.preventDefault()
    onEditPage()
  }
})

onKeyStroke('Backspace', (e) => {
  if (!editing.value && props.allowDelete && e.ctrlKey && !renameModalOpen.value) {
    e.preventDefault()
    onDeletePage()
  }
})

onKeyStroke('s', (e) => {
  if (editing.value && e.ctrlKey && !renameModalOpen.value) {
    e.preventDefault()
    onSavePage()
  }
})
</script>

<template>
  <Layout
    :breadcrumbs="breadcrumbs"
    :use-full-height="editing"
  >
    <template #title>
      <span v-if="page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">
        {{ pageTitle }}
      </span>
    </template>

    <template #actions>
      <ReactiveButton v-if="!editing && allowWrite" icon="ci:edit" :label="$t('edit')" @click="onEditPage" />

      <UDropdownMenu v-if="!editing" :items="menuItems">
        <ReactiveButton icon="ci:more-vertical" :label="$t('more')" />
      </UDropdownMenu>

      <ReactiveButton v-if="editing" icon="ci:close-md" :label="$t('cancel')" @click="onCancelEdit" />
      <ReactiveButton v-if="editing" color="success" icon="ci:save" :label="$t('save')" @click="onSavePage" />
    </template>

    <PageView
      v-if="!editing"
      :page="page"
    />
    <PageEditor
      v-else
      v-model="editablePage"
      @escape="onCancelEdit"
    />

    <PlainDialog ref="plainDialog" />

    <!-- Rename Modal -->
    <UModal v-model:open="renameModalOpen">
      <template #title>
        {{ $t('rename-page') }}
      </template>
      <template #body>
        <form id="renamePageForm" @submit.prevent="onRenamePage">
          <UInput
            v-model="newPageName"
            class="w-full"
            autofocus
            :status="newPageName && !isValidPageName ? 'error' : undefined"
          />
          <p v-if="newPageName && !isValidPageName" class="text-sm text-red-500 mt-1">
            {{ $t('invalid-page-name') }}
          </p>
        </form>
      </template>
      <template #footer>
        <UButton :label="$t('cancel')" @click="renameModalOpen = false" />
        <UButton
          color="primary"
          variant="solid"
          :label="$t('ok')"
          type="submit"
          form="renamePageForm"
          :disabled="!isValidPageName"
        />
      </template>
    </UModal>
  </Layout>
</template>
