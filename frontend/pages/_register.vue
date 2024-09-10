<script setup lang="ts">
import { FetchError } from 'ofetch'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '~/store/auth'
import type { User } from '~/types'
import { validUsernameRegex } from '~/types'

const { t } = useI18n()

const existingUsername = ref<string>()

const registerFormRef = ref<FormInstance>()
const registerFormData = ref({ displayName: '', username: '', password: '', passwordConfirm: '' })
const registerFormRules = {
  username: [
    { required: true, message: t('username-required'), trigger: 'blur' },
    { min: 4, max: 20, message: t('username-length'), trigger: 'blur' },
    { pattern: validUsernameRegex, message: t('username-invalid'), trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (existingUsername.value === value) {
          callback(new Error(t('username-not-unique')))
        } else {
          callback()
        }
      },
      trigger: 'change',
    },
  ],
  displayName: [{ required: true, message: t('displayname-required'), trigger: 'blur' }],
  password: [{ required: true, message: t('password-required'), trigger: 'blur' }],
  passwordConfirm: [
    { required: true, message: t('password-repeat-required'), trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== registerFormData.value.password) {
          callback(new Error(t('password-repeat-not-equal')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
} satisfies FormRules

useHead(() => ({ title: t('register') }))

const auth = useAuthStore()
const route = useRoute()

const loading = ref(false)

async function submit() {
  if (!registerFormRef.value) {
    return
  }

  const formValid = await new Promise<boolean>(resolve => registerFormRef.value?.validate(valid => resolve(valid)))
  if (!formValid) {
    return
  }

  loading.value = true
  try {
    await apiFetch<User>('/auth/users', { method: 'POST', body: registerFormData.value })
  } catch (e) {
    if (e instanceof FetchError && e.statusCode === 409) {
      existingUsername.value = registerFormData.value.username
      registerFormRef.value?.validate()
    } else {
      ElMessage({ message: String(e), type: 'error' })
    }
    loading.value = false
    return
  }

  const ok = await auth.login(registerFormData.value)
  if (ok) {
    const returnTo = typeof route.query.returnTo === 'string' ? route.query.returnTo : '/'
    await navigateTo(returnTo)
  } else {
    ElMessage({ message: t('invalid-credentials'), type: 'error' })
  }

  loading.value = false
}
</script>

<template>
  <div class="min-h-screen box-border p-2 flex flex-col">
    <AppHeader />

    <div class="m-auto text-center text-gray-500">
      <h2>{{ $t('register-account') }}</h2>

      <ElForm
        ref="registerFormRef"
        :model="registerFormData"
        :rules="registerFormRules"
        label-position="top"
        class="w-50"
        @submit.prevent
        @keypress.enter="submit"
      >
        <ElFormItem prop="displayName">
          <ElInput
            v-model="registerFormData.displayName"
            :placeholder="$t('display-name')"
            autofocus
          />
        </ElFormItem>
        <ElFormItem prop="username">
          <ElInput
            v-model="registerFormData.username"
            type="username"
            :placeholder="$t('username')"
          />
        </ElFormItem>
        <ElFormItem prop="password">
          <ElInput
            v-model="registerFormData.password"
            type="password"
            show-password
            :placeholder="$t('password')"
          />
        </ElFormItem>
        <ElFormItem prop="passwordConfirm">
          <ElInput
            v-model="registerFormData.passwordConfirm"
            type="password"
            show-password
            :placeholder="$t('password-repeat')"
          />
        </ElFormItem>
        <ElFormItem>
          <PlainButton
            type="primary"
            class="w-full"
            :label="$t('register')"
            :loading="loading"
            @click="submit"
          />
        </ElFormItem>
      </ElForm>
    </div>
  </div>
</template>
