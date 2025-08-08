<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type { User } from '~/types'
import { FetchError } from 'ofetch'
import { z } from 'zod'
import { useAuthStore } from '~/store/auth'
import { validUsernameRegex } from '~/types'

const { t } = useI18n()
const toast = useToast()

const existingUsername = ref<string>()

const form = useTemplateRef('formRef')
const formSchema = z.object({
  username: z.string()
    .trim()
    .min(4, t('username-length'))
    .max(20, t('username-length'))
    .regex(validUsernameRegex, t('username-invalid'))
    .refine(username => username !== existingUsername.value, { message: t('username-not-unique') }),
  displayName: z.string().min(1, t('displayname-required')),
  password: z.string().min(1, t('password-required')),
  passwordConfirm: z.string(),
})
  .refine(({ password, passwordConfirm }) => password === passwordConfirm, { message: t('password-repeat-not-equal'), path: ['passwordConfirm'] })

type FormSchema = z.output<typeof formSchema>
const formState = reactive<FormSchema>(
  { displayName: '', username: '', password: '', passwordConfirm: '' },
)

useHead(() => ({ title: t('register') }))

const auth = useAuthStore()
const route = useRoute()

const loading = ref(false)

async function submit(_event: FormSubmitEvent<FormSchema>) {
  loading.value = true
  try {
    await apiFetch<User>('/auth/users', { method: 'POST', body: formState })
  } catch (e) {
    if (e instanceof FetchError && e.statusCode === 409) {
      existingUsername.value = formState.username
      form.value?.validate()
    } else {
      toast.add({ description: String(e), color: 'error' })
    }
    loading.value = false
    return
  }

  const ok = await auth.login(formState)
  if (ok) {
    const returnTo = typeof route.query.returnTo === 'string' ? route.query.returnTo : '/'
    await navigateTo(returnTo)
  } else {
    toast.add({ description: t('invalid-credentials'), color: 'error' })
  }

  loading.value = false
}
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-gray-500">
      <h2 class="font-light text-xl my-4">
        {{ $t('register-account') }}
      </h2>

      <UForm
        ref="formRef"
        :schema="formSchema"
        :state="formState"
        class="w-50"
        @submit="submit"
      >
        <UFormField name="displayName" class="mt-4">
          <UInput
            v-model="formState.displayName"
            :placeholder="$t('display-name')"
            autofocus
            class="w-full"
          />
        </UFormField>
        <UFormField name="username" class="mt-4">
          <UInput
            v-model="formState.username"
            type="text"
            :placeholder="$t('username')"
            class="w-full"
          />
        </UFormField>
        <UFormField name="password" class="mt-4">
          <UInput
            v-model="formState.password"
            type="password"
            :placeholder="$t('password')"
            class="w-full"
          />
        </UFormField>
        <UFormField name="passwordConfirm" class="mt-4">
          <UInput
            v-model="formState.passwordConfirm"
            type="password"
            :placeholder="$t('password-repeat')"
            class="w-full"
          />
        </UFormField>
        <UFormField class="mt-4">
          <UButton
            type="submit"
            color="primary"
            variant="solid"
            class="w-full"
            :label="$t('register')"
            :loading="loading"
          />
        </UFormField>
      </UForm>
    </div>
  </div>
</template>
