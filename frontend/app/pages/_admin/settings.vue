<script setup lang="ts">
import type { Config, GetStatsResponse, RestoreBackupResponse } from '~/types'
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'
import { useAuthStore } from '~/store/auth'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()
const toast = useToast()

useHead({ title: t('configuration') })

const app = useAppStore()
const authStore = useAuthStore()
const { version } = storeToRefs(app)

const { data, error, refresh } = await useAsyncData('/config', () => apiFetch<Config>('/config'))
const { data: stats, status: statsStatus, refresh: refreshStats } = useLazyAsyncData('/stats', () => apiFetch<GetStatsResponse>('/stats'))

// Backup download options
const backupIncludeConfig = ref(true)
const backupIncludeUsers = ref(true)
const backupIsDownloading = ref(false)

// Restore options
const restoreFile = ref<File | null>(null)
const restoreIsUploading = ref(false)
const confirmDialog = useTemplateRef('confirmDialog')

async function downloadBackup() {
  backupIsDownloading.value = true
  try {
    const params = new URLSearchParams()
    if (backupIncludeConfig.value) {
      params.set('includeConfig', '')
    }
    if (backupIncludeUsers.value) {
      params.set('includeUsers', '')
    }

    const response = await apiFetch<Blob>(`/storage/download?${params.toString()}`, {
      responseType: 'blob',
    })

    // Create download link
    const url = window.URL.createObjectURL(response)
    const a = document.createElement('a')
    a.href = url
    // eslint-disable-next-line e18e/prefer-static-regex
    a.download = `plainpage-backup-${new Date().toISOString().slice(0, 19).replace(/[T:]/g, '-')}.zip`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)

    toast.add({
      description: t('_settings.backup-downloaded'),
      color: 'success',
    })
  } catch {
    toast.add({
      description: t('error'),
      color: 'error',
    })
  } finally {
    backupIsDownloading.value = false
  }
}

function onRestoreFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  restoreFile.value = input.files?.[0] || null
}

async function confirmRestore() {
  if (!restoreFile.value) {
    return
  }

  const confirmed = await confirmDialog.value?.confirm(
    t('_settings.restore-confirm'),
    {
      title: t('_settings.restore'),
      confirmButtonText: t('_settings.restore-upload'),
      confirmButtonColor: 'error',
    },
  )

  if (confirmed) {
    await performRestore()
  }
}

