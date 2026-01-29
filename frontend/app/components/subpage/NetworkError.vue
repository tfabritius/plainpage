<script setup lang="ts">
defineProps<{
  msg?: string
  onReload: () => void
}>()

const { t } = useI18n()

useHead(() => ({ title: t('error') }))
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-gray-500">
      <UIcon name="tabler:network-off" size="5em" />

      <div class="m-2">
        {{ msg || $t('there-was-an-error') }}
      </div>

      <ReactiveButton icon="tabler:refresh" :label="$t('try-again')" @click="onReload()" />
    </div>
  </div>
</template>
