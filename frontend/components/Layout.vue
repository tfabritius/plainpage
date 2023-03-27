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
              <h1 class="hover:underline font-light flex items-center">
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
</template>
