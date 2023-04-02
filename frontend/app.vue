<script setup lang="ts">
import { useAuthStore } from '~/store/auth'
import { Icon } from '#components'
import type { GetAppResponse } from '~/types'

useHead({
  bodyAttrs: {
    class: 'font-sans m-0',
  },
  link: [{
    rel: 'icon',
    href: '/favicon.svg',
  }],
})

const { data } = await useAsyncData('/app', () => apiFetch<GetAppResponse>('/app'))

const appName = computed(() => data.value?.appName ?? 'PlainPage')

useHead(() => ({ titleTemplate: `%s | ${appName.value}` }))

const auth = useAuthStore()
const route = useRoute()

const UsersIcon = h(Icon, { name: 'ci:users' })
const SettingsIcon = h(Icon, { name: 'ci:settings' })
const LogoutIcon = h(Icon, { name: 'ic:round-log-out' })

async function handleDropdownMenuCommand(command: string | number | object) {
  if (command === 'users') {
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
            <span>{{ appName }}</span>
          </span>
        </ElLink>
      </NuxtLink>

      <span v-if="auth.loggedIn">
        <ElDropdown trigger="click" class="m-1" @command="handleDropdownMenuCommand">
          <ElLink :underline="false" href="#">
            <Icon name="ci:user" class="mr-1" />
            <span class="font-normal">{{ auth.user?.realName }}</span>
          </ElLink>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="UsersIcon" command="users">
                Users
              </ElDropdownItem>
              <ElDropdownItem :icon="SettingsIcon" command="settings">
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
