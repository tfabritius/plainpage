<script lang="ts" setup>
import { nextTick, ref } from 'vue'
import { ElInput } from 'element-plus'

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
const InputRef = ref<InstanceType<typeof ElInput>>()

const handleClose = (tag: string) => {
  if (tags.value !== null) {
    tags.value.splice(tags.value.indexOf(tag), 1)
  }
}

const showInput = () => {
  inputVisible.value = true
  nextTick(() => {
    InputRef.value!.input!.focus()
  })
}

const onInputConfirm = () => {
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

const onCancelInput = () => {
  inputVisible.value = false
  inputValue.value = ''
}
</script>

<template>
  <div class="flex">
    <ElTag
      v-for="tag in modelValue"
      :key="tag"
      class="mr-1"
      :closable="editable"
      :disable-transitions="false"
      @close="handleClose(tag)"
    >
      {{ tag }}
    </ElTag>

    <ElInput
      v-if="inputVisible"
      ref="InputRef"
      v-model="inputValue"
      class="max-w-20 inline"
      size="small"
      @keyup.enter="onInputConfirm"
      @blur="onInputConfirm"
      @keyup.esc.stop="onCancelInput"
    />

    <ElButton v-if="editable && !inputVisible" class="w-20" size="small" @click="showInput">
      + {{ $t('add-tag') }}
    </ElButton>
  </div>
</template>
