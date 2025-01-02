<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import type { AccessRule, User } from '~/types'
import { AccessOp } from '~/types'

const props = defineProps<{
  acl: AccessRule[]
  showAdminRule?: boolean
  showColumns: Ops[]
}>()

const { t } = useI18n()

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

const columns: TableColumn<TableRow>[] = [
  { header: t('acl.subject'), id: 'subject', accessorKey: 'subject' },
  { header: t('read'), accessorKey: 'read' },
  { header: t('write'), accessorKey: 'write' },
  { header: t('delete'), accessorKey: 'delete' },
  { header: t('register'), accessorKey: 'register' },
  { header: t('acl.admin'), accessorKey: 'admin' },
  { header: 'TODO', id: 'actions' },
]

const columnVisibility = ref({
  read: props.showColumns.includes('read'),
  write: props.showColumns.includes('write'),
  delete: props.showColumns.includes('delete'),
  register: props.showColumns.includes('register'),
  admin: props.showColumns.includes('admin'),
})

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
  editableACL.value = [...editableACL.value, {
    subject: `user:${newUser.value.id}`,
    user: newUser.value,
    read: props.showColumns.includes('read'),
    write: false,
    delete: false,
    register: false,
    admin: false,
  }]
  newUserName.value = ''
}

function getAcl() {
  return mapTable2API(editableACL.value)
}

defineExpose({ getAcl })
</script>

<template>
  <UTable
    v-model:column-visibility="columnVisibility"
    :columns="columns"
    :data="editableACL"
  >
    <template #subject-cell="{ row }">
      <span
        v-if="row.original.subject === 'all'"
        class="italic"
      >
        {{ t('all-registered-users') }}
      </span>
      <span
        v-else-if="row.original.subject === 'anonymous'"
        class="italic"
      >
        {{ $t('anonymous-users') }}
      </span>
      <span
        v-else-if="row.original.subject === 'admin'"
        class="italic"
      >
        {{ $t('administrators') }}
      </span>
      <span v-else>
        <template v-if="row.original.user">{{ row.original.user.username }} ({{ row.original.user.displayName }})</template>
        <template v-else><samp>{{ row.original.subject }}</samp></template>
      </span>
    </template>
    <template #read-cell="{ row }">
      <UCheckbox v-model="row.original.read" :disabled="row.original.subject === 'admin'" />
    </template>
    <template #write-cell="{ row }">
      <UCheckbox v-model="row.original.write" :disabled="row.original.subject === 'admin'" />
    </template>
    <template #delete-cell="{ row }">
      <UCheckbox v-model="row.original.delete" :disabled="row.original.subject === 'admin'" />
    </template>
    <template #register-cell="{ row }">
      <UCheckbox v-model="row.original.register" :disabled="row.original.subject === 'admin'" />
    </template>
    <template #admin-cell="{ row }">
      <UCheckbox v-model="row.original.admin" :disabled="['anonymous', 'all', 'admin'].includes(row.original.subject)" />
    </template>
    <template #actions-cell="{ row }">
      <UButton
        v-if="!['anonymous', 'all', 'admin'].includes(row.original.subject)"
        variant="link"
        color="error"
        icon="ci:trash-full"
        @click="onRemoveRule(row.original.subject)"
      />
    </template>
  </UTable>

  <div class="flex mt-2">
    <UInput v-model="newUserName" class="max-w-50" :placeholder="$t('username')">
      <template #trailing>
        <UIcon v-if="newUser" class="text-green-500" name="ci:circle-check" />
        <UIcon v-else-if="newUser === null" class="text-red-500" name="ci:close-circle" />
      </template>
    </UInput>
    <UButton :disabled="!newUser" :label="$t('add')" class="ml-2" @click="onAddRule" />
  </div>
</template>
