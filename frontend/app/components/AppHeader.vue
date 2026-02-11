<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import { onKeyStroke, useDark, useToggle } from '@vueuse/core'
import { storeToRefs } from 'pinia'

import { useAppStore } from '~/store/app'
import { useAuthStore } from '~/store/auth'

const { t, locale, setLocale } = useI18n()

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
const searchOpen = ref(false)
const searchInputRef = ref<HTMLInputElement>()

async function onSearch() {
  await navigateTo({ path: '/_search', query: { q: searchQuery.value } })
  searchQuery.value = ''
  searchOpen.value = false
}

function openSearch() {
  searchOpen.value = true
}

function onSearchModalOpen() {
  // Focus the input when the modal is fully open
  nextTick(() => {
    searchInputRef.value?.focus()
  })
}

function onSearchModalClose() {
  searchQuery.value = ''
}

// Global keyboard shortcuts to open search (Ctrl+K, Cmd+K, /)
onKeyStroke('k', (e) => {
  if (e.ctrlKey || e.metaKey) {
    e.preventDefault()
    openSearch()
  }
})

onKeyStroke('/', (e) => {
  const target = e.target as HTMLElement
  if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA' && !target.isContentEditable) {
    e.preventDefault()
    openSearch()
  }
})

const isDark = useDark()
const toggleDark = useToggle(isDark)

const menuItems = computed(() => {
  const items: DropdownMenuItem[] = []

  const localeItems: DropdownMenuItem[] = locales.map(l => ({
    label: l.name,
    icon: l.icon,
    type: 'checkbox',
    checked: locale.value === l.code,
    onSelect: () => { setLocale(l.code) },
  }))

  if (auth.loggedIn) {
    items.push(
      {
        type: 'label',
        icon: 'tabler:user',
        label: auth.user?.displayName,
      },
      { type: 'separator' },
    )
  } else if (route.path !== '/_login') {
    items.push(
      {
        icon: 'tabler:login-2',
        label: t('sign-in'),
        to: `/_login?returnTo=${route.query.returnTo || route.fullPath}`,
      },
    )
  }

  if (auth.loggedIn) {
    items.push(
      {
        icon: 'tabler:user-circle',
        label: t('profile'),
        to: '/_profile',
      },
    )
  }

  if (allowAdmin.value) {
    items.push(
      {
        icon: 'tabler:users',
        label: t('users'),
        to: '/_admin/users',
      },
      {
        icon: 'tabler:settings',
        label: t('configuration'),
        to: '/_admin/settings',
      },
    )
  }

  items.push(
    {
      icon: 'tabler:language',
      label: t('language'),
      children: localeItems,
    },
    {
      icon: isDark.value ? 'tabler:sun' : 'tabler:moon',
      label: isDark.value ? t('dark-mode-off') : t('dark-mode-on'),
      onSelect: () => toggleDark(),
    },
  )

  if (auth.loggedIn) {
    items.push(
      { type: 'separator' },
      {
        icon: 'tabler:logout',
        label: t('sign-out'),
        onSelect: async () => {
          await auth.logout()
          toast.add({ description: t('signed-out'), color: 'success' })

          // Redirect to home if current page requires authentication
          const currentRoute = route.matched[route.matched.length - 1]
          const middleware = currentRoute?.meta?.middleware
          const requiresAuth = Array.isArray(middleware) && middleware.includes('require-auth')
          if (requiresAuth) {
            await navigateTo('/')
          }
        },
      },
    )
  }

  return items
})
</script>

<template>
  <div class="flex items-center py-1">
    <!-- Section 1: Title (fixed, no shrink) -->
    <ULink to="/" :active="false" class="text-xl font-light flex items-center whitespace-nowrap flex-shrink-0">
      <!-- favicon -->
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" class="w-[1em] h-[1em]"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 3H8.2c-1.12 0-1.68 0-2.108.218a1.999 1.999 0 0 0-.874.874C5 4.52 5 5.08 5 6.2v11.6c0 1.12 0 1.68.218 2.108a2 2 0 0 0 .874.874c.427.218.987.218 2.105.218h7.606c1.118 0 1.677 0 2.104-.218c.377-.192.683-.498.875-.874c.218-.428.218-.986.218-2.104V9m-6-6c.286.003.466.014.639.055c.204.05.399.13.578.24c.202.124.375.297.72.643l3.126 3.125c.346.346.518.518.642.72c.11.18.19.374.24.578c.04.173.051.354.054.639M13 3v2.8c0 1.12 0 1.68.218 2.108a2 2 0 0 0 .874.874c.427.218.987.218 2.105.218h2.802m0 0H19" /></svg>
      <span>{{ appTitle }}</span>
    </ULink>

    <!-- Section 2: Search button -->
    <div class="flex-1 flex justify-end items-center min-w-0 mr-1 ml-4">
      <UTooltip :kbds="['/']" :text="$t('search')">
        <UButton
          icon="tabler:search"
          variant="link"
          color="neutral"
          size="xs"
          :aria-label="$t('search')"
          @click="openSearch"
        />
      </UTooltip>
    </div>

    <!-- Search Modal -->
    <UModal
      v-model:open="searchOpen"
      :ui="{ content: 'sm:max-w-xl sm:top-24' }"
      @after:enter="onSearchModalOpen"
      @after:leave="onSearchModalClose"
    >
      <template #content>
        <div class="flex items-center gap-2 p-3">
          <UIcon name="tabler:search" class="w-5 h-5 text-(--ui-text-muted) flex-shrink-0" />
          <input
            ref="searchInputRef"
            v-model="searchQuery"
            type="text"
            :placeholder="$t('search')"
            class="flex-1 bg-transparent border-none outline-none text-base text-(--ui-text) placeholder:text-(--ui-text-muted)"
            @keypress.enter="onSearch"
          >
          <UKbd class="hidden sm:flex">
            Esc
          </UKbd>
        </div>
      </template>
    </UModal>

    <!-- Section 3: Menu (fixed, no shrink) -->
    <UDropdownMenu
      class="flex-shrink-0"
      :items="menuItems"
      size="lg"
    >
      <ReactiveButton icon="tabler:menu-2" variant="link" :aria-label="t('menu')" label="" />
    </UDropdownMenu>
  </div>
</template>
