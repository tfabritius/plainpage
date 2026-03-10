<script setup lang="ts">
const { t } = useI18n()

useHead(() => ({ title: t('access-denied') }))
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-muted">
      <UIcon name="tabler:lock" size="5em" />
      <div class="m-2">
        {{ $t('access-denied') }}
      </div>
    </div>
  </div>
</template>
