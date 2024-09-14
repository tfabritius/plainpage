<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { Icon } from '#components'
import { ElInput } from 'element-plus'
import { storeToRefs } from 'pinia'

import slugify from 'slugify'
import { useAppStore } from '~/store/app'
import type { Breadcrumb, Folder, PutRequest } from '~/types'

const props = defineProps<{
  urlPath: string
  folder: Folder
  breadcrumbs: Breadcrumb[]
  allowWrite: boolean
  allowDelete: boolean
  onReload: () => void
}>()

const subfolders = computed(() => props.folder.content.filter(e => e.isFolder))
const pages = computed(() => props.folder.content.filter(e => !e.isFolder))

const { t } = useI18n()

const app = useAppStore()
const { allowAdmin } = storeToRefs(app)

const pageTitle = computed(() => {
  if (props.urlPath === '') {
    return t('home')
  }
  return props.folder.meta.title || props.breadcrumbs.slice(-1)[0]?.name
})

useHead(() => ({ title: pageTitle.value }))

const PermissionsIcon = h(Icon, { name: 'ci:shield' })
const DeleteIcon = h(Icon, { name: 'ci:trash-full' })
const ReloadIcon = h(Icon, { name: 'ci:arrows-reload-01' })

const newContentDialogVisible = ref(false)
const newContentFormRef = ref<FormInstance>()
const newContentTitleInputRef = ref<typeof ElInput>()
const newContentType = ref<'page' | 'folder'>('page')
const newContentFormData = ref({ title: '', name: '' })
const newContentDialogExpanded = ref(false)
const toggleNewContentDialogExpanded = useToggle(newContentDialogExpanded)
const newContentNameChangedManually = ref(false)
const newContentFormRules = computed(() => ({
  name: [
    {
      required: true,
      message: newContentType.value === 'page' ? t('page-name-required') : t('folder-name-required'),
      trigger: 'blur',
    },
    {
      pattern: /^[a-z0-9-][a-z0-9_-]*$/,
      message: newContentType.value === 'page' ? t('invalid-page-name') : t('invalid-folder-name'),
      trigger: 'blur',
    },
  ],
} satisfies FormRules))

async function showNewContentDialog(type: 'page' | 'folder') {
  newContentType.value = type
  newContentFormData.value = { title: '', name: '' }
  newContentDialogExpanded.value = false
  newContentNameChangedManually.value = false
  newContentFormRef.value?.clearValidate()

  newContentDialogVisible.value = true
}

function focusNewContentDialog() {
  newContentTitleInputRef.value?.focus()
}

function onNewContentTitleChanged() {
  if (!newContentNameChangedManually.value) {
    newContentFormData.value.name = slugify(
      newContentFormData.value.title,
      { lower: true, strict: true },
    )
  }
}

function onNewContentNameChanged() {
  newContentNameChangedManually.value = true
}

async function submitNewContentDialog() {
  if (!newContentFormRef.value) {
    return
  }

  const formValid = await new Promise<boolean>(resolve => newContentFormRef.value?.validate(valid => resolve(valid)))
  if (!formValid) {
    newContentDialogExpanded.value = true
    return
  }

  newContentDialogVisible.value = false

  const newUrl = `${props.urlPath !== '' ? `/${props.urlPath}` : ''}/${newContentFormData.value.name}`

  if (newContentType.value === 'page') {
    await navigateTo({
      path: newUrl,
      query: { edit: 'true' },
      state: { title: newContentFormData.value.title },
    })
  } else {
    try {
      await apiFetch(`/pages${newUrl}`, {
        method: 'PUT',
        body: { folder: { meta: { title: newContentFormData.value.title } } },
      })
    } catch (err) {
      ElMessage({
        message: String(err),
        type: 'error',
      })
    }
    await navigateTo(newUrl)
  }
}

