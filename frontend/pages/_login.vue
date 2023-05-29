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

async function submit() {
  loading.value = true
  try {
    const success = await auth.login(loginFormData.value)

    if (success) {
      const returnTo = typeof route.query.returnTo === 'string' ? route.query.returnTo : '/'
      await navigateTo(returnTo)
    } else {
      ElMessage({ message: t('invalid-credentials'), type: 'error' })
    }
  } catch (err) {
    ElMessage({ message: String(err), type: 'error' })
  }
  loading.value = false
}
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-gray-500">
      <Icon name="ci:user-circle" size="5em" class="mb-3" />

      <ElForm label-position="top" class="w-50" @submit.prevent @keypress.enter="submit">
        <ElFormItem>
          <ElInput
            v-model="loginFormData.username"
            type="username"
            :placeholder="$t('username')"
            autofocus
          />
        </ElFormItem>
        <ElFormItem>
          <ElInput
            v-model="loginFormData.password"
            type="password"
            show-password
            :placeholder="$t('password')"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" class="w-full" :loading="loading" @click="submit">
            {{ $t('sign-in') }}
          </ElButton>
        </ElFormItem>
      </ElForm>

      <NuxtLink
        v-if="allowRegister"
        v-slot="{ navigate, href }"
        custom
        :to="`_register?returnTo=${encodeURIComponent(String(route.query.returnTo || '/'))}`"
      >
        <ElLink :underline="false" :href="href" @click="navigate">
          {{ $t('_login.link-to-register') }}
        </ElLink>
      </NuxtLink>
    </div>
  </div>
</template>
