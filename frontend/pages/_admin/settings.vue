<script setup lang="ts">
import type { Config } from '~/types'
import { AclTable } from '#components'

const { data, error, refresh } = await useAsyncData('/config', () => apiFetch<Config>('/config'))

const aclTableRef = ref<InstanceType<typeof AclTable>>()

async function onSave() {
  const acl = aclTableRef.value?.getAcl()

  const response = await apiFetch<Config>('/config', {
    method: 'PATCH',
    body: [
      { op: 'replace', path: '/appName', value: data.value?.appName },
      { op: 'replace', path: '/acl', value: acl },
    ],
  })

  ElMessage({
    message: 'Saved',
    type: 'success',
  })

  data.value = response
}
</script>

<template>
  <NetworkError v-if="!data" :msg="error?.message" @refresh="refresh" />
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
        <ElInput v-model="data.appName" />
      </ElFormItem>
      <ElFormItem label="Permissions">
        <AclTable ref="aclTableRef" :acl="data?.acl ?? []" :show-columns="['register', 'admin']" />
      </ElFormItem>
    </ElForm>
  </Layout>
</template>
