<script setup lang="ts">
import type { Config } from '~/types'
import { AclTable } from '#components'
import { useAppStore } from '~/store/app'

definePageMeta({
  middleware: ['require-auth'],
})

const app = useAppStore()

const { data, error, refresh } = await useAsyncData('/config', () => apiFetch<Config>('/config'))

const aclTableRef = ref<InstanceType<typeof AclTable>>()

async function onSave() {
  const acl = aclTableRef.value?.getAcl()

  const response = await apiFetch<Config>('/config', {
    method: 'PATCH',
    body: [
      { op: 'replace', path: '/appTitle', value: data.value?.appTitle },
      { op: 'replace', path: '/acl', value: acl },
    ],
  })

  ElMessage({
    message: 'Saved',
    type: 'success',
  })

  data.value = response
  app.refresh()
}
</script>

<template>
  <NetworkError
    v-if="!data"
    :msg="error?.message"
    :on-reload="refresh"
  />
  <Layout v-else>
    <template #title>
      Configuration
    </template>

    <template #actions>
      <ElButton class="m-1" type="success" @click="onSave">
        <Icon name="ci:save" /> <span class="hidden md:inline ml-1">Save</span>
      </ElButton>
    </template>

    <ElForm
      label-position="top"
      @submit.prevent
    >
      <ElFormItem label="Application title">
        <ElInput v-model="data.appTitle" />
      </ElFormItem>
      <ElFormItem label="Permissions">
        <AclTable ref="aclTableRef" :acl="data?.acl ?? []" :show-columns="['register', 'admin']" />
      </ElFormItem>
    </ElForm>
  </Layout>
</template>
