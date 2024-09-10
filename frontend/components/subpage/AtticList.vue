<script setup lang="ts">
import { Icon } from '#components'
import { UseTimeAgo } from '@vueuse/components'
import { format } from 'date-fns'
import { timeAgoMessages } from '~/composables/timeAgoMessages'
import type { GetAtticListResponse } from '~/types'

const props = defineProps<{
  urlPath: string
  title: string
}>()

const { t } = useI18n()

useHead(() => ({ title: `${t('revisions')}: ${props.title}` }))

const { data } = await useAsyncData(`/attic/${props.urlPath}`, async () => {
  const data = await apiFetch<GetAtticListResponse>(`/attic/${props.urlPath}`)

  const entries = data.entries.map(e => ({ ...e, date: new Date(e.rev * 1000) }))
    .sort((a, b) => b.rev - a.rev)

  return {
    entries,
    breadcrumbs: data.breadcrumbs,
  }
})

const navTo = navigateTo
</script>

<template>
  <Layout :breadcrumbs="data?.breadcrumbs">
    <template #title>
      {{ $t('old-revisions-of') }}: {{ title }}
    </template>

    <template #actions>
      <PlainButton
        icon="ic:baseline-update"
        :label="$t('current-version')"
        @click="navTo({ query: { rev: undefined } })"
      />
    </template>

    <div v-for="(el, idx) in data?.entries" :key="el.rev">
      <NuxtLink v-slot="{ navigate, href }" :to="`?rev=${el.rev}`" custom>
        <ElLink :href="href" @click="navigate">
          {{ format(el.date, 'yyyy-MM-dd HH:mm:ss') }}
          <UseTimeAgo v-slot="{ timeAgo }" :time="el.date" :messages="timeAgoMessages()">
            ({{ timeAgo }})
          </UseTimeAgo>
          <span v-if="idx === 0"> <Icon class="ml-2" name="ci:show" /> ({{ $t('current-version') }})</span>
        </ElLink>
      </NuxtLink>
    </div>
  </Layout>
</template>
