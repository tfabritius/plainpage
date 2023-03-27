<script setup lang="ts">
const route = useRoute()
</script>

<template>
  <ElCard>
    <template #header>
      <slot name="breadcrumbs" />

      <div class="flex justify-between items-center">
        <NuxtLink v-slot="{ navigate, href }" custom :to="route.path">
          <ElLink :href="href" :underline="false" @click="navigate">
            <h1 class="hover:underline font-light flex items-center">
              <slot name="title" />
            </h1>
          </ElLink>
        </NuxtLink>

        <div class="flex items-center">
          <slot name="actions" />
        </div>
      </div>
    </template>

    <slot />
  </ElCard>
</template>
