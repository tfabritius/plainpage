<script setup lang="ts">
defineProps<{
  showPreview: boolean
  showFullscreen: boolean
}>()

const emit = defineEmits<{
  (e: 'click', payload: string): void
}>()

const { t } = useI18n()
</script>

<template>
  <div class="flex justify-between p-1">
    <div class="flex items-center gap-0.5 flex-wrap">
      <MdToolbarButton icon="tabler:bold" :tooltip="t('editor.bold')" @click="emit('click', 'bold')" />
      <MdToolbarButton icon="tabler:italic" :tooltip="t('editor.italic')" @click="emit('click', 'italic')" />
      <MdToolbarButton icon="tabler:underline" :tooltip="t('editor.underline')" @click="emit('click', 'underline')" />
      <MdToolbarButton icon="tabler:strikethrough" :tooltip="t('editor.strikethrough')" @click="emit('click', 'strikethrough')" />
      <MdToolbarButton icon="tabler:superscript" :tooltip="t('editor.superscript')" @click="emit('click', 'superscript')" />
      <MdToolbarButton icon="tabler:subscript" :tooltip="t('editor.subscript')" @click="emit('click', 'subscript')" />
      <span class="text-(--ui-text)/30 mx-1">|</span>

      <UPopover
        mode="hover"
        style="font-size:inherit;color:inherit;line-height:inherit;display:inline-block"
      >
        <MdToolbarButton icon="tabler:heading" :tooltip="t('editor.heading')" />
        <template #content>
          <div class="flex gap-0.5 p-1">
            <MdToolbarButton icon="tabler:h-1" :tooltip="t('editor.heading-1')" style="font-size:initial" @click="emit('click', 'heading-1')" />
            <MdToolbarButton icon="tabler:h-2" :tooltip="t('editor.heading-2')" style="font-size:initial" @click="emit('click', 'heading-2')" />
            <MdToolbarButton icon="tabler:h-3" :tooltip="t('editor.heading-3')" style="font-size:initial" @click="emit('click', 'heading-3')" />
            <MdToolbarButton icon="tabler:h-4" :tooltip="t('editor.heading-4')" style="font-size:initial" @click="emit('click', 'heading-4')" />
            <MdToolbarButton icon="tabler:h-5" :tooltip="t('editor.heading-5')" style="font-size:initial" @click="emit('click', 'heading-5')" />
            <MdToolbarButton icon="tabler:h-6" :tooltip="t('editor.heading-6')" style="font-size:initial" @click="emit('click', 'heading-6')" />
          </div>
        </template>
      </UPopover>

      <MdToolbarButton icon="tabler:list" :tooltip="t('editor.list-unordered')" @click="emit('click', 'list-unordered')" />
      <MdToolbarButton icon="tabler:list-numbers" :tooltip="t('editor.list-ordered')" @click="emit('click', 'list-ordered')" />
      <MdToolbarButton icon="tabler:checkbox" :tooltip="t('editor.checkbox')" @click="emit('click', 'checkbox')" />
      <MdToolbarButton icon="tabler:quote" :tooltip="t('editor.quote')" @click="emit('click', 'quote')" />

      <span class="text-(--ui-text)/30 mx-1">|</span>
      <MdToolbarButton icon="tabler:link" :tooltip="t('editor.link')" @click="emit('click', 'link')" />
      <MdToolbarButton icon="tabler:code" :tooltip="t('editor.code-inline')" @click="emit('click', 'code-inline')" />
      <MdToolbarButton icon="tabler:braces" :tooltip="t('editor.code-block')" @click="emit('click', 'code-block')" />
      <MdToolbarButton icon="tabler:table" :tooltip="t('editor.table')" @click="emit('click', 'table')" />
      <span class="text-(--ui-text)/30 mx-1">|</span>
      <MdToolbarButton icon="tabler:wand" :tooltip="t('editor.format')" @click="emit('click', 'format')" />
    </div>
    <div class="flex items-center gap-0.5">
      <MdToolbarButton
        :icon="showFullscreen ? 'tabler:minimize' : 'tabler:maximize'"
        :tooltip="showFullscreen ? t('editor.minimize') : t('editor.fullscreen')"
        @click="emit('click', 'fullscreen')"
      />
      <MdToolbarButton
        :icon="showPreview ? 'tabler:eye-off' : 'tabler:eye'"
        :tooltip="showPreview ? t('editor.hide-preview') : t('editor.show-preview')"
        @click="emit('click', 'preview')"
      />
    </div>
  </div>
</template>
