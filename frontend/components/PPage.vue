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

onKeyStroke('Backspace', (e) => {
  if (!editing.value && props.allowDelete && e.ctrlKey) {
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
      <div v-if="!editing">
        <PlainButton v-if="allowWrite" icon="ci:edit" :label="$t('edit')" @click="onEditPage" />

        <ElDropdown trigger="click" class="ml-3" @command="handleDropdownMenuCommand">
          <PlainButton icon="ci:more-vertical" :label="$t('more')" />
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
        <PlainButton class="ml-2" icon="ci:close-md" :label="$t('cancel')" @click="onCancelEdit" />
        <PlainButton type="success" icon="ci:save" :label="$t('save')" @click="onSavePage" />
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
