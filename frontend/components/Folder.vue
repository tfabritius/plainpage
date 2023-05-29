<script setup lang="ts">
import { storeToRefs } from 'pinia'
import slugify from 'slugify'
import type { FormInstance, FormRules } from 'element-plus'

import type { Breadcrumb, Folder } from '~/types'
import { Icon } from '#components'
import { useAppStore } from '~/store/app'

const props = defineProps<{
  urlPath: string
  folder: Folder
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
  onReload: Function
}>()

const { t } = useI18n()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const pageTitle = computed(() => {
  if (props.urlPath === '') {
    return t('home')
  }
  return props.breadcrumbs.slice(-1)[0]?.name
})

useHead(() => ({ title: pageTitle.value }))

const PermissionsIcon = h(Icon, { name: 'ci:shield' })
const DeleteIcon = h(Icon, { name: 'ci:trash-full' })
const ReloadIcon = h(Icon, { name: 'ci:arrows-reload-01' })

const newPageDialogVisible = ref(true)
const newPageFormRef = ref<FormInstance>()
const newPageFormData = ref({ title: '', name: '' })
const newPageDialogExpanded = ref(false)
const toggleNewPageDialogExpanded = useToggle(newPageDialogExpanded)
const newPageNameChangedManually = ref(false)
const newPageFormRules = {
  name: [
    { required: true, message: t('page-name-required'), trigger: 'blur' },
    { pattern: /^[a-z0-9-][a-z0-9_-]*$/, message: t('invalid-page-name'), trigger: 'blur' },
  ],
} satisfies FormRules

async function showNewPageDialog() {
  newPageFormData.value = { title: '', name: '' }
  newPageDialogExpanded.value = false
  newPageNameChangedManually.value = false
  newPageDialogVisible.value = true
}

function onNewPageTitleChanged() {
  if (!newPageNameChangedManually.value) {
    newPageFormData.value.name = slugify(
      newPageFormData.value.title,
      { lower: true, strict: true },
    )
  }
}

function onNewPageNameChanged() {
  newPageNameChangedManually.value = true
}

async function submitNewPageDialog() {
  if (!newPageFormRef.value) {
    return
  }

  const formValid = await new Promise<boolean>(resolve => newPageFormRef.value?.validate(valid => resolve(valid)))
  if (!formValid) {
    newPageDialogExpanded.value = true
    return
  }

  newPageDialogVisible.value = false

  await navigateTo({ path: `${props.urlPath}/${newPageFormData.value.name}`, query: { edit: 'true' }, state: { title: newPageFormData.value.title } })
}

async function createFolder() {
  let name
  try {
    const msgBox = await ElMessageBox.prompt(t('enter-folder-name'), t('create-folder'), {
      confirmButtonText: t('ok'),
      cancelButtonText: t('cancel'),
      inputPattern: /^[a-z0-9-][a-z0-9_-]*$/,
      inputErrorMessage: t('invalid-folder-name'),
    })
    name = msgBox.value
  } catch (e) {
    return
  }

  try {
    await apiFetch(`/pages${props.urlPath}/${name}`, { method: 'PUT', body: { page: null } })

    ElMessage({
      message: t('folder-created'),
      type: 'success',
    })
    await navigateTo(`${props.urlPath}/${name}`)
  } catch (err) {
    ElMessage({
      message: String(err),
      type: 'error',
    })
  }
}

const deleteConfirmOpen = ref(false)
async function onDeleteFolder() {
  if (deleteConfirmOpen.value) {
    // Prevent multiple dialogs at the same time
    return
  }

  deleteConfirmOpen.value = true
  try {
    await ElMessageBox.confirm(
      t('are-you-sure-to-delete-this-folder'),
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
    await apiFetch(`/pages${props.urlPath}`, { method: 'DELETE' })

    ElMessage({
      message: t('folder-deleted'),
      type: 'success',
    })

    await navigateTo('./')
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
    ElMessage({ message: t('folder-reloaded'), type: 'success' })
  } else if (command === 'acl') {
    await navigateTo({ query: { acl: null } })
  } else if (command === 'delete') {
    onDeleteFolder()
  } else {
    throw new Error(`Unhandled command ${command}`)
  }
}

onKeyStroke('Backspace', (e) => {
  if (props.urlPath !== '' && props.allowDelete && e.ctrlKey) {
    e.preventDefault()
    onDeleteFolder()
  }
})
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <Icon name="ci:folder" class="mr-1" />
      {{ pageTitle }}
    </template>

    <template #actions>
      <div>
        <PlainButton v-if="allowWrite" icon="ci:file-add" :label="$t('create-page')" @click="showNewPageDialog" />
        <PlainButton v-if="allowWrite" icon="ci:folder-add" :label="$t('create-folder')" @click="createFolder" />

        <ElDropdown trigger="click" class="ml-3" @command="handleDropdownMenuCommand">
          <PlainButton icon="ci:more-vertical" :label="$t('more')" />
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="ReloadIcon" command="reload">
                {{ $t('reload') }}
              </ElDropdownItem>
              <ElDropdownItem v-if="allowAdmin" :icon="PermissionsIcon" command="acl">
                {{ $t('permissions') }}
              </ElDropdownItem>
              <ElDropdownItem v-if="urlPath !== '' && allowDelete" :icon="DeleteIcon" command="delete">
                {{ $t('delete') }}
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>
    </template>

    <div>
      <h2
        v-if="folder.content.some(e => e.isFolder)"
        class="font-light text-xl"
      >
        {{ $t('folders') }}
      </h2>
      <div v-for="entry of folder.content.filter(e => e.isFolder)" :key="entry.name">
        <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
          <ElLink :href="href" @click="navigate">
            <Icon name="ci:folder" class="mr-1" /> {{ entry.name }}
          </ElLink>
        </NuxtLink>
      </div>

      <h2
        v-if="folder.content.some(e => !e.isFolder)"
        class="font-light text-xl"
      >
        {{ $t('pages') }}
      </h2>
      <div v-for="entry of folder.content.filter(e => !e.isFolder)" :key="entry.name">
        <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
          <ElLink :href="href" @click="navigate">
            {{ entry.name }}
          </ElLink>
        </NuxtLink>
      </div>
    </div>

    <ClientOnly>
      <ElDialog
        v-model="newPageDialogVisible"
        :title="$t('create-page')"
        width="40%"
      >
        <ElForm
          ref="newPageFormRef"
          :model="newPageFormData"
          :rules="newPageFormRules"
          label-position="top"
          @submit.prevent
          @keypress.enter="submitNewPageDialog"
        >
          <ElFormItem :label="$t('page-title')" prop="title">
            <ElInput v-model="newPageFormData.title" @input="onNewPageTitleChanged" />
          </ElFormItem>

          <span class="cursor-pointer text-xs" @click="toggleNewPageDialogExpanded()">
            <Icon name="ci:chevron-right" :class="newPageDialogExpanded && 'rotate-90 transform'" />
            {{ $t('more') }}
          </span>
          <div v-show="newPageDialogExpanded">
            <ElFormItem :label="$t('page-name')" prop="name">
              <ElInput v-model="newPageFormData.name" @input="onNewPageNameChanged" />
            </ElFormItem>
          </div>
        </ElForm>
        <template #footer>
          <PlainButton :label="$t('cancel')" @click="newPageDialogVisible = false" />
          <PlainButton type="primary" :label="$t('ok')" @click="submitNewPageDialog" />
        </template>
      </ElDialog>
    </ClientOnly>
  </Layout>
</template>
