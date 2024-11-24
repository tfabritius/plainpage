<script lang="ts" setup>
import { nextTick, ref } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: string[] | null
  editable: boolean
}>(), { editable: false })

const emit = defineEmits<{
  (e: 'update:modelValue', payload: string[]): void
}>()

const tags = computed({
  get: () => props.modelValue,
  set: (value) => { emit('update:modelValue', value || []) },
})

const inputValue = ref('')
const inputVisible = ref(false)
const input = useTemplateRef('InputRef')

function handleClose(tag: string) {
  if (tags.value !== null) {
    tags.value.splice(tags.value.indexOf(tag), 1)
  }
}

function showInput() {
  inputVisible.value = true
  nextTick(() => {
    input.value?.inputRef?.focus()
  })
}

function onInputConfirm() {
  if (inputValue.value !== '') {
    if (tags.value === null) {
      tags.value = [inputValue.value]
    } else {
      tags.value.push(inputValue.value)
    }
  }

  inputVisible.value = false
  inputValue.value = ''
}

function onCancelInput() {
  inputVisible.value = false
  inputValue.value = ''
}
</script>

<template>
  <div class="flex">
    <UBadge
      v-for="tag in modelValue"
      :key="tag"
      variant="outline"
      class="mr-1"
      :closable="editable"
      @close="handleClose(tag)"
    >
      {{ tag }}
    </UBadge>

    <UInput
      v-if="inputVisible"
      ref="InputRef"
      v-model="inputValue"
      class="max-w-20 inline"
      size="sm"
      @keyup.enter="onInputConfirm"
      @blur="onInputConfirm"
      @keyup.esc.stop="onCancelInput"
    />

    <PlainButton
      v-if="editable && !inputVisible"
      icon="ci:plus"
      :label="$t('add-tag')"
      size="small"
      @click="showInput"
    />
  </div>
</template>
