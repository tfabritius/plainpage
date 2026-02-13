<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import type { GetTrashListResponse, GetTrashPageResponse, TrashActionRequest, TrashEntry, TrashItemRef } from '~/types'
import { UseTimeAgo } from '@vueuse/components'
import { useRouteQuery } from '@vueuse/router'
import { format } from 'date-fns'
import { FetchError } from 'ofetch'
import { timeAgoMessages } from '~/composables/timeAgoMessages'

definePageMeta({
  middleware: ['require-auth'],
})

const { t } = useI18n()
const toast = useToast()

// Query params for page view mode
const viewUrl = useRouteQuery<string | undefined>('url', undefined, {
  transform: val => Array.isArray(val) ? val[0] : val,
})
const viewDeletedAt = useRouteQuery('deletedAt', 0, {
  transform: (val) => {
    const str = Array.isArray(val) ? val[0] : val
    return str ? Number.parseInt(str, 10) : 0
  },
})

// Check if we're viewing a single page or the list
const isPageView = computed(() => !!viewUrl.value && viewDeletedAt.value > 0)

// ==================== LIST VIEW STATE ====================
const listPage = ref(1)
const limit = 20

// Sort state
const sortBy = ref<'deletedAt' | 'url'>('deletedAt')
const sortOrder = ref<'asc' | 'desc'>('desc')

const { data: listData, error: listError, refresh: listRefresh, status: listStatus } = await useAsyncData(
  'trash-list',
  async () => {
    if (isPageView.value) {
      return null
    }
    const response = await apiFetch<GetTrashListResponse>(
      `/trash/?page=${listPage.value}&limit=${limit}&sortBy=${sortBy.value}&sortOrder=${sortOrder.value}`,
    )
    return response
  },
  {
    watch: [listPage, sortBy, sortOrder, isPageView],
  },
)

const selectedItems = ref<Set<string>>(new Set())

function getItemKey(item: TrashEntry): string {
  return `${item.url}:${item.deletedAt}`
}

function isSelected(item: TrashEntry): boolean {
  return selectedItems.value.has(getItemKey(item))
}

function toggleSelection(item: TrashEntry) {
  const key = getItemKey(item)
  if (selectedItems.value.has(key)) {
    selectedItems.value.delete(key)
  } else {
    selectedItems.value.add(key)
  }
}

function toggleSelectAll(value: boolean | 'indeterminate') {
  if (!listData.value?.items) {
    return
  }

  if (value) {
    // Select all current items
    for (const item of listData.value.items) {
      selectedItems.value.add(getItemKey(item))
    }
  } else {
    // Deselect all current items
    for (const item of listData.value.items) {
      selectedItems.value.delete(getItemKey(item))
    }
  }
}

const selectAllState = computed((): boolean | 'indeterminate' => {
  if (!listData.value?.items || listData.value.items.length === 0) {
    return false
  }
  const allCurrentlySelected = listData.value.items.every(item => selectedItems.value.has(getItemKey(item)))
  if (allCurrentlySelected) {
    return true
  }
  const someCurrentlySelected = listData.value.items.some(item => selectedItems.value.has(getItemKey(item)))
  if (someCurrentlySelected) {
    return 'indeterminate'
  }
  return false
})

function getSelectedItemRefs(): TrashItemRef[] {
  const refs: TrashItemRef[] = []
  for (const key of selectedItems.value) {
    const parts = key.split(':')
    const deletedAtStr = parts.pop() ?? '0'
    const url = parts.join(':')
    refs.push({ url, deletedAt: Number.parseInt(deletedAtStr, 10) })
  }
  return refs
}

const plainDialog = useTemplateRef('plainDialog')

