<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import type { ChangePasswordRequest, DeleteUserRequest, PatchOperation, User } from '~/types'
import { FetchError } from 'ofetch'
import { z } from 'zod'
import { validUsernameRegex } from '~/types'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()
const toast = useToast()

useHead({ title: t('users') })

const { data, error, refresh } = await useAsyncData('/auth/users', () => apiFetch<User[]>('/auth/users'))

const columns: TableColumn<User>[] = [
  { header: t('username'), accessorKey: 'username' },
  { header: t('display-name'), accessorKey: 'displayName' },
  { header: '', id: 'actions' },
]

const userFormVisible = ref(false)
const userForm = useTemplateRef('userFormRef')
const userFormSelectedUsername = ref('')
const userFormSchema = z.object({
  username: z.string()
    .min(4, t('username-length'))
    .max(20, t('username-length'))
    .regex(validUsernameRegex, t('username-invalid')),
  displayName: z.string().min(1, t('displayname-required')),
  adminCurrentPassword: z.string().refine(
    password => !userFormSelectedUsername.value || password.length > 0,
    t('current-password-required'),
  ),
  password: z.string().refine(password => userFormSelectedUsername.value || password.length > 0, t('password-required')),
  passwordConfirm: z.string(),
})
  .refine(({ password, passwordConfirm }) => password === passwordConfirm, { message: t('password-repeat-not-equal'), path: ['passwordConfirm'] })

type UserFormSchema = z.output<typeof userFormSchema>
const userFormState = reactive<UserFormSchema>(
  { displayName: '', username: '', adminCurrentPassword: '', password: '', passwordConfirm: '' },
)

async function onCreate() {
  userFormState.displayName = ''
  userFormState.username = ''
  userFormState.adminCurrentPassword = ''
  userFormState.password = ''
  userFormState.passwordConfirm = ''
  userFormSelectedUsername.value = ''
  userForm.value?.clear()
  userFormVisible.value = true
}

async function onEdit(user: User) {
  userFormState.displayName = user.displayName
  userFormState.username = user.username
  userFormState.adminCurrentPassword = ''
  userFormState.password = ''
  userFormState.passwordConfirm = ''
  userFormSelectedUsername.value = user.username
  userForm.value?.clear()
  userFormVisible.value = true
}

async function onSubmit() {
  if (!userForm.value) {
    return
  }

  try {
    if (userFormSelectedUsername.value) {
      // Update user details (username and displayName)
      const ops: PatchOperation[] = [
        { op: 'replace', path: '/username', value: userFormState.username },
        { op: 'replace', path: '/displayName', value: userFormState.displayName },
      ]
      await apiFetch(`/auth/users/${userFormSelectedUsername.value}`, {
        method: 'PATCH',
        body: ops,
      })

      // Change password if provided
      if (userFormState.password) {
        const passwordRequest: ChangePasswordRequest = {
          currentPassword: userFormState.adminCurrentPassword,
          newPassword: userFormState.password,
        }
        await apiFetch(`/auth/users/${userFormState.username}/password`, {
          method: 'POST',
          body: passwordRequest,
        })
      }

      toast.add({ description: t('user-updated'), color: 'success' })
    } else {
      await apiFetch('/auth/users', { method: 'POST', body: userFormState })
      toast.add({ description: t('user-created'), color: 'success' })
    }
    userFormVisible.value = false
    refresh()
  } catch (err) {
    toast.add({ description: String(err), color: 'error' })
  }
}

const deleteModalOpen = ref(false)
const deletePasswordInput = ref('')
const deleteTargetUser = ref<User | null>(null)

function onDeleteClick(user: User) {
  deletePasswordInput.value = ''
  deleteTargetUser.value = user
  deleteModalOpen.value = true
}

async function onDeleteConfirm() {
  if (!deleteTargetUser.value) {
    return
  }

  if (!deletePasswordInput.value) {
    toast.add({ description: t('current-password-required'), color: 'error' })
    return
  }

  const request: DeleteUserRequest = { password: deletePasswordInput.value }
  try {
    await apiFetch(`/auth/users/${deleteTargetUser.value.username}/delete`, {
      method: 'POST',
      body: request,
    })
    deleteModalOpen.value = false
    toast.add({ description: t('user-deleted'), color: 'success' })
    refresh()
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 403) {
      toast.add({ description: t('incorrect-password'), color: 'error' })
      return
    }
    toast.add({ description: String(err), color: 'error' })
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
      <ReactiveButton icon="tabler:user-plus" :label="$t('create-user')" @click="onCreate" />
    </template>

    <UModal
      v-model:open="userFormVisible"
      :title="userFormSelectedUsername ? $t('edit-user') : $t('create-user')"
    >
      <template #body>
        <UForm
          id="userForm"
          ref="userFormRef"
          :state="userFormState"
          :schema="userFormSchema"
          @submit="onSubmit"
        >
          <UFormField :label="$t('username')" name="username">
            <UInput v-model="userFormState.username" autocomplete="off" class="w-full" />
          </UFormField>
          <UFormField :label="$t('display-name')" name="displayName">
            <UInput v-model="userFormState.displayName" autocomplete="off" class="w-full" />
          </UFormField>
          <UFormField v-if="userFormSelectedUsername" :label="$t('admin-current-password')" name="adminCurrentPassword">
            <UInput v-model="userFormState.adminCurrentPassword" type="password" autocomplete="off" class="w-full" />
          </UFormField>
          <UFormField :label="$t('password')" name="password">
            <UInput v-model="userFormState.password" type="password" autocomplete="off" class="w-full" />
          </UFormField>
          <UFormField :label="$t('password-repeat')" name="passwordConfirm">
            <UInput v-model="userFormState.passwordConfirm" type="password" autocomplete="off" class="w-full" />
          </UFormField>
        </UForm>
      </template>
      <template #footer>
        <UButton :label="$t('cancel')" @click="userFormVisible = false" />
        <UButton color="primary" variant="solid" :label="userFormSelectedUsername ? $t('save') : $t('create')" type="submit" form="userForm" />
      </template>
    </UModal>

    <UTable
      :data="data" :columns="columns"
    >
      <template #actions-cell="{ row }">
        <UButton variant="link" icon="tabler:edit" @click="onEdit(row.original)" />
        <UButton variant="link" icon="tabler:trash" color="error" @click="onDeleteClick(row.original)" />
      </template>
    </UTable>

    <UModal v-model:open="deleteModalOpen" :title="$t('delete-user')">
      <template #body>
        <p class="mb-4">
          {{ $t('are-you-sure-to-delete-user', [deleteTargetUser?.username]) }}
        </p>
        <UFormField :label="$t('admin-current-password')">
          <UInput v-model="deletePasswordInput" type="password" autocomplete="off" class="w-full" />
        </UFormField>
      </template>
      <template #footer>
        <UButton :label="$t('cancel')" @click="deleteModalOpen = false" />
        <UButton color="warning" variant="solid" :label="$t('delete')" @click="onDeleteConfirm" />
      </template>
    </UModal>
  </Layout>
</template>
