<script setup lang="ts">
import { FetchError } from 'ofetch'
import { z } from 'zod'

import { useAuthStore } from '~/store/auth'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()
const toast = useToast()

useHead({ title: t('profile') })

const auth = useAuthStore()

const formSchema = z.object({
  displayName: z.string().min(1, t('displayname-required')),
  currentPassword: z.string(),
  password: z.string(),
  passwordConfirm: z.string(),
})
  .refine(({ password, passwordConfirm }) => password === passwordConfirm, { message: t('password-repeat-not-equal'), path: ['passwordConfirm'] })
  .refine(({ password, currentPassword }) => !password || currentPassword, { message: t('current-password-required'), path: ['currentPassword'] })

type FormSchema = z.output<typeof formSchema>
const formState = reactive<FormSchema>({
  displayName: auth.user?.displayName || '',
  currentPassword: '',
  password: '',
  passwordConfirm: '',
})

async function onSave() {
  try {
    await auth.updateMe({ displayName: formState.displayName })

    if (formState.password && formState.currentPassword) {
      try {
        await auth.changePassword(formState.currentPassword, formState.password)
      } catch (err) {
        if (err instanceof FetchError && err.statusCode === 403) {
          toast.add({ description: t('incorrect-password'), color: 'error' })
          return
        }
        throw err
      }
    }

    formState.currentPassword = ''
    formState.password = ''
    formState.passwordConfirm = ''
    toast.add({ description: t('saved'), color: 'success' })
  } catch (err) {
    toast.add({ description: String(err), color: 'error' })
  }
}

const plainDialog = useTemplateRef('plainDialog')

async function onDelete() {
  if (!await plainDialog.value?.confirm(
    t('are-you-sure-to-delete-this-account'),
    {
      title: t('delete-my-account'),
      confirmButtonText: t('delete'),
      confirmButtonColor: 'warning',
    },
  )) {
    return
  }

  try {
    await auth.deleteMe()
    toast.add({ description: t('account-deleted'), color: 'success' })
  } catch (err) {
    toast.add({ description: String(err), color: 'error' })
  }
}
</script>

<template>
  <Layout>
    <template #title>
      {{ $t('profile') }}
    </template>

    <template #actions>
      <ReactiveButton color="success" icon="tabler:device-floppy" :label="$t('save')" @click="onSave" />
    </template>

    <UForm
      :schema="formSchema"
      :state="formState"
    >
      <UFormField :label="$t('username')">
        <UInput :model-value="auth.user?.username" :disabled="true" class="w-full" />
      </UFormField>
      <UFormField :label="$t('display-name')" name="displayName" class="mt-4">
        <UInput v-model="formState.displayName" autocomplete="off" class="w-full" />
      </UFormField>
      <UFormField :label="$t('current-password')" name="currentPassword" class="mt-4">
        <UInput v-model="formState.currentPassword" type="password" autocomplete="off" class="w-full" />
      </UFormField>
      <UFormField :label="$t('new-password')" name="password" class="mt-4">
        <UInput v-model="formState.password" type="password" autocomplete="off" class="w-full" />
      </UFormField>
      <UFormField :label="$t('password-repeat')" name="passwordConfirm" class="mt-4">
        <UInput v-model="formState.passwordConfirm" type="password" autocomplete="off" class="w-full" />
      </UFormField>
    </UForm>

    <UButton color="warning" icon="tabler:trash" :label="$t('delete-my-account')" class="mt-8" @click="onDelete" />

    <PlainDialog ref="plainDialog" />
  </Layout>
</template>