async function onEditTitle() {
  let title = props.folder.meta.title
  try {
    const msgBox = await ElMessageBox.prompt(t('folder-title'), {
      inputValue: title,
      confirmButtonText: t('ok'),
      cancelButtonText: t('cancel'),
    })
    title = msgBox.value
  } catch {
    return
  }

  try {
    const body = { folder: { meta: { title, tags: null }, content: [] } } satisfies PutRequest
    await apiFetch(`/pages/${props.urlPath}`, { method: 'PUT', body })

    props.onReload()
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
      },
    )
  } catch {
    // do nothing
    deleteConfirmOpen.value = false
    return
  }
  deleteConfirmOpen.value = false

  try {
    await apiFetch(`/pages/${props.urlPath}`, { method: 'DELETE' })

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

    <template v-if="urlPath !== ''" #title:suffix>
      <ElLink :underline="false" class="opacity-0 group-hover:opacity-100 duration-100" @click="onEditTitle">
        <Icon name="ci:edit" class="w-6 h-6" />
      </ElLink>
    </template>

    <template #actions>
      <div>
        <PlainButton v-if="allowWrite" icon="ci:file-add" :label="$t('create-page')" @click="showNewContentDialog('page')" />
        <PlainButton v-if="allowWrite" icon="ci:folder-add" :label="$t('create-folder')" @click="showNewContentDialog('folder')" />

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

    <div v-if="subfolders.length > 0">
      <h2 class="font-light text-xl">
        {{ $t('folders') }}
      </h2>
      <MultiColumnList
        :items="subfolders"
        :sort-and-group-by="item => item.title || item.name"
        :group-if-more-than="10"
      >
        <template #item="{ item }">
          <NuxtLink v-slot="{ navigate, href }" :to="`/${item.url}`" custom>
            <ElLink :href="href" @click="navigate">
              <Icon name="ci:folder" class="mr-1" /> {{ item.title || item.name }}
            </ElLink>
          </NuxtLink>
        </template>
      </MultiColumnList>
    </div>

    <div v-if="pages.length > 0">
      <h2 class="font-light text-xl">
        {{ $t('pages') }}
      </h2>
      <MultiColumnList
        :items="pages"
        :sort-and-group-by="item => item.title || item.name"
        :group-if-more-than="10"
      >
        <template #item="{ item }">
          <NuxtLink v-slot="{ navigate, href }" :to="`/${item.url}`" custom>
            <ElLink :href="href" @click="navigate">
              {{ item.title || item.name }}
            </ElLink>
          </NuxtLink>
        </template>
      </MultiColumnList>
    </div>

    <ClientOnly>
      <ElDialog
        v-model="newContentDialogVisible"
        :title="newContentType === 'page' ? $t('create-page') : $t('create-folder')"
        width="40%"
        @opened="focusNewContentDialog"
      >
        <ElForm
          ref="newContentFormRef"
          :model="newContentFormData"
          :rules="newContentFormRules"
          :validate-on-rule-change="false"
          label-position="top"
          @submit.prevent
          @keypress.enter="submitNewContentDialog"
        >
          <ElFormItem :label="newContentType === 'page' ? $t('page-title') : $t('folder-title')" prop="title">
            <ElInput ref="newContentTitleInputRef" v-model="newContentFormData.title" @input="onNewContentTitleChanged" />
          </ElFormItem>

          <span class="cursor-pointer text-xs" @click="toggleNewContentDialogExpanded()">
            <Icon name="ci:chevron-right" :class="newContentDialogExpanded && 'rotate-90 transform'" />
            {{ $t('more') }}
          </span>
          <div v-show="newContentDialogExpanded">
            <ElFormItem :label="newContentType === 'page' ? $t('page-name') : $t('folder-name')" prop="name">
              <ElInput v-model="newContentFormData.name" @input="onNewContentNameChanged" />
            </ElFormItem>
          </div>
        </ElForm>
        <template #footer>
          <PlainButton :label="$t('cancel')" @click="newContentDialogVisible = false" />
          <PlainButton type="primary" :label="$t('ok')" @click="submitNewContentDialog" />
        </template>
      </ElDialog>
    </ClientOnly>
  </Layout>
</template>
