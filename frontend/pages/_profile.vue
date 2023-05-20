<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '~/store/auth'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()

useHead({ title: t('profile') })

const auth = useAuthStore()

const profileFormRef = ref<FormInstance>()
const profileFormData = ref({
  displayName: auth.user?.displayName || '',
  password: '',
  passwordConfirm: '',
})
const profileFormRules = {
  displayName: [{ required: true, message: t('displayname-required'), trigger: 'blur' }],
  passwordConfirm: [
    {
      validator: (rule, value, callback) => {
        if (value !== profileFormData.value.password) {
          callback(new Error(t('password-repeat-not-equal')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
} satisfies FormRules

async function onSave() {
  if (!profileFormRef.value) {
    return
  }
  await profileFormRef.value.validate(async (valid, _fields) => {
    if (valid) {
      try {
        await auth.updateMe(profileFormData.value)
        profileFormData.value.password = ''
        profileFormData.value.passwordConfirm = ''
        ElMessage({ message: t('saved'), type: 'success' })
      } catch (err) {
        ElMessage({ message: String(err), type: 'error' })
      }
    }
  })
}

async function onDelete() {
  try {
    await ElMessageBox.confirm(
      t('are-you-sure-to-delete-this-account'),
      {
        confirmButtonText: t('delete'),
        cancelButtonText: t('cancel'),
        type: 'warning',
      },
    )
  } catch {
    return
  }

  try {
    await auth.deleteMe()
    ElMessage({ message: t('account-deleted'), type: 'success' })
  } catch (err) {
    ElMessage({ message: String(err), type: 'error' })
  }
}
</script>

<template>
  <Layout>
    <template #title>
      {{ $t('profile') }}
    </template>

    <template #actions>
      <ElButton class="m-1" type="success" @click="onSave">
        <Icon name="ci:save" /> <span class="hidden md:inline ml-1">{{ $t('save') }}</span>
      </ElButton>
    </template>
    <ElForm ref="profileFormRef" :model="profileFormData" label-position="top" :rules="profileFormRules">
      <ElFormItem :label="$t('username')">
        <ElInput :value="auth.user?.username" :disabled="true" />
      </ElFormItem>
      <ElFormItem :label="$t('display-name')" prop="displayName">
        <ElInput v-model="profileFormData.displayName" autocomplete="off" />
      </ElFormItem>
      <ElFormItem :label="$t('new-password')" prop="password">
        <ElInput v-model="profileFormData.password" show-password autocomplete="off" />
      </ElFormItem>
      <ElFormItem :label="$t('password-repeat')" prop="passwordConfirm">
        <ElInput v-model="profileFormData.passwordConfirm" show-password autocomplete="off" />
      </ElFormItem>
    </ElForm>

    <ElButton type="danger" @click="onDelete">
      {{ $t('delete-my-account') }}
    </ElButton>
  </Layout>
</template>
