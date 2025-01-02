<script setup lang="ts">
import slugify from 'slugify'
import { z } from 'zod'

const props = defineProps<{
  type: 'page' | 'folder'
  urlPath: string
}>()
const toast = useToast()
const { t } = useI18n()

const open = ref(false)
const form = useTemplateRef('formRef')
const titleInput = useTemplateRef('titleInputRef')

const formSchema = z.object({
  title: z.string(),
  name: z.string()
    .min(1, props.type === 'page' ? t('page-name-required') : t('folder-name-required'))
    .regex(/^[a-z0-9-][a-z0-9_-]*$/, props.type === 'page' ? t('invalid-page-name') : t('invalid-folder-name')),
})
type FormSchema = z.output<typeof formSchema>

const formState = reactive<FormSchema>({ title: '', name: '' })
const expanded = ref(false)
const nameChangedManually = ref(false)

function onTitleChanged(ev: InputEvent) {
  if (!nameChangedManually.value) {
    formState.name = slugify(
      (ev.target as HTMLInputElement).value,
      { lower: true, strict: true },
    )
  }
}

function onNameChanged() {
  nameChangedManually.value = true
}

async function submit() {
  const newUrl = `${props.urlPath !== '' ? `/${props.urlPath}` : ''}/${formState.name}`

  if (props.type === 'page') {
    await navigateTo({
      path: newUrl,
      query: { edit: 'true' },
      state: { title: formState.title },
    })
  } else {
    try {
      await apiFetch(`/pages${newUrl}`, {
        method: 'PUT',
        body: { folder: { meta: { title: formState.title } } },
      })
    } catch (err) {
      toast.add({
        description: String(err),
        color: 'error',
      })
    }
    await navigateTo(newUrl)
  }
}

async function onOpenChanged(open: boolean) {
  if (open) {
    formState.title = ''
    formState.name = ''
    expanded.value = false
    nameChangedManually.value = false
    form.value?.clear()
    await nextTick()
    titleInput.value?.inputRef?.focus()
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="props.type === 'page' ? $t('create-page') : $t('create-folder')"
    @update:open="onOpenChanged"
  >
    <slot />
    <template #title />
    <template #body>
      <UForm
        id="newContentForm"
        ref="formRef"
        :state="formState"
        :schema="formSchema"
        @submit="submit"
        @error="expanded = true"
      >
        <UFormField
          :label="props.type === 'page' ? $t('page-title') : $t('folder-title')"
          name="title"
        >
          <UInput
            ref="titleInputRef"
            v-model="formState.title"
            class="w-full"
            @input="onTitleChanged"
          />
        </UFormField>

        <UCollapsible v-model:open="expanded">
          <span class="cursor-pointer text-xs">
            <UIcon name="ci:chevron-right" :class="expanded && 'rotate-90 transform'" />
            {{ $t('more') }}
          </span>
          <template #content>
            <UFormField
              :label="props.type === 'page' ? $t('page-name') : $t('folder-name')"
              name="name"
            >
              <UInput
                v-model="formState.name"
                class="w-full"
                @input="onNameChanged"
              />
            </UFormField>
          </template>
        </UCollapsible>
      </UForm>
    </template>
    <template #footer>
      <UButton :label="$t('cancel')" @click="open = false" />
      <UButton color="primary" variant="solid" :label="$t('ok')" type="submit" form="newContentForm" />
    </template>
  </UModal>
</template>
