<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAppStore } from '~/store/app'
import { useAuthStore } from '~/store/auth'

const { t } = useI18n()

const loginFormData = ref({ username: '', password: '' })

useHead(() => ({ title: t('sign-in') }))

const auth = useAuthStore()
const route = useRoute()

const app = useAppStore()
const { allowRegister } = storeToRefs(app)

const loading = ref(false)

const toast = useToast()

async function submit() {
  loading.value = true

  try {
    const loginResponse = await auth.login(loginFormData.value)

    if (loginResponse === true) {
      const returnTo = typeof route.query.returnTo === 'string' ? route.query.returnTo : '/'
      await navigateTo(returnTo)
    } else {
      if (loginResponse.statusCode === 401) {
        toast.add({ description: t('invalid-credentials'), color: 'error' })
      } else if (loginResponse.statusCode === 429) {
        toast.add({ description: t('too-many-login-requests', [loginResponse.retryAfter]), color: 'error' })
      }

      loading.value = false
    }
  } catch (err) {
    toast.add({ description: String(err), color: 'error' })
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-gray-500">
      <UCard class="w-80" variant="elevated">
        <div class="text-center mb-4">
          <div class="w-16 h-16 rounded-2xl bg-[var(--ui-primary)] mx-auto flex items-center justify-center mb-4 shadow-lg">
            <UIcon name="tabler:user" size="2em" class="text-white" />
          </div>
          <h1 class="text-xl font-semibold text-slate-700 dark:text-slate-200">
            {{ $t('sign-in') }}
          </h1>
        </div>

        <form @submit.prevent @keypress.enter="submit">
          <UFormField>
            <UInput
              v-model="loginFormData.username"
              type="text"
              :placeholder="$t('username')"
              autofocus
              class="w-full"
              size="lg"
            />
          </UFormField>
          <UFormField class="mt-4">
            <UInput
              v-model="loginFormData.password"
              type="password"
              :placeholder="$t('password')"
              class="w-full"
              size="lg"
            />
          </UFormField>
          <UFormField class="mt-4">
            <UButton
              color="primary"
              variant="solid"
              class="w-full"
              :label="$t('sign-in')"
              size="lg"
              :loading="loading"
              @click="submit"
            />
          </UFormField>
        </form>

        <div
          v-if="allowRegister"
          class="mt-6 pt-4 border-t border-default text-center"
        >
          <ULink
            :to="`_register?returnTo=${String(route.query.returnTo || '/')}`"
            class="text-sm text-slate-500 dark:text-slate-400 hover:text-[var(--ui-primary)]"
          >
            {{ $t('_login.link-to-register') }}
          </ULink>
        </div>
      </UCard>
    </div>

    <AppFooter />
  </div>
</template>
