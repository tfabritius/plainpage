<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useAuthStore } from '~/store/auth'
import { Icon } from '#components'
import { useAppStore } from '~/store/app'

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
const { appTitle, allowAdmin } = storeToRefs(app)

refresh()
watch(() => auth.loggedIn, () => refresh())

useHead(() => ({ titleTemplate: `%s | ${appTitle.value}` }))

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
</script>

<template>
  <div class="p-2">
    <div class="flex justify-between">
      <NuxtLink v-slot="{ navigate, href }" custom to="/">
        <ElLink :underline="false" :href="href" @click="navigate">
          <span class="text-xl font-light flex items-center">
            <Icon name="ci:file-blank" />
            <span>{{ appTitle }}</span>
          </span>
        </ElLink>
      </NuxtLink>

      <span v-if="auth.loggedIn">
        <ElDropdown trigger="click" class="m-1" @command="handleDropdownMenuCommand">
          <ElLink :underline="false" href="#">
            <Icon name="ci:user" class="mr-1" />
            <span class="font-normal">{{ auth.user?.displayName }}</span>
          </ElLink>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="ProfileIcon" command="profile">
                Profile
              </ElDropdownItem>
              <ElDropdownItem v-if="allowAdmin" :icon="UsersIcon" command="users">
                Users
              </ElDropdownItem>
              <ElDropdownItem v-if="allowAdmin" :icon="SettingsIcon" command="settings">
                Settings
              </ElDropdownItem>
              <ElDropdownItem :icon="LogoutIcon" command="logout">
                Sign out
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </span>

      <NuxtLink v-else v-slot="{ navigate, href }" custom :to="`/_login?returnTo=${encodeURIComponent(route.fullPath)}`">
        <ElLink :underline="false" :href="href" @click="navigate">
          <Icon name="ic:round-log-in" class="mr-1" /> <span class="font-normal">Sign in</span>
        </ElLink>
      </NuxtLink>
    </div>
    <NuxtPage />
    <NuxtLoadingIndicator />
  </div>
</template>
