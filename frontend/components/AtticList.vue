<script setup lang="ts">
import { format } from 'date-fns'
import type { AtticEntry, Breadcrumb } from '~/types'
import { Icon } from '#components'

const props = defineProps<{
  urlPath: string
  title: string
  breadcrumbs: Breadcrumb[]
}>()

const urlPath = computed(() => props.urlPath)

const { data } = await useAsyncData(`/api/_attic${urlPath.value}`, async () => {
  const data = await $fetch<AtticEntry[]>(`/_api/attic${urlPath.value}`)

  return data
    .map(e => ({ ...e, date: new Date(e.rev * 1000) }))
    .sort((a, b) => b.rev - a.rev)
})
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      Old revisions of: {{ title }}
    </template>

    <template #actions>
      <ElButton class="m-1" @click="navigateTo({ query: { rev: undefined } })">
        <Icon name="ic:baseline-update" /> <span class="hidden md:inline ml-1">Current version</span>
      </ElButton>
    </template>

    <div v-for="el in data" :key="el.rev">
      <NuxtLink v-slot="{ navigate, href }" :to="`?rev=${el.rev}`" custom>
        <ElLink :href="href" @click="navigate">
          {{ format(el.date, 'yyyy-MM-dd HH:mm') }} ({{ el.rev }})
        </ElLink>
      </NuxtLink>
    </div>
  </Layout>
</template>
