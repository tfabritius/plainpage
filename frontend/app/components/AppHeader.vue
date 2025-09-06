<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import { useDark, useToggle } from '@vueuse/core'
import { storeToRefs } from 'pinia'

import { useAppStore } from '~/store/app'
import { useAuthStore } from '~/store/auth'

const { t, locale, setLocaleCookie } = useI18n()

type Locale = typeof locale.value

const locales: { code: Locale, name: string, icon: string }[] = [
  {
    code: 'en',
    name: 'English',
    icon: 'flag:us-1x1',
  },
  {
    code: 'de',
    name: 'Deutsch',
    icon: 'flag:de-1x1',
  },
  {
    code: 'es',
    name: 'Español',
    icon: 'flag:es-1x1',
  },
]

const auth = useAuthStore()

const app = useAppStore()
const { appTitle, allowAdmin } = storeToRefs(app)

const route = useRoute()

const toast = useToast()

const searchQuery = ref('')
async function onSearch() {
  await navigateTo({ path: '/_search', query: { q: searchQuery.value } })
  searchQuery.value = ''
}

async function onKeyPressInSearch(e: Event) {
  // Prevent keyboard shortcuts being fired if focus is in input field
  e.stopPropagation()
  e.stopImmediatePropagation()
}

const isDark = useDark()
const toggleDark = useToggle(isDark)

const menuItems = computed(() => {
  const items: DropdownMenuItem[] = []

  const localeItems: DropdownMenuItem[] = locales.map(l => ({
    label: l.name,
    icon: l.icon,
    type: 'checkbox',
    checked: locale.value === l.code,
    onSelect: () => {
      locale.value = l.code
      setLocaleCookie(l.code)
    },
  }))

  if (auth.loggedIn) {
    items.push(
      {
        type: 'label',
        icon: 'ci:user',
        label: auth.user?.displayName,
      },
      { type: 'separator' },
    )
  } else if (route.path !== '/_login') {
    items.push(
      {
        icon: 'ic:round-log-in',
        label: t('sign-in'),
        to: `/_login?returnTo=${route.query.returnTo || route.fullPath}`,
      },
    )
  }

  if (auth.loggedIn) {
    items.push(
      {
        icon: 'ci:user-circle',
        label: t('profile'),
        to: '/_profile',
      },
    )
  }

  if (allowAdmin.value) {
    items.push(
      {
        icon: 'ci:users',
        label: t('users'),
        to: '/_admin/users',
      },
      {
        icon: 'ci:settings',
        label: t('configuration'),
        to: '/_admin/settings',
      },
    )
  }

  items.push(
    {
      icon: 'lucide:languages',
      label: t('language'),
      children: localeItems,
    },
    {
      icon: isDark.value ? 'ci:sun' : 'ci:moon',
      label: isDark.value ? t('dark-mode-off') : t('dark-mode-on'),
      onSelect: () => toggleDark(),
    },
  )

  if (auth.loggedIn) {
    items.push(
      { type: 'separator' },
      {
        icon: 'ic:round-log-out',
        label: t('sign-out'),
        onSelect: () => {
          toast.add({ description: t('signed-out'), color: 'success' })
          auth.logout()
        },
      },
    )
  }

  return items
})
</script>

<template>
  <div class="flex justify-between">
    <ULink to="/" :active="false" class="text-xl font-light flex items-center whitespace-nowrap">
      <UIcon name="ci:file-blank" />
      <span>{{ appTitle }}</span>
    </ULink>

    <span class="flex">
      <UInput
        v-model="searchQuery"
        :placeholder="$t('search')"
        class="max-w-40 mx-1"
        size="xs"
        trailing-icon="ci:search"
        @keypress.enter="onSearch"
        @keydown="onKeyPressInSearch"
      />

      <UDropdownMenu
        class="m-1"
        :items="menuItems"
        size="lg"
      >
        <ReactiveButton icon="ci:hamburger" variant="link" :label="t('menu')" />
      </UDropdownMenu>
    </span>
  </div>
</template>
