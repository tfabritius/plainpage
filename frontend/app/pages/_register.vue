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
      form.value?.validate({})
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
      <UCard class="w-80">
        <div class="text-center mb-4">
          <div class="w-16 h-16 rounded-2xl bg-[var(--ui-primary)] mx-auto flex items-center justify-center mb-4 shadow-lg">
            <UIcon name="tabler:user-plus" size="2em" class="text-white" />
          </div>
          <h1 class="text-xl font-semibold text-slate-700 dark:text-slate-200">
            {{ $t('register-account') }}
          </h1>
        </div>

        <UForm
          ref="formRef"
          :schema="formSchema"
          :state="formState"
          @submit="submit"
        >
          <UFormField name="displayName">
            <UInput
              v-model="formState.displayName"
              :placeholder="$t('display-name')"
              autofocus
              class="w-full"
              size="lg"
            />
          </UFormField>
          <UFormField name="username" class="mt-4">
            <UInput
              v-model="formState.username"
              type="text"
              :placeholder="$t('username')"
              class="w-full"
              size="lg"
            />
          </UFormField>
          <UFormField name="password" class="mt-4">
            <UInput
              v-model="formState.password"
              type="password"
              :placeholder="$t('password')"
              class="w-full"
              size="lg"
            />
          </UFormField>
          <UFormField name="passwordConfirm" class="mt-4">
            <UInput
              v-model="formState.passwordConfirm"
              type="password"
              :placeholder="$t('password-repeat')"
              class="w-full"
              size="lg"
            />
          </UFormField>
          <UFormField class="mt-4">
            <UButton
              type="submit"
              color="primary"
              variant="solid"
              class="w-full"
              :label="$t('register')"
              size="lg"
              :loading="loading"
            />
          </UFormField>
        </UForm>

        <div class="mt-6 pt-4 border-t border-slate-200/80 dark:border-slate-700/50 text-center">
          <ULink
            :to="`_login?returnTo=${String(route.query.returnTo || '/')}`"
            class="text-sm text-slate-500 dark:text-slate-400 hover:text-[var(--ui-primary)]"
          >
            {{ $t('_register.link-to-login') }}
          </ULink>
        </div>
      </UCard>
    </div>

    <AppFooter />
  </div>
</template>
