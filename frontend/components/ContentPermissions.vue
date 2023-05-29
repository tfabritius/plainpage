<script setup lang="ts">
import type { Breadcrumb, ContentMeta } from '~/types/'
import { AclTable } from '#components'

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

const aclTableRef = ref<InstanceType<typeof AclTable>>()

async function onGoBack() {
  await navigateTo({ query: { } })
}

async function onSave() {
  const apiData = (customPermissions.value || props.urlPath === '') ? aclTableRef.value?.getAcl() : null

  await apiFetch(`/pages${props.urlPath}`, {
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
      <Icon v-if="isFolder" name="ci:folder" class="mr-1" />
      <span v-if="title">{{ title }}</span>
      <span v-else class="italic">{{ $t('untitled') }}</span>
    </template>

    <template #actions>
      <PlainButton icon="ci:skip-back" :label="$t('back-to-content')" @click="onGoBack" />
      <PlainButton icon="ci:save" :label="$t('save')" type="success" @click="onSave" />
    </template>

    <ElSwitch
      v-if="urlPath !== ''"
      v-model="customPermissions"
      :active-text="$t('define-custom-permissions')"
      :inactive-text="$t('inherit-permissions')"
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
