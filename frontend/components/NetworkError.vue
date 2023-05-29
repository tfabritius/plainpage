<script setup lang="ts">
const _props = defineProps<{
  msg?: string
  onReload: Function
}>()

const { t } = useI18n()

useHead(() => ({ title: t('error') }))
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-gray-500">
      <Icon name="ci:cloud-off" size="5em" />

      <div class="m-2">
        {{ msg || $t('there-was-an-error') }}
      </div>

      <PlainButton icon="ci:arrows-reload-01" :label="$t('try-again')" @click="onReload()" />
    </div>
  </div>
</template>
