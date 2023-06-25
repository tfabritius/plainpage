<script setup lang="ts">
/**
 * This component is meant to be used together with
 * @unocss/preset-icons (prefix: '') like this:
 * <Icon name="carbon-accessibility" />
 */

defineProps<{
  name: string
  size?: string
}
>()
</script>

<template>
  <span
    class="icon"
    :class="name"
    :style="{ width: size, height: size }"
  />
</template>

<style scoped>
.icon {
  display: inline-block;
  vertical-align: middle;
}
</style>
