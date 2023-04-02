<script setup lang="ts">
import { useAuthStore } from '~/store/auth'

const loginFormData = ref({ username: '', password: '' })

useHead({ title: 'Sign in' })

const auth = useAuthStore()
const route = useRoute()

const loading = ref(false)

async function submit() {
  loading.value = true
  const success = await auth.login(loginFormData.value)

  if (success) {
    const returnTo = typeof route.query.returnTo === 'string' ? route.query.returnTo : '/'
    await navigateTo(returnTo)
  } else {
    ElMessage({ message: 'Invalid credentials', type: 'error' })
  }
  loading.value = false
}
</script>

<template>
  <div class="absolute inset-0 bg-white flex">
    <div class="m-auto text-center text-gray-500">
      <Icon name="ci:user-circle" size="5em" class="mb-3" />

      <ElForm label-position="top" class="w-50" @submit.prevent @keypress.enter="submit">
        <ElFormItem>
          <ElInput v-model="loginFormData.username" type="username" placeholder="Username" autofocus />
        </ElFormItem>
        <ElFormItem>
          <ElInput
            v-model="loginFormData.password" type="password" show-password placeholder="Password"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" class="w-full" :loading="loading" @click="submit">
            Sign in
          </ElButton>
        </ElFormItem>
      </ElForm>
    </div>
  </div>
</template>