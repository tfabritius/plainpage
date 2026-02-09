<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { Breadcrumb, Page } from '~/types'
import { useTimeAgo } from '@vueuse/core'
import { useRouteQuery } from '@vueuse/router'
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'
import { validUrlPartRegex } from '~/types/api'

const props = defineProps<{
  page: Page
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
  onReload: () => void
}>()

const { t } = useI18n()
const toast = useToast()

const route = useRoute()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const pageTitle = computed(() => props.page.meta.title || t('untitled'))

useHead(() => ({ title: pageTitle.value }))

// Modified info - check for non-zero time (Go's zero time starts with year 0001)
const hasModifiedAt = computed(() => props.page.meta.modifiedAt && !props.page.meta.modifiedAt.startsWith('0001-'))
const modifiedAt = computed(() => new Date(props.page.meta.modifiedAt ?? ''))
const modifiedAtTimeAgo = useTimeAgo(modifiedAt, { messages: timeAgoMessages() })
const modifiedAtFormatted = computed(() => modifiedAt.value.toLocaleString())

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

// Move functionality
const moveModalOpen = ref(false)
async function onPageMoved(newPath: string) {
  await navigateTo(`/${newPath}`)
}

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
      title: t('delete-page'),
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
    icon: 'tabler:refresh',
    label: t('reload'),
    onSelect: async () => {
      await props.onReload()
      toast.add({ description: t('page-reloaded'), color: 'success' })
    },
  })

  items.push({
    icon: 'tabler:restore',
    label: t('revisions'),
    onSelect: async () => {
      await navigateTo({ query: { rev: null } })
    },
  })

  if (props.allowWrite && props.allowDelete) {
    items.push({
      icon: 'tabler:pencil',
      label: t('rename'),
      onSelect: openRenameModal,
    })

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

  if (props.allowDelete) {
    items.push(
      {
        icon: 'tabler:trash',
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
        title: t('discard-changes'),
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

    <template v-if="!editing && hasModifiedAt" #subtitle>
      <span class="text-[var(--ui-text-muted)]/50">
        {{ $t('modified') }} <span :title="modifiedAtFormatted">{{ modifiedAtTimeAgo }}</span> {{ $t('modified-by') }}
        <span v-if="page.meta.modifiedByUsername" :title="page.meta.modifiedByUsername">{{ page.meta.modifiedByDisplayName }}</span>
        <span v-else class="italic">{{ $t('anonymous') }}</span>
      </span>
    </template>

    <template #actions>
      <UTooltip v-if="!editing && allowWrite" :text="$t('edit')" :kbds="['E']">
        <ReactiveButton icon="tabler:edit" :label="$t('edit')" @click="onEditPage" />
      </UTooltip>

      <UDropdownMenu v-if="!editing" :items="menuItems">
        <ReactiveButton icon="tabler:dots-vertical" :label="$t('more')" />
      </UDropdownMenu>

      <UTooltip v-if="editing" :text="$t('cancel')" :kbds="['Esc']">
        <ReactiveButton icon="tabler:x" :label="$t('cancel')" @click="onCancelEdit" />
      </UTooltip>
      <UTooltip v-if="editing" :text="$t('save')" :kbds="['meta', 'S']">
        <ReactiveButton color="success" icon="tabler:device-floppy" :label="$t('save')" @click="onSavePage" />
      </UTooltip>
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

    <!-- Move Modal -->
    <MoveContentModal
      v-model:open="moveModalOpen"
      :current-path="page.url"
      :is-folder="false"
      @moved="onPageMoved"
    />

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
