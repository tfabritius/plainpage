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

const profileSchema = z.object({
  displayName: z.string().min(1, t('displayname-required')),
})

type ProfileSchema = z.output<typeof profileSchema>
const profileState = reactive<ProfileSchema>({
  displayName: auth.user?.displayName || '',
})

async function onSaveProfile() {
  try {
    await auth.updateMe({ displayName: profileState.displayName })
    toast.add({ description: t('saved'), color: 'success' })
  } catch (err) {
    toast.add({ description: String(err), color: 'error' })
  }
}

const passwordExpanded = ref(false)
const passwordSchema = z.object({
  currentPassword: z.string().min(1, t('current-password-required')),
  password: z.string().min(1, t('new-password-required')),
  passwordConfirm: z.string(),
})
  .refine(({ password, passwordConfirm }) => password === passwordConfirm, { message: t('password-repeat-not-equal'), path: ['passwordConfirm'] })

type PasswordSchema = z.output<typeof passwordSchema>
const passwordState = reactive<PasswordSchema>({
  currentPassword: '',
  password: '',
  passwordConfirm: '',
})

async function onSavePassword() {
  // Validate password confirmation before submitting
  const result = passwordSchema.safeParse(passwordState)
  if (!result.success) {
    return
  }

  try {
    await auth.changePassword(passwordState.currentPassword, passwordState.password)
    passwordState.currentPassword = ''
    passwordState.password = ''
    passwordState.passwordConfirm = ''
    passwordExpanded.value = false
    toast.add({ description: t('password-changed'), color: 'success' })
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 403) {
      toast.add({ description: t('incorrect-password'), color: 'error' })
      return
    }
    toast.add({ description: String(err), color: 'error' })
  }
}

function onCancelPassword() {
  passwordExpanded.value = false
  passwordState.currentPassword = ''
  passwordState.password = ''
  passwordState.passwordConfirm = ''
}

const deleteExpanded = ref(false)
const deletePasswordInput = ref('')

async function onDeleteConfirm() {
  if (!deletePasswordInput.value) {
    toast.add({ description: t('current-password-required'), color: 'error' })
    return
  }

  try {
    await auth.deleteMe(deletePasswordInput.value)
    toast.add({ description: t('account-deleted'), color: 'success' })
    await navigateTo('/')
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 403) {
      toast.add({ description: t('incorrect-password'), color: 'error' })
      return
    }
    toast.add({ description: String(err), color: 'error' })
  }
}

function onDeleteCancel() {
  deleteExpanded.value = false
  deletePasswordInput.value = ''
}
</script>

<template>
  <Layout>
    <template #title>
      {{ $t('profile') }}
    </template>

    <template #actions>
      <ReactiveButton color="success" icon="tabler:device-floppy" :label="$t('save')" @click="onSaveProfile" />
    </template>

    <UForm
      :schema="profileSchema"
      :state="profileState"
    >
      <UFormField :label="$t('username')">
        <UInput :model-value="auth.user?.username" :disabled="true" class="w-full" />
      </UFormField>
      <UFormField :label="$t('display-name')" name="displayName" class="mt-4">
        <UInput v-model="profileState.displayName" autocomplete="off" class="w-full" />
      </UFormField>
    </UForm>

    <UCollapsible v-model:open="passwordExpanded" class="mt-8">
      <UButton color="neutral" icon="tabler:key" :label="$t('change-password')" />

      <template #content>
        <UForm
          :schema="passwordSchema"
          :state="passwordState"
          class="mt-4 p-4 border border-default rounded-lg"
        >
          <UFormField :label="$t('current-password')" name="currentPassword">
            <UInput v-model="passwordState.currentPassword" type="password" autocomplete="off" class="w-full" />
          </UFormField>
          <UFormField :label="$t('new-password')" name="password" class="mt-4">
            <UInput v-model="passwordState.password" type="password" autocomplete="off" class="w-full" />
          </UFormField>
          <UFormField :label="$t('password-repeat')" name="passwordConfirm" class="mt-4">
            <UInput v-model="passwordState.passwordConfirm" type="password" autocomplete="off" class="w-full" />
          </UFormField>
          <div class="mt-4 flex gap-2">
            <UButton :label="$t('cancel')" @click="onCancelPassword" />
            <UButton color="success" variant="solid" :label="$t('save')" @click="onSavePassword" />
          </div>
        </UForm>
      </template>
    </UCollapsible>

    <UCollapsible v-model:open="deleteExpanded" class="mt-4">
      <UButton color="warning" icon="tabler:trash" :label="$t('delete-my-account')" />

      <template #content>
        <div class="mt-4 p-4 border border-warning-500 rounded-lg">
          <p class="mb-4">
            {{ $t('are-you-sure-to-delete-this-account') }}
          </p>
          <UFormField :label="$t('current-password')">
            <UInput v-model="deletePasswordInput" type="password" autocomplete="off" class="w-full" />
          </UFormField>
          <div class="mt-4 flex gap-2">
            <UButton :label="$t('cancel')" @click="onDeleteCancel" />
            <UButton color="warning" variant="solid" :label="$t('delete')" @click="onDeleteConfirm" />
          </div>
        </div>
      </template>
    </UCollapsible>
  </Layout>
</template>
