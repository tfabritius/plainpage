<script setup lang="ts">
import { Icon } from '#components'
import { useDark, useToggle } from '@vueuse/core'

import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'
import { useAuthStore } from '~/store/auth'

const auth = useAuthStore()

const app = useAppStore()
const { appTitle, allowAdmin } = storeToRefs(app)

const route = useRoute()

const ProfileIcon = h(Icon, { name: 'ci:user-circle' })
const UsersIcon = h(Icon, { name: 'ci:users' })
const SettingsIcon = h(Icon, { name: 'ci:settings' })
const LogoutIcon = h(Icon, { name: 'ic:round-log-out' })

async function handleDropdownMenuCommand(command: string | number | object) {
  if (command === 'profile') {
    await navigateTo('/_profile')
  } else if (command === 'users') {
    await navigateTo('/_admin/users')
  } else if (command === 'settings') {
    await navigateTo('/_admin/settings')
  } else if (command === 'logout') {
    auth.logout()
  } else {
    throw new Error(`Unhandled command ${command}`)
  }
}

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
</script>

<template>
  <div class="flex justify-between">
    <NuxtLink v-slot="{ navigate, href }" custom to="/">
      <ElLink :underline="false" :href="href" @click="navigate">
        <span class="text-xl font-light flex items-center  whitespace-nowrap">
          <Icon name="ci:file-blank" />
          <span>{{ appTitle }}</span>
        </span>
      </ElLink>
    </NuxtLink>

    <span>
      <ElInput
        v-model="searchQuery"
        :placeholder="$t('search')"
        class="max-w-40 mx-1"
        size="small"
        @keypress.enter="onSearch"
        @keydown="onKeyPressInSearch"
      >
        <template #suffix>
          <Icon name="ci:search" />
        </template>
      </ElInput>

      <ElLink :underline="false" class="m-1" @click="toggleDark()">
        <Icon :name="isDark ? 'ci:moon' : 'ci:sun'" />
      </ElLink>

      <LocaleSwitcher class="m-1" />

      <ElDropdown
        v-if="auth.loggedIn"
        trigger="click"
        class="m-1"
        @command="handleDropdownMenuCommand"
      >
        <ElLink :underline="false" href="#">
          <Icon name="ci:user" class="mr-1" />
          <span class="font-normal">{{ auth.user?.displayName }}</span>
        </ElLink>
        <template #dropdown>
          <ElDropdownMenu>
            <ElDropdownItem :icon="ProfileIcon" command="profile">
              {{ $t('profile') }}
            </ElDropdownItem>
            <ElDropdownItem v-if="allowAdmin" :icon="UsersIcon" command="users">
              {{ $t('users') }}
            </ElDropdownItem>
            <ElDropdownItem v-if="allowAdmin" :icon="SettingsIcon" command="settings">
              {{ $t('configuration') }}
            </ElDropdownItem>
            <ElDropdownItem :icon="LogoutIcon" command="logout">
              {{ $t('sign-out') }}
            </ElDropdownItem>
          </ElDropdownMenu>
        </template>
      </ElDropdown>

      <NuxtLink
        v-else-if="route.path !== '/_login'"
        v-slot="{ navigate, href }"
        custom
        :to="`/_login?returnTo=${route.query.returnTo || route.fullPath}`"
      >
        <ElLink :underline="false" :href="href" @click="navigate">
          <Icon name="ic:round-log-in" class="mr-1" /> <span class="font-normal">{{ $t('sign-in') }}</span>
        </ElLink>
      </NuxtLink>
    </span>
  </div>
</template>
