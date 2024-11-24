<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAppStore } from '~/store/app'
import { useAuthStore } from '~/store/auth'

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
  <UApp>
    <NuxtPage />
    <NuxtLoadingIndicator />
  </UApp>
</template>

<style>
@import 'tailwindcss';
@import '@nuxt/ui';

@theme {
  /* https://www.tints.dev/plainpage/00BFE6 */

  --color-plainpage-50: #e5fbff;
  --color-plainpage-100: #c7f6ff;
  --color-plainpage-200: #8fecff;
  --color-plainpage-300: #57e3ff;
  --color-plainpage-400: #1fdaff;
  --color-plainpage-500: #00bfe6;
  --color-plainpage-600: #0099b8;
  --color-plainpage-700: #00738a;
  --color-plainpage-800: #004d5c;
  --color-plainpage-900: #00262e;
  --color-plainpage-950: #001519;
}

:root {
  --ui-radius: var(--radius-sm);
}
</style>
