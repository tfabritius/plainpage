<script setup lang="ts">
import { format } from 'date-fns'
import type { GetAtticListResponse } from '~/types'
import { Icon } from '#components'

const props = defineProps<{
  urlPath: string
  title: string
}>()

const urlPath = computed(() => props.urlPath)

useHead(() => ({ title: `Revisions: ${props.title}` }))

const { data } = await useAsyncData(`/attic${urlPath.value}`, async () => {
  const data = await apiFetch<GetAtticListResponse>(`/attic${urlPath.value}`)

  const entries = data.entries.map(e => ({ ...e, date: new Date(e.rev * 1000) }))
    .sort((a, b) => b.rev - a.rev)

  return {
    entries,
    breadcrumbs: data.breadcrumbs,
  }
})
</script>

<template>
  <Layout :breadcrumbs="data?.breadcrumbs">
    <template #title>
      Old revisions of: {{ title }}
    </template>

    <template #actions>
      <ElButton class="m-1" @click="navigateTo({ query: { rev: undefined } })">
        <Icon name="ic:baseline-update" /> <span class="hidden md:inline ml-1">Current version</span>
      </ElButton>
    </template>

    <div v-for="(el, idx) in data?.entries" :key="el.rev">
      <NuxtLink v-slot="{ navigate, href }" :to="`?rev=${el.rev}`" custom>
        <ElLink :href="href" @click="navigate">
          {{ format(el.date, 'yyyy-MM-dd HH:mm') }} ({{ el.rev }})
          <span v-if="idx === 0"> <Icon class="ml-2" name="ci:show" /> (current version)</span>
        </ElLink>
      </NuxtLink>
    </div>
  </Layout>
</template>
