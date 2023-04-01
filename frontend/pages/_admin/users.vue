<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import type { PatchOperation, User } from '~/types'

const { data, error, refresh } = await useFetch<User[]>('/_api/auth/users')

const userFormVisible = ref(false)
const userFormRef = ref<FormInstance>()
const emptyUser = {
  currentUsername: '',
  username: '',
  realName: '',
  password: '',
  passwordConfirm: '',
}
const userFormData = ref({ ...emptyUser })
const userFormRules = computed(() => ({
  username: [
    { required: true, message: 'Please enter username', trigger: 'blur' },
    { min: 3, max: 20, message: 'Length should be 3 to 20', trigger: 'blur' },
  ],
  realName: [{ required: true, message: 'Please enter real name', trigger: 'blur' }],
  password: [{ required: !userFormData.value.currentUsername, message: 'Please enter password', trigger: 'blur' }],
  passwordConfirm: [
    { required: !userFormData.value.currentUsername, message: 'Please confirm password', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== userFormData.value.password) {
          callback(new Error('Passwords don\'t match'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
} satisfies FormRules))

const onCreate = async () => {
  userFormData.value = { ...emptyUser }
  userFormRef.value?.clearValidate()
  userFormVisible.value = true
}

const onEdit = async (user: User) => {
  userFormData.value = { ...user, currentUsername: user.username, password: '', passwordConfirm: '' }
  userFormRef.value?.clearValidate()
  userFormVisible.value = true
}

const onSubmit = async () => {
  if (!userFormRef.value) {
    return
  }
  await userFormRef.value.validate(async (valid, _fields) => {
    if (valid) {
      try {
        if (userFormData.value.currentUsername) {
          const ops: PatchOperation[] = [
            { op: 'replace', path: '/username', value: userFormData.value.username },
            { op: 'replace', path: '/realName', value: userFormData.value.realName },
          ]
          if (userFormData.value.password) {
            ops.push({ op: 'replace', path: '/password', value: userFormData.value.password })
          }
          await $fetch(`/_api/auth/users/${userFormData.value.currentUsername}`, {
            method: 'PATCH',
            body: ops,
          })
          ElMessage({ message: 'User updated', type: 'success' })
        } else {
          await $fetch('/_api/auth/users', { method: 'POST', body: userFormData.value })
          ElMessage({ message: 'User created', type: 'success' })
        }
        userFormVisible.value = false
        refresh()
      } catch (err) {
        ElMessage({ message: String(err), type: 'error' })
      }
    }
  })
}

const onDelete = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `Are you sure to delete user "${user.username}"?`,
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
    await $fetch(`/_api/auth/users/${user.username}`, { method: 'DELETE' })
    ElMessage({ message: 'User deleted', type: 'success' })
    refresh()
  } catch (err) {
    ElMessage({ message: String(err), type: 'error' })
  }
}
</script>

<template>
  <NetworkError v-if="!data" :msg="error?.message" @refresh="refresh" />
  <Layout v-else>
    <template #title>
      Users
    </template>

    <template #actions>
      <ElButton class="m-1" @click="onCreate">
        <Icon name="ci:user-add" /> <span class="hidden md:inline ml-1">Create user</span>
      </ElButton>
    </template>

    <ElDialog
      v-model="userFormVisible"
      :title="`${userFormData.currentUsername ? 'Edit' : 'Create'} user`"
      width="50%"
    >
      <ElForm ref="userFormRef" :model="userFormData" label-position="top" :rules="userFormRules" :validate-on-rule-change="false">
        <ElFormItem label="Username" prop="username">
          <ElInput v-model="userFormData.username" autocomplete="off" />
        </ElFormItem>
        <ElFormItem label="Real name" prop="realName">
          <ElInput v-model="userFormData.realName" autocomplete="off" />
        </ElFormItem>
        <ElFormItem label="Password" prop="password">
          <ElInput v-model="userFormData.password" show-password autocomplete="off" />
        </ElFormItem>
        <ElFormItem label="Repeat password" prop="passwordConfirm">
          <ElInput v-model="userFormData.passwordConfirm" show-password autocomplete="off" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="userFormVisible = false">Cancel</ElButton>
          <ElButton type="primary" @click="onSubmit">
            {{ userFormData.currentUsername ? 'Save' : 'Create' }}
          </ElButton>
        </span>
      </template>
    </ElDialog>

    <ElTable :data="data">
      <ElTableColumn label="Username" prop="username" />
      <ElTableColumn label="Real Name" prop="realName" />
      <ElTableColumn>
        <template #default="{ row }">
          <ElButton text>
            <Icon name="ci:edit" @click="onEdit(row)" />
          </ElButton>
          <ElButton text @click="onDelete(row)">
            <Icon class="text-red" name="ci:trash-full" />
          </ElButton>
        </template>
      </ElTableColumn>
    </ElTable>
  </Layout>
</template>