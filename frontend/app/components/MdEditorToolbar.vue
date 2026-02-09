<script setup lang="ts">
defineProps<{
  showPreview: boolean
  showFullscreen: boolean
}>()

const emit = defineEmits<{
  (e: 'click', payload: string): void
}>()
</script>

<template>
  <div class="flex justify-between p-1">
    <div class="flex items-center gap-0.5 flex-wrap">
      <MdToolbarButton icon="tabler:bold" @click="emit('click', 'bold')" />
      <MdToolbarButton icon="tabler:italic" @click="emit('click', 'italic')" />
      <MdToolbarButton icon="tabler:underline" @click="emit('click', 'underline')" />
      <MdToolbarButton icon="tabler:strikethrough" @click="emit('click', 'strikethrough')" />
      <MdToolbarButton icon="tabler:superscript" @click="emit('click', 'superscript')" />
      <MdToolbarButton icon="tabler:subscript" @click="emit('click', 'subscript')" />
      <span class="text-[var(--ui-text)]/30 mx-1">|</span>

      <UPopover
        mode="hover"
        style="font-size:inherit;color:inherit;line-height:inherit;display:inline-block"
      >
        <MdToolbarButton icon="tabler:heading" />
        <template #content>
          <div class="flex gap-0.5 p-1">
            <MdToolbarButton icon="tabler:h-1" style="font-size:initial" @click="emit('click', 'heading-1')" />
            <MdToolbarButton icon="tabler:h-2" style="font-size:initial" @click="emit('click', 'heading-2')" />
            <MdToolbarButton icon="tabler:h-3" style="font-size:initial" @click="emit('click', 'heading-3')" />
            <MdToolbarButton icon="tabler:h-4" style="font-size:initial" @click="emit('click', 'heading-4')" />
            <MdToolbarButton icon="tabler:h-5" style="font-size:initial" @click="emit('click', 'heading-5')" />
            <MdToolbarButton icon="tabler:h-6" style="font-size:initial" @click="emit('click', 'heading-6')" />
          </div>
        </template>
      </UPopover>

      <MdToolbarButton icon="tabler:list" @click="emit('click', 'list-unordered')" />
      <MdToolbarButton icon="tabler:list-numbers" @click="emit('click', 'list-ordered')" />
      <MdToolbarButton icon="tabler:checkbox" @click="emit('click', 'checkbox')" />
      <MdToolbarButton icon="tabler:quote" @click="emit('click', 'quote')" />

      <span class="text-[var(--ui-text)]/30 mx-1">|</span>
      <MdToolbarButton icon="tabler:link" @click="emit('click', 'link')" />
      <MdToolbarButton icon="tabler:code" @click="emit('click', 'code-inline')" />
      <MdToolbarButton icon="tabler:braces" @click="emit('click', 'code-block')" />
      <MdToolbarButton icon="tabler:table" @click="emit('click', 'table')" />
    </div>
    <div class="flex items-center gap-0.5">
      <MdToolbarButton
        :icon="showFullscreen ? 'tabler:minimize' : 'tabler:maximize'"
        @click="emit('click', 'fullscreen')"
      />
      <MdToolbarButton
        :icon="showPreview ? 'tabler:eye-off' : 'tabler:eye'"
        @click="emit('click', 'preview')"
      />
    </div>
  </div>
</template>
