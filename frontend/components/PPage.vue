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

const { t } = useI18n()

const route = useRoute()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const pageTitle = computed(() => props.page.meta.title || t('untitled'))

useHead(() => ({ title: pageTitle.value }))

const emptyPage: Page = { url: '', content: '', meta: { title: '', tags: [] } }
const editablePage = ref(deepClone(emptyPage))

const editQuery = useRouteQuery('edit')
const editing = computed({
  get() { return editQuery.value === 'true' && props.allowWrite },
  set(value) {
    editQuery.value = value ? 'true' : null
  },
})

watch(editing, (editing) => {
  if (editing) {
    editablePage.value = deepClone(props.page)
  }
}, { immediate: true })

const PermissionsIcon = h(Icon, { name: 'ci:shield' })
const RevisionsIcon = h(Icon, { name: 'ic:baseline-restore' })
const DeleteIcon = h(Icon, { name: 'ci:trash-full' })
const ReloadIcon = h(Icon, { name: 'ci:arrows-reload-01' })

function onEditPage() {
  editing.value = true
}

async function onSavePage() {
  try {
    await apiFetch(`/pages${editablePage.value.url}`, { method: 'PUT', body: { page: editablePage.value } })
    editing.value = false

    ElMessage({
      message: t('saved'),
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
async function onDeletePage() {
  if (deleteConfirmOpen.value) {
    // Prevent multiple dialogs at the same time
    return
  }

  deleteConfirmOpen.value = true
  try {
    await ElMessageBox.confirm(
      t('are-you-sure-to-delete-this-page'),
      {
        confirmButtonText: t('delete'),
        cancelButtonText: t('cancel'),
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
      message: t('page-deleted'),
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

async function handleDropdownMenuCommand(command: string | number | object) {
  if (command === 'reload') {
    await props.onReload()
    ElMessage({ message: t('page-reloaded'), type: 'success' })
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

async function onCancelEdit() {
  if (cancelEditConfirmOpen.value) {
    ElMessageBox.close()
    cancelEditConfirmOpen.value = false
    return
  }

  if (!deepEqual(props.page, editablePage.value)) {
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

onKeyStroke('e', (e) => {
  if (!editing.value && props.allowWrite) {
    e.preventDefault()
    onEditPage()
  }
})

onKeyStroke('Delete', (e) => {
  if (!editing.value && props.allowDelete) {
    e.preventDefault()
    onDeletePage()
  }
})

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
      <span v-if="page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">
        {{ pageTitle }}
      </span>
    </template>

    <template #actions>
      <div v-if="!editing">
        <ElButton v-if="allowWrite" class="m-1" @click="onEditPage">
          <Icon name="ci:edit" /> <span class="hidden md:inline ml-1">{{ $t('edit') }}</span>
        </ElButton>

        <ElDropdown trigger="click" class="m-1" @command="handleDropdownMenuCommand">
          <ElButton>
            <Icon name="ci:more-vertical" /> <span class="hidden md:inline ml-1">{{ $t('more') }}</span>
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="ReloadIcon" command="reload">
                {{ $t('reload') }}
              </ElDropdownItem>
              <ElDropdownItem :icon="RevisionsIcon" command="rev">
                {{ $t('revisions') }}
              </ElDropdownItem>
              <ElDropdownItem v-if="allowAdmin" :icon="PermissionsIcon" command="acl">
                {{ $t('permissions') }}
              </ElDropdownItem>
              <ElDropdownItem v-if="allowDelete" :icon="DeleteIcon" command="delete">
                {{ $t('delete') }}
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>

      <div v-if="editing">
        <ElButton class="ml-2" @click="onCancelEdit">
          <Icon name="ci:close-md" /> <span class="hidden md:inline ml-1">{{ $t('cancel') }}</span>
        </ElButton>
        <ElButton type="success" @click="onSavePage">
          <Icon name="ci:save" /> <span class="hidden md:inline ml-1">{{ $t('save') }}</span>
        </ElButton>
      </div>
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
  </Layout>
</template>
