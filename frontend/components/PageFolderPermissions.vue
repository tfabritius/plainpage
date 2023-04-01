<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core'
import { AccessOp } from '~/types/'
import type { AccessRule, Breadcrumb, PageMeta, User } from '~/types/'

const props = defineProps<{
  urlPath: string
  meta: PageMeta
  title: string | undefined
  breadcrumbs: Breadcrumb[]
  isFolder: boolean
}>()

const emit = defineEmits<{ (e: 'refresh'): void }>()

const urlPath = computed(() => props.urlPath)
const meta = computed(() => props.meta)
const title = computed(() => props.title)
const isFolder = computed(() => props.isFolder)

const customPermissions = ref(!!meta.value.acls)

function mapAPI2Table(acls: AccessRule[]) {
  const ret = acls.map(acl => ({
    subject: acl.subject,
    user: acl.user,
    read: acl.ops?.includes(AccessOp.read) ?? false,
    write: acl.ops?.includes(AccessOp.write) ?? false,
    delete: acl.ops?.includes(AccessOp.delete) ?? false,
  })) ?? []

  ret.push({ subject: 'admin', user: undefined, read: true, write: true, delete: true })

  if (!ret.some(acl => acl.subject === 'all')) {
    ret.push({ subject: 'all', user: undefined, read: false, write: false, delete: false })
  }

  if (!ret.some(acl => acl.subject === 'anonymous')) {
    ret.push({ subject: 'anonymous', user: undefined, read: false, write: false, delete: false })
  }

  // Sort the list so that the first elements are always these:
  const sortOrder: Record<string, number> = {
    admin: 1,
    all: 2,
    anonymous: 3,
  }
  const sorted = ret.sort((a, b) => {
    const aOrder = sortOrder[a.subject] || Number.MAX_VALUE
    const bOrder = sortOrder[b.subject] || Number.MAX_VALUE

    // If both items have the same order and it's MAX_VALUE, sort alphabetically by name
    if (aOrder === bOrder && aOrder === Number.MAX_VALUE) {
      return a.subject.localeCompare(b.subject)
    }

    return aOrder - bOrder
  })

  return sorted
}

function mapTable2API(acls: ReturnType<typeof mapAPI2Table>): AccessRule[] {
  return acls
    .map((acl) => {
      const ops = []
      if (acl.read) {
        ops.push(AccessOp.read)
      }
      if (acl.write) {
        ops.push(AccessOp.write)
      }
      if (acl.delete) {
        ops.push(AccessOp.delete)
      }
      return {
        subject: acl.subject,
        ops,
      }
    })
    .filter(acl => acl.ops.length > 0)
    .filter(acl => acl.subject !== 'admin')
}

const editableACLs = ref(mapAPI2Table(meta.value.acls ?? []))

const newUserName = ref('')
const newUser = ref<User | null>()

const checkNewUser = useDebounceFn(async () => {
  newUserName.value = newUserName.value.trim()

  if (!newUserName.value || newUserName.value === '') {
    newUser.value = undefined
    return
  }

  try {
    const user = await $fetch<User>(`/_api/auth/users/${newUserName.value}`)

    // Check if user is in list already
    if (editableACLs.value.some(acl => acl.subject === `user:${user.id}`)) {
      newUser.value = null
    } else {
      newUser.value = user
    }
  } catch (e) {
    newUser.value = null
  }
}, 300)

watch(newUserName, () => {
  newUser.value = undefined
  checkNewUser()
})

const onDelete = (subject: string) => {
  editableACLs.value = editableACLs.value?.filter(acl => acl.subject !== subject)
}

const onAdd = () => {
  if (!newUser.value) {
    return
  }
  editableACLs.value?.push({
    subject: `user:${newUser.value.id}`,
    user: newUser.value,
    read: true,
    write: false,
    delete: false,
  })
  newUserName.value = ''
}

const onGoBack = async () => {
  await navigateTo({ query: { } })
}

const onSave = async () => {
  const apiData = customPermissions.value ? mapTable2API(editableACLs.value) : null

  await $fetch(`/_api/pages${urlPath.value}`, { method: 'PATCH', body: [{ op: 'replace', path: isFolder.value ? '/folder/meta/acls' : '/page/meta/acls', value: apiData }] })

  emit('refresh')
  onGoBack()
}
</script>

<template>
  <Layout :breadcrumbs="breadcrumbs">
    <template #title>
      <Icon v-if="isFolder" name="ci:folder" class="mr-1" />
      <span v-if="title">{{ title }}</span>
      <span v-else class="italic">Untitled</span>
    </template>

    <template #actions>
      <ElButton class="m-1" @click="onGoBack">
        <Icon name="ci:skip-back" /> <span class="hidden md:inline ml-1">Back to page</span>
      </ElButton>
      <span />
      <ElButton class="m-1" type="success" @click="onSave">
        <Icon name="ci:save" /> <span class="hidden md:inline ml-1">Save</span>
      </ElButton>
    </template>

    <ElSwitch
      v-model="customPermissions"
      active-text="Define custom permissions"
      inactive-text="Inherit permissions from parent folder"
    />
    <div v-if="customPermissions">
      <ElTable :data="editableACLs">
        <ElTableColumn label="Subject">
          <template #default="{ row }">
            <span v-if="row.subject === 'all'" class="italic">All registered users</span>
            <span v-else-if="row.subject === 'anonymous'" class="italic">Anonymous users</span>
            <span v-else-if="row.subject === 'admin'" class="italic">Administrators</span>
            <span v-else>{{ row.user ? row.user.realName : row.subject }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="Read">
          <template #default="{ row }">
            <ElCheckbox v-model="row.read" :disabled="row.subject === 'admin'" />
          </template>
        </ElTableColumn>
        <ElTableColumn label="Write">
          <template #default="{ row }">
            <ElCheckbox v-model="row.write" :disabled="row.subject === 'admin'" />
          </template>
        </ElTableColumn>
        <ElTableColumn label="Delete">
          <template #default="{ row }">
            <ElCheckbox v-model="row.delete" :disabled="row.subject === 'admin'" />
          </template>
        </ElTableColumn>
        <ElTableColumn>
          <template #default="{ row }">
            <ElButton v-if="row.subject !== 'all' && row.subject !== 'anonymous' && row.subject !== 'admin'" text @click="onDelete(row.subject)">
              <Icon class="text-red" name="ci:trash-full" />
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="flex mt-2">
        <ElInput v-model="newUserName" class="max-w-50" placeholder="Enter username">
          <template #suffix>
            <Icon v-if="newUser" class="text-green" name="ci:circle-check" />
            <Icon v-else-if="newUser === null" class="text-red" name="ci:close-circle" />
          </template>
        </ElInput>
        <ElButton :disabled="!newUser" class="ml-2" @click="onAdd">
          Add
        </ElButton>
      </div>
    </div>
  </Layout>
</template>