async function onDeleteSelected() {
  if (selectedItems.value.size === 0) {
    return
  }

  const confirmed = await plainDialog.value?.confirm(
    t('confirm-delete-items-permanent', { count: selectedItems.value.size }),
    {
      title: t('delete-permanently'),
      confirmButtonText: t('delete'),
      confirmButtonColor: 'warning',
    },
  )

  if (!confirmed) {
    return
  }

  try {
    await apiFetch('/trash/delete', {
      method: 'POST',
      body: { items: getSelectedItemRefs() },
    })

    toast.add({
      description: t('items-deleted-permanent', { count: selectedItems.value.size }),
      color: 'success',
    })

    selectedItems.value.clear()
    await listRefresh()
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

async function onRestoreSelected() {
  if (selectedItems.value.size === 0) {
    return
  }

  try {
    await apiFetch('/trash/restore', {
      method: 'POST',
      body: { items: getSelectedItemRefs() },
    })

    toast.add({
      description: t('items-restored', { count: selectedItems.value.size }),
      color: 'success',
    })

    selectedItems.value.clear()
    await listRefresh()
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

const totalPages = computed(() => {
  if (!listData.value) {
    return 1
  }
  return Math.ceil(listData.value.totalCount / limit)
})

// Reset selection when page changes
watch(listPage, () => {
  selectedItems.value.clear()
})

// Reset to first page when sort changes
watch([sortBy, sortOrder], () => {
  listPage.value = 1
  selectedItems.value.clear()
})

function toggleSort(column: 'url' | 'deletedAt') {
  if (sortBy.value === column) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortOrder.value = 'desc'
  }
}

function getSortIcon(column: 'url' | 'deletedAt'): string {
  if (sortBy.value !== column) {
    return 'tabler:arrows-sort'
  }
  return sortOrder.value === 'asc' ? 'tabler:sort-ascending' : 'tabler:sort-descending'
}

const columns: TableColumn<TrashEntry>[] = [
  { header: '', id: 'select' },
  { header: t('title'), id: 'title' },
  { header: t('url'), id: 'url' },
  { header: t('deleted-at'), id: 'deletedAt' },
]

// ==================== PAGE VIEW STATE ====================
const { data: pageData, error: pageError, refresh: pageRefresh } = await useAsyncData(
  'trash-page',
  async () => {
    if (!isPageView.value) {
      return null
    }

    try {
      const response = await apiFetch<GetTrashPageResponse>(
        `/trash/page?url=${encodeURIComponent(viewUrl.value!)}&deletedAt=${viewDeletedAt.value}`,
      )
      return {
        notFound: false,
        page: response.page,
      }
    } catch (err) {
      if (err instanceof FetchError && err.statusCode === 404) {
        return { notFound: true, page: null, deletedAt: 0 }
      }
      throw err
    }
  },
  {
    watch: [isPageView, viewUrl, viewDeletedAt],
  },
)

const pageTitle = computed(() => {
  if (pageData.value?.page) {
    return pageData.value.page.meta.title || t('untitled')
  }
  return t('not-found')
})

const deletedAtDate = computed(() => new Date(viewDeletedAt.value * 1000))

// Dynamic page title
useHead(() => {
  if (isPageView.value) {
    return { title: `${t('trash')}: ${pageTitle.value}` }
  }
  return { title: t('trash') }
})

async function onRestorePage() {
  const confirmed = await plainDialog.value?.confirm(
    t('confirm-restore-page'),
    {
      title: t('restore'),
      confirmButtonText: t('restore'),
    },
  )

  if (!confirmed) {
    return
  }

  try {
    await apiFetch('/trash/restore', {
      method: 'POST',
      body: {
        items: [{ url: viewUrl.value, deletedAt: viewDeletedAt.value }],
      },
    })

    toast.add({
      description: t('page-restored'),
      color: 'success',
    })

    await navigateTo(`/${viewUrl.value}`)
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

async function onDeletePagePermanently() {
  const confirmed = await plainDialog.value?.confirm(
    t('confirm-delete-page-permanent'),
    {
      title: t('delete-permanently'),
      confirmButtonText: t('delete'),
      confirmButtonColor: 'warning',
    },
  )

  if (!confirmed) {
    return
  }

  try {
    const request: TrashActionRequest = {
      items: [{ url: viewUrl.value!, deletedAt: viewDeletedAt.value }],
    }
    await apiFetch('/trash/delete', {
      method: 'POST',
      body: request,
    })

    toast.add({
      description: t('page-deleted-permanent'),
      color: 'success',
    })

    await navigateTo('/_admin/trash')
  } catch (err) {
    toast.add({
      description: String(err),
      color: 'error',
    })
  }
}

function goToList() {
  navigateTo('/_admin/trash')
}

function getTrashPageUrl(item: TrashEntry): string {
  return `/_admin/trash?url=${item.url}&deletedAt=${item.deletedAt}`
}
</script>

<template>
  <!-- PAGE VIEW MODE -->
  <SubpageNetworkError
    v-if="isPageView && pageError && !pageData"
    :msg="pageError?.message"
    :on-reload="pageRefresh"
  />
  <Layout v-else-if="isPageView">
    <template #title>
      <UIcon name="tabler:trash" class="mr-2" />
      <span v-if="pageData?.page?.meta.title">{{ pageTitle }}</span>
      <span v-else class="italic">{{ pageTitle }}</span>
    </template>

    <template v-if="!pageData?.notFound" #subtitle>
      <span class="text-muted">
        /{{ viewUrl }}
      </span>
      <span class="mx-2">•</span>
      <UIcon name="tabler:trash" class="mr-1" />
      {{ $t('deleted-at') }}: {{ format(deletedAtDate, 'yyyy-MM-dd HH:mm') }}
    </template>

    <template #actions>
      <ReactiveButton
        icon="tabler:arrow-back-up"
        :label="$t('back-to-trash')"
        @click="goToList"
      />
      <ReactiveButton
        v-if="!pageData?.notFound"
        icon="tabler:restore"
        :label="$t('restore')"
        @click="onRestorePage"
      />
      <ReactiveButton
        v-if="!pageData?.notFound"
        icon="tabler:trash-x"
        :label="$t('delete-permanently')"
        color="warning"
        @click="onDeletePagePermanently"
      />
    </template>

    <div v-if="pageData?.page">
      <PageView :page="pageData.page" />
    </div>
    <div v-else>
      <div class="text-center">
        <span class="text-3xl">😟</span>
        <div class="font-medium m-4">
          {{ $t('this-page-doesnt-exist') }}
        </div>

        <ReactiveButton
          icon="tabler:arrow-back-up"
          :label="$t('back-to-trash')"
          @click="goToList"
        />
      </div>
    </div>

    <PlainDialog ref="plainDialog" />
  </Layout>

  <!-- LIST VIEW MODE -->
  <SubpageNetworkError
    v-else-if="!listData && listError"
    :msg="listError?.message"
    :on-reload="listRefresh"
  />
  <Layout v-else>
    <template #title>
      <UIcon name="tabler:trash" class="mr-2" />
      {{ $t('trash') }}
    </template>

    <template #actions>
      <ReactiveButton
        icon="tabler:refresh"
        :label="$t('reload')"
        :loading="listStatus === 'pending'"
        @click="listRefresh"
      />
      <ReactiveButton
        v-if="selectedItems.size > 0"
        icon="tabler:restore"
        :label="$t('restore')"
        @click="onRestoreSelected"
      />
      <ReactiveButton
        v-if="selectedItems.size > 0"
        icon="tabler:trash-x"
        :label="$t('delete-permanently')"
        color="warning"
        @click="onDeleteSelected"
      />
    </template>

    <div v-if="listData?.items && listData.items.length > 0">
      <UTable
        :data="listData.items"
        :columns="columns"
        class="w-full"
      >
        <template #select-header>
          <UCheckbox
            :model-value="selectAllState"
            @update:model-value="toggleSelectAll"
          />
        </template>
        <template #select-cell="{ row }">
          <UCheckbox
            :model-value="isSelected(row.original)"
            @change="toggleSelection(row.original)"
          />
        </template>
        <template #title-cell="{ row }">
          <ULink :to="getTrashPageUrl(row.original)" class="text-inherit">
            <span v-if="row.original.meta?.title">{{ row.original.meta.title }}</span>
            <span v-else class="italic">{{ $t('untitled') }}</span>
          </ULink>
        </template>
        <template #url-header>
          <button
            class="flex items-center gap-1 cursor-pointer"
            @click="toggleSort('url')"
          >
            {{ $t('url') }}
            <UIcon :name="getSortIcon('url')" class="w-4 h-4" />
          </button>
        </template>
        <template #url-cell="{ row }">
          /{{ row.original.url }}
        </template>
        <template #deletedAt-header>
          <button
            class="flex items-center gap-1 cursor-pointer"
            @click="toggleSort('deletedAt')"
          >
            {{ $t('deleted-at') }}
            <UIcon :name="getSortIcon('deletedAt')" class="w-4 h-4" />
          </button>
        </template>
        <template #deletedAt-cell="{ row }">
          <UseTimeAgo v-slot="{ timeAgo }" :time="new Date(row.original.deletedAt * 1000)" :messages="timeAgoMessages()">
            <span :title="format(new Date(row.original.deletedAt * 1000), 'yyyy-MM-dd HH:mm:ss')">
              {{ timeAgo }}
            </span>
          </UseTimeAgo>
        </template>
      </UTable>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex justify-center mt-4">
        <UPagination
          v-model:page="listPage"
          :total="listData.totalCount"
          :items-per-page="limit"
        />
      </div>
    </div>

    <div v-else-if="listData?.items && listData.items.length === 0" class="text-center py-8 text-muted">
      <UIcon name="tabler:trash" class="text-4xl mb-2" />
      <p>
        {{ $t('trash-is-empty') }}
      </p>
    </div>

    <PlainDialog ref="plainDialog" />
  </Layout>
</template>
