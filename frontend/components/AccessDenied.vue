<script setup lang="ts">
const { t } = useI18n()

useHead(() => ({ title: t('access-denied') }))
</script>

<template>
  <div class="flex">
    <div class="m-auto text-center text-gray-500">
      <Icon name="ci:stop-sign" size="5em" />
      <div class="m-2">
        {{ $t('access-denied') }}
      </div>
    </div>
  </div>
</template>
