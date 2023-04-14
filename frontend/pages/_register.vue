<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { FetchError } from 'ofetch'
import { useAuthStore } from '~/store/auth'
import type { User } from '~/types'
import { validUsernameRegex } from '~/types'

const existingUsername = ref<string>()

const registerFormRef = ref<FormInstance>()
const registerFormData = ref({ displayName: '', username: '', password: '', passwordConfirm: '' })
const registerFormRules = {
  username: [
    { required: true, message: 'Please enter username', trigger: 'blur' },
    { min: 4, max: 20, message: 'Length should be 4 to 20', trigger: 'blur' },
    { pattern: validUsernameRegex, message: 'Invalid username', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (existingUsername.value === value) {
          callback(new Error('Username is taken already.'))
        } else {
          callback()
        }
      },
      trigger: 'change',
    },
  ],
  displayName: [{ required: true, message: 'Please enter display name', trigger: 'blur' }],
  password: [{ required: true, message: 'Please enter password', trigger: 'blur' }],
  passwordConfirm: [
    { required: true, message: 'Please confirm password', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== registerFormData.value.password) {
          callback(new Error('Passwords don\'t match'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
} satisfies FormRules

useHead({ title: 'Register' })

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
  }

  const ok = await auth.login(registerFormData.value)
  if (ok) {
    const returnTo = typeof route.query.returnTo === 'string' ? route.query.returnTo : '/'
    await navigateTo(returnTo)
  } else {
    ElMessage({ message: 'Could not sign in', type: 'error' })
  }

  loading.value = false
}
</script>

<template>
  <div class="absolute inset-0 bg-white flex">
    <div class="m-auto text-center text-gray-500">
      <h2>Register account</h2>

      <ElForm ref="registerFormRef" :model="registerFormData" :rules="registerFormRules" label-position="top" class="w-50" @submit.prevent @keypress.enter="submit">
        <ElFormItem prop="displayName">
          <ElInput v-model="registerFormData.displayName" placeholder="Display name" autofocus />
        </ElFormItem>
        <ElFormItem prop="username">
          <ElInput v-model="registerFormData.username" type="username" placeholder="Username" autofocus />
        </ElFormItem>
        <ElFormItem prop="password">
          <ElInput
            v-model="registerFormData.password" type="password" show-password placeholder="Password"
          />
        </ElFormItem>
        <ElFormItem prop="passwordConfirm">
          <ElInput
            v-model="registerFormData.passwordConfirm" type="password" show-password placeholder="Repeat password"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" class="w-full" :loading="loading" @click="submit">
            Register
          </ElButton>
        </ElFormItem>
      </ElForm>

      <NuxtLink v-slot="{ navigate, href }" custom :to="`_login?returnTo=${encodeURIComponent(String(route.query.returnTo || '/'))}`">
        <ElLink :underline="false" :href="href" @click="navigate">
          <Icon name="ic:round-log-in" class="mr-1" /> Sign in
        </ElLink>
      </NuxtLink>
    </div>
  </div>
</template>
