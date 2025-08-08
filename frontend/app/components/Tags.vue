<script lang="ts" setup>
import { nextTick, ref } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: string[] | null
  editable?: boolean
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
  <div class="flex gap-1">
    <UBadge
      v-for="tag in modelValue"
      :key="tag"
      variant="outline"
    >
      {{ tag }}
      <UButton
        v-if="editable"
        icon="ci:close-md"
        size="xs"
        color="primary"
        variant="link"
        :ui="{ base: 'p-0' }"
        @click="handleClose(tag)"
      />
    </UBadge>

    <UInput
      v-if="inputVisible"
      ref="InputRef"
      v-model="inputValue"
      class="max-w-20 inline leading-4"
      size="xs"
      @keyup.enter="onInputConfirm"
      @blur="onInputConfirm"
      @keyup.esc.stop="onCancelInput"
    />

    <ReactiveButton
      v-if="editable && !inputVisible"
      color="primary"
      class="px-2"
      icon="ci:plus"
      :label="$t('add-tag')"
      size="xs"
      @click="showInput"
    />
  </div>
</template>
