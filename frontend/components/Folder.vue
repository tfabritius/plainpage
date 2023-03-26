<script setup lang="ts">
import type { FolderEntry } from '~/types/'

const props = defineProps<{
  folder: FolderEntry[]
}>()

const folder = computed(() => props.folder)
</script>

<template>
  <div>
    <h2 v-if="folder.some(e => e.isFolder)" class="font-light text-xl">
      Folders
    </h2>
    <div v-for="entry of folder.filter(e => e.isFolder)" :key="entry.name">
      <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
        <ElLink :href="href" @click="navigate">
          <Icon name="ci:folder" class="mr-1" /> {{ entry.name }}
        </ElLink>
      </NuxtLink>
    </div>

    <h2
      v-if="folder.some(e => !e.isFolder)" class="font-light text-xl"
    >
      Pages
    </h2>
    <div v-for="entry of folder.filter(e => !e.isFolder)" :key="entry.name">
      <NuxtLink v-slot="{ navigate, href }" :to="entry.url" custom>
        <ElLink :href="href" @click="navigate">
          {{ entry.name }}
        </ElLink>
      </NuxtLink>
    </div>
  </div>
</template>
