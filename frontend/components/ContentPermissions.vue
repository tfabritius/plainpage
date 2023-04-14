<script setup lang="ts">
import type { Breadcrumb, PageMeta } from '~/types/'
import { AclTable } from '#components'

const props = defineProps<{
  urlPath: string
  meta: PageMeta
  title: string | undefined
  breadcrumbs: Breadcrumb[]
  isFolder: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

const urlPath = computed(() => props.urlPath)
const meta = computed(() => props.meta)
const title = computed(() => props.title)
const isFolder = computed(() => props.isFolder)

const customPermissions = ref(!!meta.value.acl)

const aclTableRef = ref<InstanceType<typeof AclTable>>()

const onGoBack = async () => {
  await navigateTo({ query: { } })
}

const onSave = async () => {
  const apiData = (customPermissions.value || urlPath.value === '') ? aclTableRef.value?.getAcl() : null

  await apiFetch(`/pages${urlPath.value}`, { method: 'PATCH', body: [{ op: 'replace', path: isFolder.value ? '/folder/meta/acl' : '/page/meta/acl', value: apiData }] })

  emit('refresh')
  onGoBack()
}
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <Icon v-if="isFolder" name="ci:folder" class="mr-1" />
      <span v-if="title">{{ title }}</span>
      <span v-else class="italic">Untitled</span>
    </template>

    <template #actions>
      <ElButton class="m-1" @click="onGoBack">
        <Icon name="ci:skip-back" /> <span class="hidden md:inline ml-1">Back to content</span>
      </ElButton>
      <span />
      <ElButton class="m-1" type="success" @click="onSave">
        <Icon name="ci:save" /> <span class="hidden md:inline ml-1">Save</span>
      </ElButton>
    </template>

    <ElSwitch
      v-if="urlPath !== ''"
      v-model="customPermissions"
      active-text="Define custom permissions"
      inactive-text="Inherit permissions from parent folder"
    />
    <div v-if="customPermissions">
      <AclTable ref="aclTableRef" :acl="meta.acl ?? []" :show-admin-rule="true" :show-columns="['read', 'write', 'delete']" />
    </div>
  </Layout>
</template>
