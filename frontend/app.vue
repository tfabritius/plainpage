<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAuthStore } from '~/store/auth'
import { useAppStore } from '~/store/app'

import 'element-plus/theme-chalk/src/dark/css-vars.scss'

useHead({
  bodyAttrs: {
    class: 'font-sans m-0',
  },
  link: [{
    rel: 'icon',
    href: '/favicon.svg',
  }],
})

const auth = useAuthStore()

const app = useAppStore()
const { refresh } = app
const { appTitle } = storeToRefs(app)

refresh()
watch(() => auth.loggedIn, () => refresh())

useHead(() => ({ titleTemplate: `%s | ${appTitle.value}` }))
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />
    <NuxtPage class="flex-grow" />
    <NuxtLoadingIndicator />
  </div>
</template>
