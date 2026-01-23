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
      <UIcon name="ci:user-circle" size="5em" class="mb-3" />

      <form class="w-52" @submit.prevent @keypress.enter="submit">
        <UFormField>
          <UInput
            v-model="loginFormData.username"
            type="text"
            :placeholder="$t('username')"
            autofocus
            class="w-full"
          />
        </UFormField>
        <UFormField class="mt-4">
          <UInput
            v-model="loginFormData.password"
            type="password"
            :placeholder="$t('password')"
            class="w-full"
          />
        </UFormField>
        <UFormField class="mt-4">
          <UButton
            color="primary"
            variant="solid"
            class="w-full"
            :label="$t('sign-in')"
            :loading="loading"
            @click="submit"
          />
        </UFormField>
      </form>

      <ULink
        v-if="allowRegister"
        :to="`_register?returnTo=${String(route.query.returnTo || '/')}`"
        class="text-sm"
      >
        {{ $t('_login.link-to-register') }}
      </ULink>
    </div>

    <AppFooter />
  </div>
</template>
