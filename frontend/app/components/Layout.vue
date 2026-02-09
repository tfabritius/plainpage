<script setup lang="ts">
import type { BreadcrumbItem } from '@nuxt/ui'
import type { Breadcrumb } from '~/types'

const props = defineProps<{
  breadcrumbs?: Breadcrumb[]
  useFullHeight?: boolean
}>()

const route = useRoute()

const breadcrumbItems = computed(() => {
  const items: BreadcrumbItem[] = []

  items.push({ icon: 'tabler:home', class: 'text-[var(--ui-text-muted)]', to: '/' })

  props.breadcrumbs?.forEach((crumb, idx) => {
    if (idx !== (props.breadcrumbs?.length ?? 1) - 1) {
      items.push({ label: crumb.title || crumb.name, to: { path: `/${crumb.url}` } })
    } else {
      items.push({ label: crumb.title || crumb.name, to: { path: `/${crumb.url}` }, class: 'text-[var(--ui-text-muted)]' })
    }
  })

  return items
})
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
      class="border rounded-xl border-gray-300 border-solid shadow-sm flex flex-col min-h-0"
      :class="{ grow: useFullHeight }"
    >
      <div class="p-5 border-b border-b-gray-300 border-b-solid">
        <div v-if="breadcrumbs">
          <UBreadcrumb
            :items="breadcrumbItems"
          />
        </div>

        <div class="flex justify-between items-center">
          <div>
            <span class="group flex items-top">
              <ULink :to="route.path" :active="false">
                <h1 class="my-0 py-2 text-2xl md:text-3xl font-light flex items-center">
                  <slot name="title" />
                </h1>
              </ULink>
              <slot name="title:suffix" />
            </span>

            <div class="flex items-center text-sm">
              <slot name="subtitle" />
            </div>
          </div>

          <div class="flex items-center gap-2 md:gap-3">
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

    <AppFooter />
  </div>
</template>
