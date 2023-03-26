<script setup lang="ts">
const emit = defineEmits<{
  (e: 'click', payload: string): void
}>()
</script>

<template>
  <div class="flex justify-between border-b border-b-gray-300 border-b-solid p-1">
    <div>
      <MdToolbarButton icon="lucide:bold" @click="emit('click', 'bold')" />
      <MdToolbarButton icon="lucide:italic" @click="emit('click', 'italic')" />
      <MdToolbarButton icon="lucide:underline" @click="emit('click', 'underline')" />
      <MdToolbarButton icon="lucide:strikethrough" @click="emit('click', 'strikethrough')" />
      <MdToolbarButton icon="lucide:superscript" @click="emit('click', 'superscript')" />
      <MdToolbarButton icon="lucide:subscript" @click="emit('click', 'subscript')" />
      <span class="text-gray mx-2">|</span>

      <ElDropdown trigger="hover" style="font-size:inherit;color:inherit;line-height:inherit">
        <MdToolbarButton icon="lucide:heading" />
        <template #dropdown>
          <MdToolbarButton icon="lucide:heading-1" style="font-size:initial" @click="emit('click', 'heading-1')" />
          <MdToolbarButton icon="lucide:heading-2" style="font-size:initial" @click="emit('click', 'heading-2')" />
          <MdToolbarButton icon="lucide:heading-3" style="font-size:initial" @click="emit('click', 'heading-3')" />
          <MdToolbarButton icon="lucide:heading-4" style="font-size:initial" @click="emit('click', 'heading-4')" />
          <MdToolbarButton icon="lucide:heading-5" style="font-size:initial" @click="emit('click', 'heading-5')" />
          <MdToolbarButton icon="lucide:heading-6" style="font-size:initial" @click="emit('click', 'heading-6')" />
        </template>
      </ElDropdown>

      <MdToolbarButton icon="lucide:list" @click="emit('click', 'list-unordered')" />
      <MdToolbarButton icon="lucide:list-ordered" @click="emit('click', 'list-ordered')" />
      <MdToolbarButton icon="lucide:check-square" @click="emit('click', 'checkbox')" />
      <MdToolbarButton icon="lucide:quote" @click="emit('click', 'quote')" />

      <span class="text-gray mx-2">|</span>
      <MdToolbarButton icon="lucide:link" @click="emit('click', 'link')" />
      <MdToolbarButton icon="lucide:code" @click="emit('click', 'code-inline')" />
      <MdToolbarButton icon="lucide:file-code" @click="emit('click', 'code-block')" />
      <MdToolbarButton icon="lucide:table" @click="emit('click', 'table')" />
    </div>
    <div>
      <MdToolbarButton icon="lucide:eye" @click="emit('click', 'preview')" />
    </div>
  </div>
</template>
