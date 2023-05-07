<script setup lang="ts">
import { storeToRefs } from 'pinia'

import type { Config } from '~/types'
import { AclTable } from '#components'
import { useAppStore } from '~/store/app'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()

useHead({ title: t('configuration') })

const app = useAppStore()
const { gitSha } = storeToRefs(app)

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
    message: t('saved'),
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
      {{ $t('configuration') }}
    </template>

    <template #actions>
      <ElButton class="m-1" type="success" @click="onSave">
        <Icon name="ci:save" /> <span class="hidden md:inline ml-1">{{ $t('save') }}</span>
      </ElButton>
    </template>

    <ElForm
      label-position="top"
      @submit.prevent
    >
      <ElFormItem :label="$t('application-title')">
        <ElInput v-model="data.appTitle" />
      </ElFormItem>
      <ElFormItem :label="$t('permissions')">
        <AclTable ref="aclTableRef" :acl="data?.acl ?? []" :show-columns="['register', 'admin']" />
      </ElFormItem>
      <ElFormItem :label="$t('version')">
        <ElInput
          :value="gitSha" disabled
        />
      </ElFormItem>
    </ElForm>
  </Layout>
</template>
