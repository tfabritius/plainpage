<script setup lang="ts">
import type { GetAtticListResponse } from '~/types'
import { UseTimeAgo } from '@vueuse/components'
import { format } from 'date-fns'
import { timeAgoMessages } from '~/composables/timeAgoMessages'

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
      <ULink :to="`?rev=${el.rev}`" :active="false">
        {{ format(el.date, 'yyyy-MM-dd HH:mm:ss') }}
        <UseTimeAgo v-slot="{ timeAgo }" :time="el.date" :messages="timeAgoMessages()">
          ({{ timeAgo }})
        </UseTimeAgo>
        <span v-if="idx === 0">
          <UIcon class="ml-2 align-middle" name="ci:show" />
          <span class="align-middle">
            ({{ $t('current-version') }})
          </span>
        </span>
      </ULink>
    </div>
  </Layout>
</template>
