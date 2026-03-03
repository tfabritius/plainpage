<script setup lang="ts">
import type { Config, GetStatsResponse } from '~/types'
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
const { data: stats, status: statsStatus, refresh: refreshStats } = useLazyAsyncData('/stats', () => apiFetch<GetStatsResponse>('/stats'))

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

function formatBytes(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KiB`
  }
  if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`
  }
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GiB`
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

      <PlainFieldset :legend="$t('_settings.statistics')">
        <div class="flex items-center justify-between mb-4">
          <p class="text-sm text-[var(--ui-text-muted)]">
            {{ $t('_settings.statistics-description') }}
          </p>
          <UButton variant="ghost" icon="tabler:refresh" @click="() => refreshStats()" />
        </div>

        <div v-if="statsStatus === 'pending'" class="flex items-center justify-center py-8">
          <UIcon name="tabler:loader-2" class="w-6 h-6 animate-spin text-[var(--ui-text-muted)]" />
        </div>

        <div v-else-if="stats" class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h4 class="text-sm font-medium mb-2">
              {{ $t('_settings.memory-usage') }}
            </h4>
            <dl class="space-y-1 text-sm">
              <div class="flex justify-between">
                <dt class="text-[var(--ui-text-muted)]">
                  {{ $t('_settings.memory-alloc') }}
                </dt>
                <dd>{{ formatBytes(stats.memory.alloc) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-[var(--ui-text-muted)]">
                  {{ $t('_settings.memory-total-alloc') }}
                </dt>
                <dd>{{ formatBytes(stats.memory.totalAlloc) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-[var(--ui-text-muted)]">
                  {{ $t('_settings.memory-sys') }}
                </dt>
                <dd>{{ formatBytes(stats.memory.sys) }}</dd>
              </div>
            </dl>
          </div>

          <div>
            <h4 class="text-sm font-medium mb-2">
              {{ $t('_settings.disk-usage') }}
            </h4>
            <dl class="space-y-1 text-sm">
              <div class="flex justify-between">
                <dt class="text-[var(--ui-text-muted)]">
                  {{ $t('_settings.disk-pages') }}
                </dt>
                <dd>{{ formatBytes(stats.diskUsage.pages) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-[var(--ui-text-muted)]">
                  {{ $t('_settings.disk-attic') }}
                </dt>
                <dd>{{ formatBytes(stats.diskUsage.attic) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-[var(--ui-text-muted)]">
                  {{ $t('_settings.disk-trash') }}
                </dt>
                <dd>{{ formatBytes(stats.diskUsage.trash) }}</dd>
              </div>
              <div class="flex justify-between font-medium border-t border-[var(--ui-border)] pt-1 mt-1">
                <dt>{{ $t('_settings.disk-total') }}</dt>
                <dd>{{ formatBytes(stats.diskUsage.total) }}</dd>
              </div>
            </dl>
          </div>
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
