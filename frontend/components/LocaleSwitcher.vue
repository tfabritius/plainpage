<script setup lang="ts">
import { Icon } from '#components'

const { locale, setLocaleCookie } = useI18n()

const locales = [
  {
    code: 'en',
    name: 'English',
    icon: 'flagpack:us',
  },
  {
    code: 'de',
    name: 'Deutsch',
    icon: 'flagpack:de',
  },
  {
    code: 'es',
    name: 'Español',
    icon: 'flagpack:es',
  },
]

const createIcon = (name: string) => h(Icon, { name })

async function handleCommand(command: string | number | object) {
  if (typeof command === 'string' && locales.map(l => l.code).includes(command)) {
    locale.value = command
    setLocaleCookie(command)
  } else {
    throw new Error(`Unhandled command ${command}`)
  }
}
</script>

<template>
  <ElDropdown trigger="click" @command="handleCommand">
    <ElLink href="#" :underline="false">
      <Icon :name="locales.find((l) => l.code === locale)?.icon || ''" />
    </ElLink>
    <template #dropdown>
      <ElDropdownMenu>
        <ElDropdownItem v-for="l of locales" :key="l.code" :icon="createIcon(l.icon)" :command="l.code">
          {{ l.name }}
        </ElDropdownItem>
      </ElDropdownMenu>
    </template>
  </ElDropdown>
</template>
