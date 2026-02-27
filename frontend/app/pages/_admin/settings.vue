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
      { op: 'replace', path: '/retention/trash/maxAgeDays', value: data.value?.retention.trash.maxAgeDays },
      { op: 'replace', path: '/retention/attic/maxAgeDays', value: data.value?.retention.attic.maxAgeDays },
      { op: 'replace', path: '/retention/attic/maxVersions', value: data.value?.retention.attic.maxVersions },
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
      <ReactiveButton color="success" icon="tabler:device-floppy" :label="$t('save')" type="submit" form="settingsForm" />
    </template>

    <UForm
      id="settingsForm"
      :state="data"
      class="space-y-6"
      @submit="onSave"
    >
      <UFormField :label="$t('application-title')">
        <UInput v-model="data.appTitle" class="w-full" />
      </UFormField>

      <PlainFieldset :legend="$t('permissions')">
        <AclTable ref="aclTableRef" :acl="data?.acl ?? []" :show-columns="['register', 'admin']" />
      </PlainFieldset>

      <PlainFieldset :legend="$t('retention')">
        <p class="text-sm text-[var(--ui-text-muted)] mb-4">
          {{ $t('retention-description') }}
        </p>

        <div class="space-y-4">
          <UFormField :label="$t('trash-retention-age')">
            <div class="flex items-center gap-2">
              <UInput
                v-model.number="data.retention.trash.maxAgeDays"
                type="number"
                min="0"
                class="w-24"
              />
              <span class="text-sm text-[var(--ui-text-muted)]">{{ $t('days') }}</span>
              <span class="text-xs text-[var(--ui-text-dimmed)]">{{ $t('zero-disabled') }}</span>
            </div>
          </UFormField>

          <UFormField :label="$t('attic-retention-age')">
            <div class="flex items-center gap-2">
              <UInput
                v-model.number="data.retention.attic.maxAgeDays"
                type="number"
                min="0"
                class="w-24"
              />
              <span class="text-sm text-[var(--ui-text-muted)]">{{ $t('days') }}</span>
              <span class="text-xs text-[var(--ui-text-dimmed)]">{{ $t('zero-disabled') }}</span>
            </div>
          </UFormField>

          <UFormField :label="$t('attic-retention-versions')">
            <div class="flex items-center gap-2">
              <UInput
                v-model.number="data.retention.attic.maxVersions"
                type="number"
                min="0"
                class="w-24"
              />
              <span class="text-sm text-[var(--ui-text-muted)]">{{ $t('versions-per-page') }}</span>
              <span class="text-xs text-[var(--ui-text-dimmed)]">{{ $t('zero-unlimited') }}</span>
            </div>
          </UFormField>
        </div>
      </PlainFieldset>

      <UFormField :label="$t('version')" class="mt-4">
        <UInput
          :model-value="version" disabled class="w-full"
        />
      </UFormField>
    </UForm>
  </Layout>
</template>
