<script setup lang="ts">
import type { Config } from '~/types'
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()
const toast = useToast()

useHead({ title: t('configuration') })

const app = useAppStore()
const { version } = storeToRefs(app)

const { data, error, refresh } = await useAsyncData('/config', () => apiFetch<Config>('/config'))

const aclTable = useTemplateRef('aclTableRef')

async function onSave() {
  const acl = aclTable.value?.getAcl()

  const response = await apiFetch<Config>('/config', {
    method: 'PATCH',
    body: [
      { op: 'replace', path: '/appTitle', value: data.value?.appTitle },
      { op: 'replace', path: '/acl', value: acl },
    ],
  })

  toast.add({
    description: t('saved'),
    color: 'success',
  })

  data.value = response
  app.refresh()
}
</script>

<template>
  <SubpageNetworkError
    v-if="!data"
    :msg="error?.message"
    :on-reload="refresh"
  />
  <Layout v-else>
    <template #title>
      {{ $t('configuration') }}
    </template>

    <template #actions>
      <ReactiveButton color="success" icon="ci:save" :label="$t('save')" type="submit" form="settingsForm" />
    </template>

    <UForm
      id="settingsForm"
      :state="data"
      @submit="onSave"
    >
      <UFormField :label="$t('application-title')">
        <UInput v-model="data.appTitle" class="w-full" />
      </UFormField>
      <UFormField :label="$t('permissions')" class="mt-4">
        <AclTable ref="aclTableRef" :acl="data?.acl ?? []" :show-columns="['register', 'admin']" />
      </UFormField>
      <UFormField :label="$t('version')" class="mt-4">
        <UInput
          :model-value="version" disabled class="w-full"
        />
      </UFormField>
    </UForm>
  </Layout>
</template>
