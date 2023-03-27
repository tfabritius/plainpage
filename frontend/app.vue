<script setup lang="ts">
useHead({
  bodyAttrs: {
    class: 'font-sans m-0',
  },
  link: [{
    rel: 'icon',
    href: '/favicon.svg',
  }],
})

const appName = 'PlainPage'

useHead(() => ({ titleTemplate: `%s | ${appName}` }))
</script>

<template>
  <div class="p-2">
    <div class="flex justify-between">
      <NuxtLink v-slot="{ navigate, href }" custom to="/">
        <ElLink :underline="false" :href="href" @click="navigate">
          <span class="text-xl font-light flex items-center">
            <Icon name="ci:file-blank" />
            <span>{{ appName }}</span>
          </span>
        </ElLink>
      </NuxtLink>

      <ElLink :underline="false" @click="ElMessage('Not implemented yet')">
        <Icon name="ic:baseline-login" class="mr-1" /> <span class="font-normal">Sign in</span>
      </ElLink>
    </div>
    <NuxtPage />
    <NuxtLoadingIndicator />
  </div>
</template>
