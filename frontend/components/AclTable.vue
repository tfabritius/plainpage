<script setup lang="ts">
import { AccessOp } from '~/types'
import type { AccessRule, User } from '~/types'

const props = defineProps<{
  acl: AccessRule[]
  showAdminRule?: boolean
  showColumns: Ops[]
}>()

function enumKeys<O extends object, K extends keyof O = keyof O>(obj: O): K[] {
  return Object.keys(obj).filter(k => Number.isNaN(+k)) as K[]
}

type Ops = keyof typeof AccessOp

type TableRow = {
  [k in Ops]: boolean;
} & {
  subject: string
  user?: User
}

function mapAPI2Table(acl: AccessRule[]): TableRow[] {
  const table = acl.map((rule) => {
    const row = {
      subject: rule.subject,
      user: rule.user,
    } as TableRow

    for (const op of enumKeys(AccessOp)) {
      row[op] = rule.ops?.includes(AccessOp[op]) ?? false
    }

    return row
  })

  if (props.showAdminRule) {
    table.push({
      subject: 'admin',
      user: undefined,
      read: true,
      write: true,
      delete: true,
      register: true,
      admin: true,
    })
  }

  if (!table.some(acl => acl.subject === 'all')) {
    table.push({
      subject: 'all',
      user: undefined,
      read: false,
      write: false,
      delete: false,
      register: false,
      admin: false,
    })
  }

  if (!table.some(acl => acl.subject === 'anonymous')) {
    table.push({
      subject: 'anonymous',
      user: undefined,
      read: false,
      write: false,
      delete: false,
      register: false,
      admin: false,
    })
  }

  // Sort the list so that the first elements are always these:
  const sortOrder: Record<string, number> = {
    admin: 1,
    all: 2,
    anonymous: 3,
  }
  const sorted = table.sort((a, b) => {
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

function mapTable2API(table: ReturnType<typeof mapAPI2Table>): AccessRule[] {
  return table
    .map((row) => {
      const ops = []

      for (const op of enumKeys(AccessOp)) {
        if (row[op]) {
          ops.push(AccessOp[op])
        }
      }

      return {
        subject: row.subject,
        ops,
      }
    })
    .filter(acl => acl.ops.length > 0)
    .filter(acl => acl.subject !== 'admin')
}

const editableACL = ref(mapAPI2Table(props.acl))

const newUserName = ref('')
const newUser = ref<User | null>()

const checkNewUser = useDebounceFn(async () => {
  newUserName.value = newUserName.value.trim()

  if (!newUserName.value || newUserName.value === '') {
    newUser.value = undefined
    return
  }

  try {
    const user = await apiFetch<User>(`/auth/users/${newUserName.value}`)

    // Check if user is in list already
    if (editableACL.value.some(acl => acl.subject === `user:${user.id}`)) {
      newUser.value = null
    } else {
      newUser.value = user
    }
  } catch {
    newUser.value = null
  }
}, 300)

watch(newUserName, () => {
  newUser.value = undefined
  checkNewUser()
})

function onRemoveRule(subject: string) {
  editableACL.value = editableACL.value?.filter(rule => rule.subject !== subject)
}

function onAddRule() {
  if (!newUser.value) {
    return
  }
  editableACL.value?.push({
    subject: `user:${newUser.value.id}`,
    user: newUser.value,
    read: props.showColumns.includes('read'),
    write: false,
    delete: false,
    register: false,
    admin: false,
  })
  newUserName.value = ''
}

function getAcl() {
  return mapTable2API(editableACL.value)
}

defineExpose({ getAcl })
</script>

<template>
  <ElTable :data="editableACL">
    <ElTableColumn :label="$t('acl.subject')">
      <template #default="{ row }">
        <span
          v-if="row.subject === 'all'"
          class="italic"
        >
          {{ $t('all-registered-users') }}
        </span>
        <span
          v-else-if="row.subject === 'anonymous'"
          class="italic"
        >
          {{ $t('anonymous-users') }}
        </span>
        <span
          v-else-if="row.subject === 'admin'"
          class="italic"
        >
          {{ $t('administrators') }}
        </span>
        <span v-else>
          <template v-if="row.user">{{ row.user.username }} ({{ row.user.displayName }})</template>
          <template v-else><samp>{{ row.subject }}</samp></template>
        </span>
      </template>
    </ElTableColumn>
    <ElTableColumn
      v-if="props.showColumns.includes('read')"
      :label="$t('read')"
    >
      <template #default="{ row }">
        <ElCheckbox v-model="row.read" :disabled="row.subject === 'admin'" />
      </template>
    </ElTableColumn>
    <ElTableColumn
      v-if="props.showColumns.includes('write')"
      :label="$t('write')"
    >
      <template #default="{ row }">
        <ElCheckbox v-model="row.write" :disabled="row.subject === 'admin'" />
      </template>
    </ElTableColumn>
    <ElTableColumn
      v-if="props.showColumns.includes('delete')"
      :label="$t('delete')"
    >
      <template #default="{ row }">
        <ElCheckbox v-model="row.delete" :disabled="row.subject === 'admin'" />
      </template>
    </ElTableColumn>
    <ElTableColumn
      v-if="props.showColumns.includes('register')"
      :label="$t('register')"
    >
      <template #default="{ row }">
        <ElCheckbox v-model="row.register" :disabled="row.subject === 'admin'" />
      </template>
    </ElTableColumn>
    <ElTableColumn
      v-if="props.showColumns.includes('admin')"
      :label="$t('acl.admin')"
    >
      <template #default="{ row }">
        <ElCheckbox
          v-model="row.admin"
          :disabled="['anonymous', 'all', 'admin'].includes(row.subject)"
        />
      </template>
    </ElTableColumn>
    <ElTableColumn>
      <template #default="{ row }">
        <PlainButton
          v-if="!['anonymous', 'all', 'admin'].includes(row.subject)"
          text
          type="danger"
          icon="ci:trash-full"
          @click="onRemoveRule(row.subject)"
        />
      </template>
    </ElTableColumn>
  </ElTable>

  <div class="flex mt-2">
    <ElInput v-model="newUserName" class="max-w-50" :placeholder="$t('username')">
      <template #suffix>
        <Icon v-if="newUser" class="text-green" name="ci:circle-check" />
        <Icon v-else-if="newUser === null" class="text-red" name="ci:close-circle" />
      </template>
    </ElInput>
    <PlainButton :disabled="!newUser" :label="$t('add')" class="ml-2" @click="onAddRule" />
  </div>
</template>
