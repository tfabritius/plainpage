<script setup lang="ts">
import type { Breadcrumb, ContentMeta } from '~/types/'

const props = defineProps<{
  urlPath: string
  meta: ContentMeta
  title: string | undefined
  breadcrumbs: Breadcrumb[]
  isFolder: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

useHead(() => ({ title: `Permissions: ${props.title}` }))

const customPermissions = ref(!!props.meta.acl)

const aclTable = useTemplateRef('aclTableRef')

async function onGoBack() {
  await navigateTo({ query: { } })
}

async function onSave() {
  const apiData = (customPermissions.value || props.urlPath === '') ? aclTable.value?.getAcl() : null

  await apiFetch(`/pages/${props.urlPath}`, {
    method: 'PATCH',
    body: [{
      op: 'replace',
      path: props.isFolder ? '/folder/meta/acl' : '/page/meta/acl',
      value: apiData,
    }],
  })

  emit('refresh')
  onGoBack()
}
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <UIcon v-if="isFolder" name="ci:folder" class="mr-1" />
      <span v-if="title">{{ title }}</span>
      <span v-else class="italic">{{ $t('untitled') }}</span>
    </template>

    <template #actions>
      <UButton icon="ci:skip-back" :label="$t('back-to-content')" @click="onGoBack" />
      <UButton icon="ci:save" :label="$t('save')" color="success" variant="solid" class="ml-3" @click="onSave" />
    </template>

    <USwitch
      v-if="urlPath !== ''"
      v-model="customPermissions"
      :label="customPermissions ? $t('define-custom-permissions') : $t('inherit-permissions')"
    />
    <div v-if="customPermissions">
      <AclTable
        ref="aclTableRef"
        :acl="meta.acl ?? []"
        :show-admin-rule="true"
        :show-columns="['read', 'write', 'delete']"
      />
    </div>
  </Layout>
</template>
