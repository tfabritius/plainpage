<script setup lang="ts">
/*
   This button shows an icon and a label. On small screens, only the icon is shown.
*/

defineProps<{
  icon: string
  label: string
}>()

// Height of button should be the same if text is shown or not.
//
// original style of md size button: base: 'px-2.5 py-1.5 text-sm gap-1.5',
//
// height of icon + text: 22px + 5 px padding top/bottom = 32px
// height of icon only:   20px + 6 px padding top/bottom = 32px
//
// --> base: 'py-[6px] md:py-[5px]',
//
// add some additional horizontal padding for the icon on small screens:
// --> base: 'px-4 md:px-2.5'
</script>

<template>
  <UButton
    :icon="icon"
    :label="label"
    :ui="{
      label: 'hidden md:inline',
      base: 'px-4 md:px-2.5 py-[6px] md:py-[5px]',
    }"
  />
</template>
