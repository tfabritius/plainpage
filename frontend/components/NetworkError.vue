<script setup lang="ts">
const _props = defineProps<{
  msg?: string
  onReload: Function
}>()

const { t } = useI18n()

useHead(() => ({ title: t('error') }))
</script>

<template>
  <div class="flex">
    <div class="m-auto text-center text-gray-500">
      <Icon name="ci:cloud-off" size="5em" />

      <div class="m-2">
        {{ msg || $t('there-was-an-error') }}
      </div>

      <ElButton @click="onReload()">
        <Icon name="ci:arrows-reload-01" class="mr-1" />
        {{ $t('try-again') }}
      </ElButton>
    </div>
  </div>
</template>
