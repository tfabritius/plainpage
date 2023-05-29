<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import type { PatchOperation, User } from '~/types'
import { validUsernameRegex } from '~/types'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()

useHead({ title: t('users') })

const { data, error, refresh } = await useAsyncData('/auth/users', () => apiFetch<User[]>('/auth/users'))

const userFormVisible = ref(false)
const userFormRef = ref<FormInstance>()
const emptyUser = {
  currentUsername: '',
  username: '',
  displayName: '',
  password: '',
  passwordConfirm: '',
}
const userFormData = ref({ ...emptyUser })
const userFormRules = computed(() => ({
  username: [
    { required: true, message: t('username-required'), trigger: 'blur' },
    { min: 4, max: 20, message: t('username-length'), trigger: 'blur' },
    { pattern: validUsernameRegex, message: t('username-invalid'), trigger: 'blur' },
  ],
  displayName: [{ required: true, message: t('displayname-required'), trigger: 'blur' }],
  password: [{ required: !userFormData.value.currentUsername, message: t('password-required'), trigger: 'blur' }],
  passwordConfirm: [
    { required: !userFormData.value.currentUsername, message: t('password-repeat-required'), trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== userFormData.value.password) {
          callback(new Error(t('password-repeat-not-equal')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
} satisfies FormRules))

async function onCreate() {
  userFormData.value = { ...emptyUser }
  userFormRef.value?.clearValidate()
  userFormVisible.value = true
}

async function onEdit(user: User) {
  userFormData.value = { ...user, currentUsername: user.username, password: '', passwordConfirm: '' }
  userFormRef.value?.clearValidate()
  userFormVisible.value = true
}

async function onSubmit() {
  if (!userFormRef.value) {
    return
  }

  const formValid = await new Promise<boolean>(resolve => userFormRef.value?.validate(valid => resolve(valid)))
  if (!formValid) {
    return
  }

  try {
    if (userFormData.value.currentUsername) {
      const ops: PatchOperation[] = [
        { op: 'replace', path: '/username', value: userFormData.value.username },
        { op: 'replace', path: '/displayName', value: userFormData.value.displayName },
      ]
      if (userFormData.value.password) {
        ops.push({ op: 'replace', path: '/password', value: userFormData.value.password })
      }
      await apiFetch(`/auth/users/${userFormData.value.currentUsername}`, {
        method: 'PATCH',
        body: ops,
      })
      ElMessage({ message: t('user-updated'), type: 'success' })
    } else {
      await apiFetch('/auth/users', { method: 'POST', body: userFormData.value })
      ElMessage({ message: t('user-created'), type: 'success' })
    }
    userFormVisible.value = false
    refresh()
  } catch (err) {
    ElMessage({ message: String(err), type: 'error' })
  }
}

async function onDelete(user: User) {
  try {
    await ElMessageBox.confirm(
      t('are-you-sure-to-delete-user', [user.username]),
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
    await apiFetch(`/auth/users/${user.username}`, { method: 'DELETE' })
    ElMessage({ message: t('user-deleted'), type: 'success' })
    refresh()
  } catch (err) {
    ElMessage({ message: String(err), type: 'error' })
  }
}
</script>

<template>
  <SubpageNetworkError
    v-if="!data"
    :msg="error?.message"
    :on-reload="refresh"
  />
  <Layout v-else>
    <template #title>
      {{ $t('users') }}
    </template>

    <template #actions>
      <PlainButton icon="ci:user-add" :label="$t('create-user')" @click="onCreate" />
    </template>

    <ElDialog
      v-model="userFormVisible"
      :title="userFormData.currentUsername ? $t('edit-user') : $t('create-user')"
      width="50%"
    >
      <ElForm
        ref="userFormRef"
        :model="userFormData"
        label-position="top"
        :rules="userFormRules"
        :validate-on-rule-change="false"
      >
        <ElFormItem :label="$t('username')" prop="username">
          <ElInput v-model="userFormData.username" autocomplete="off" />
        </ElFormItem>
        <ElFormItem :label="$t('display-name')" prop="displayName">
          <ElInput v-model="userFormData.displayName" autocomplete="off" />
        </ElFormItem>
        <ElFormItem :label="$t('password')" prop="password">
          <ElInput v-model="userFormData.password" show-password autocomplete="off" />
        </ElFormItem>
        <ElFormItem :label="$t('password-repeat')" prop="passwordConfirm">
          <ElInput v-model="userFormData.passwordConfirm" show-password autocomplete="off" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <PlainButton :label="$t('cancel')" @click="userFormVisible = false" />
          <PlainButton v-if="userFormData.currentUsername" type="primary" :label="$t('save')" @click="onSubmit" />
          <PlainButton v-else type="primary" :label="$t('create')" @click="onSubmit" />
        </span>
      </template>
    </ElDialog>

    <ElTable :data="data">
      <ElTableColumn :label="$t('username')" prop="username" />
      <ElTableColumn :label="$t('display-name')" prop="displayName" />
      <ElTableColumn>
        <template #default="{ row }">
          <PlainButton text icon="ci:edit" @click="onEdit(row)" />
          <PlainButton text icon="ci:trash-full" type="danger" @click="onDelete(row)" />
        </template>
      </ElTableColumn>
    </ElTable>
  </Layout>
</template>
