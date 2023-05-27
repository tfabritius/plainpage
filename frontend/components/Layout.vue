<script setup lang="ts">
import type { Breadcrumb } from '~/types'
import { Icon } from '#components'

const _props = defineProps<{
  breadcrumbs?: Breadcrumb[]
}>()

const ChevronIcon = h(Icon, { name: 'ci:chevron-right' })

const route = useRoute()
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <ElCard>
      <template #header>
        <div v-if="breadcrumbs">
          <ElBreadcrumb :separator-icon="ChevronIcon">
            <ElBreadcrumbItem :to="{ path: '/' }">
              <Icon name="ic:outline-home" />
            </ElBreadcrumbItem>

            <ElBreadcrumbItem v-for="crumb in breadcrumbs" :key="crumb.url" :to="{ path: crumb.url }">
              {{ crumb.name }}
            </ElBreadcrumbItem>
          </ElBreadcrumb>
        </div>

        <div class="flex justify-between items-center">
          <div>
            <NuxtLink v-slot="{ navigate, href }" custom :to="route.path">
              <ElLink :href="href" :underline="false" @click="navigate">
                <h1 class="hover:underline my-0 py-3 font-light flex items-center">
                  <slot name="title" />
                </h1>
              </ElLink>
            </NuxtLink>

            <div class="flex items-center text-sm">
              <slot name="subtitle" />
            </div>
          </div>

          <div class="flex items-center">
            <slot name="actions" />
          </div>
        </div>
      </template>

      <slot />
    </ElCard>

    <div class="text-center">
      <ElLink :underline="false" href="https://github.com/tfabritius/plainpage">
        <span class="font-normal text-gray-300 dark:text-gray-500 hover:text-current">PlainPage</span>
      </ElLink>
    </div>
  </div>
</template>
