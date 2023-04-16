<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '~/store/auth'

const auth = useAuthStore()

const profileFormRef = ref<FormInstance>()
const profileFormData = ref({
  displayName: auth.user?.displayName || '',
  password: '',
  passwordConfirm: '',
})
const profileFormRules = {
  displayName: [{ required: true, message: 'Please enter display name', trigger: 'blur' }],
  passwordConfirm: [
    {
      validator: (rule, value, callback) => {
        if (value !== profileFormData.value.password) {
          callback(new Error('Passwords don\'t match'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
} satisfies FormRules

const onSave = async () => {
  if (!profileFormRef.value) {
    return
  }
  await profileFormRef.value.validate(async (valid, _fields) => {
    if (valid) {
      try {
        await auth.updateMe(profileFormData.value)
        profileFormData.value.password = ''
        profileFormData.value.passwordConfirm = ''
        ElMessage({ message: 'Saved', type: 'success' })
      } catch (err) {
        ElMessage({ message: String(err), type: 'error' })
      }
    }
  })
}

const onDelete = async () => {
  try {
    await ElMessageBox.confirm(
      'Are you sure to delete this account?',
      {
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        type: 'warning',
      },
    )
  } catch {
    return
  }

  try {
    await auth.deleteMe()
    ElMessage({ message: 'Deleted', type: 'success' })
  } catch (err) {
    ElMessage({ message: String(err), type: 'error' })
  }
}
</script>

<template>
  <Layout>
    <template #title>
      Profile
    </template>

    <template #actions>
      <ElButton class="m-1" type="success" @click="onSave">
        <Icon name="ci:save" /> <span class="hidden md:inline ml-1">Save</span>
      </ElButton>
    </template>
    <ElForm ref="profileFormRef" :model="profileFormData" label-position="top" :rules="profileFormRules">
      <ElFormItem label="Username">
        <ElInput :value="auth.user?.username" :disabled="true" />
      </ElFormItem>
      <ElFormItem label="Display name" prop="displayName">
        <ElInput v-model="profileFormData.displayName" autocomplete="off" />
      </ElFormItem>
      <ElFormItem label="New password" prop="password">
        <ElInput v-model="profileFormData.password" show-password autocomplete="off" />
      </ElFormItem>
      <ElFormItem label="Repeat password" prop="passwordConfirm">
        <ElInput v-model="profileFormData.passwordConfirm" show-password autocomplete="off" />
      </ElFormItem>
    </ElForm>

    <ElButton type="danger" @click="onDelete">
      Delete my account
    </ElButton>
  </Layout>
</template>