async function performRestore() {
  if (!restoreFile.value) {
    return
  }

  restoreIsUploading.value = true

  try {
    const response = await apiFetch<RestoreBackupResponse>('/storage/restore', {
      method: 'POST',
      body: restoreFile.value,
      headers: {
        'Content-Type': 'application/zip',
      },
    })

    if (response.usersRestored) {
      toast.add({
        description: t('_settings.restore-success-logout'),
        color: 'success',
      })
      // Log out and redirect to login
      authStore.accessToken = ''
      authStore.user = undefined
      await navigateTo('/_login')
    } else {
      toast.add({
        description: t('_settings.restore-success'),
        color: 'success',
      })
      // Refresh the page data
      await refresh()
      await refreshStats()
      app.refresh()
    }

    restoreFile.value = null
  } catch {
    toast.add({
      description: t('error'),
      color: 'error',
    })
  } finally {
    restoreIsUploading.value = false
  }
}

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
        <p class="text-sm text-muted mb-4">
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
              <span class="text-sm text-muted">{{ $t('days') }}</span>
              <span class="text-xs text-dimmed">{{ $t('zero-disabled') }}</span>
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
              <span class="text-sm text-muted">{{ $t('days') }}</span>
              <span class="text-xs text-dimmed">{{ $t('zero-disabled') }}</span>
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
              <span class="text-sm text-muted">{{ $t('versions-per-page') }}</span>
              <span class="text-xs text-dimmed">{{ $t('zero-unlimited') }}</span>
            </div>
          </UFormField>
        </div>
      </PlainFieldset>

      <PlainFieldset :legend="$t('_settings.statistics')">
        <div class="flex items-center justify-between mb-4">
          <p class="text-sm text-muted">
            {{ $t('_settings.statistics-description') }}
          </p>
          <UButton variant="ghost" icon="tabler:refresh" @click="() => refreshStats()" />
        </div>

        <div v-if="statsStatus === 'pending'" class="flex items-center justify-center py-8">
          <UIcon name="tabler:loader-2" class="w-6 h-6 animate-spin text-muted" />
        </div>

        <div v-else-if="stats" class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h4 class="text-sm font-medium mb-2">
              {{ $t('_settings.memory-usage') }}
            </h4>
            <dl class="space-y-1 text-sm">
              <div class="flex justify-between">
                <dt class="text-muted">
                  {{ $t('_settings.memory-alloc') }}
                </dt>
                <dd>{{ formatBytes(stats.memory.alloc) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted">
                  {{ $t('_settings.memory-total-alloc') }}
                </dt>
                <dd>{{ formatBytes(stats.memory.totalAlloc) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted">
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
                <dt class="text-muted">
                  {{ $t('_settings.disk-pages') }}
                </dt>
                <dd>{{ formatBytes(stats.diskUsage.pages) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted">
                  {{ $t('_settings.disk-attic') }}
                </dt>
                <dd>{{ formatBytes(stats.diskUsage.attic) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted">
                  {{ $t('_settings.disk-trash') }}
                </dt>
                <dd>{{ formatBytes(stats.diskUsage.trash) }}</dd>
              </div>
              <div class="flex justify-between font-medium border-t border-default pt-1 mt-1">
                <dt>{{ $t('_settings.disk-total') }}</dt>
                <dd>{{ formatBytes(stats.diskUsage.total) }}</dd>
              </div>
            </dl>
          </div>
        </div>
      </PlainFieldset>

      <PlainFieldset :legend="$t('_settings.backup')">
        <p class="text-sm text-muted mb-4">
          {{ $t('_settings.backup-description') }}
        </p>

        <div class="space-y-4">
          <div class="flex flex-col gap-3">
            <UCheckbox v-model="backupIncludeConfig" :label="$t('_settings.backup-include-config')" />
            <UCheckbox v-model="backupIncludeUsers" :label="$t('_settings.backup-include-users')" />
            <div v-if="backupIncludeUsers" class="flex items-center gap-1 ml-6 text-warning">
              <UIcon name="tabler:alert-triangle" class="w-4 h-4 shrink-0" />
              <p class="text-xs">
                {{ $t('_settings.backup-users-warning') }}
              </p>
            </div>
          </div>

          <UButton
            icon="tabler:download"
            :loading="backupIsDownloading"
            :disabled="backupIsDownloading"
            @click="downloadBackup"
          >
            {{ $t('_settings.backup-download') }}
          </UButton>
        </div>
      </PlainFieldset>

      <PlainFieldset :legend="$t('_settings.restore')">
        <p class="text-sm text-muted mb-4">
          {{ $t('_settings.restore-description') }}
        </p>

        <div class="flex items-center gap-1 mb-4 text-warning">
          <UIcon name="tabler:alert-triangle" class="w-4 h-4 shrink-0" />
          <p class="text-xs">
            {{ $t('_settings.restore-warning') }}
          </p>
        </div>

        <div class="space-y-4">
          <div class="flex items-center gap-4">
            <input
              type="file"
              accept=".zip"
              class="text-sm file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-elevated file:text-default hover:file:bg-accented file:cursor-pointer"
              @change="onRestoreFileChange"
            >
          </div>

          <UButton
            icon="tabler:upload"
            color="error"
            :loading="restoreIsUploading"
            :disabled="restoreIsUploading || !restoreFile"
            @click="confirmRestore"
          >
            {{ $t('_settings.restore-upload') }}
          </UButton>
        </div>

        <PlainDialog ref="confirmDialog" />
      </PlainFieldset>

      <UFormField :label="$t('version')" class="mt-4">
        <UInput
          :model-value="version" disabled class="w-full"
        />
      </UFormField>
    </UForm>
  </Layout>
</template>
