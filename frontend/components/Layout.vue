<script setup lang="ts">
import type { Breadcrumb } from '~/types'
import { Icon } from '#components'

const _props = defineProps<{
  breadcrumbs?: Breadcrumb[]
  useFullHeight?: boolean
}>()

const ChevronIcon = h(Icon, { name: 'ci:chevron-right' })

const route = useRoute()
</script>

<template>
  <div
    class="box-border p-2 flex flex-col"
    :class="{
      'min-h-screen': !useFullHeight,
      'h-screen': useFullHeight,
    }"
  >
    <AppHeader />

    <div
      class="border rounded border-gray-300 border-solid flex flex-col min-h-0"
      :class="{ grow: useFullHeight }"
    >
      <div class="p-5 border-b border-b-gray-300 border-b-solid">
        <div v-if="breadcrumbs">
          <ElBreadcrumb :separator-icon="ChevronIcon">
            <ElBreadcrumbItem :to="{ path: '/' }">
              <Icon name="ic:outline-home" />
            </ElBreadcrumbItem>

            <ElBreadcrumbItem
              v-for="crumb in breadcrumbs"
              :key="crumb.url"
              :to="{ path: crumb.url }"
            >
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
      </div>

      <div
        class="p-5 flex flex-col min-h-0"
        :class="{ grow: useFullHeight }"
      >
        <slot />
      </div>
    </div>

    <div class="text-center">
      <ElLink :underline="false" href="https://github.com/tfabritius/plainpage">
        <span
          class="font-normal text-gray-300 dark:text-gray-500 hover:text-current"
        >
          PlainPage
        </span>
      </ElLink>
    </div>
  </div>
</template>
