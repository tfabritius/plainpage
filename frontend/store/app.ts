import { defineStore } from 'pinia'
import type { GetAppResponse } from '~/types'

export const useAppStore = defineStore(
  'app',
  () => {
    const data = ref<GetAppResponse>()

    async function refresh() {
      const res = await apiFetch<GetAppResponse>('/app')
      data.value = res
    }

    const appName = computed(() => data.value?.appName ?? 'PlainPage')
    const allowAdmin = computed(() => data.value?.allowAdmin || false)
    const allowRegister = computed(() => data.value?.allowRegister ?? false)

    return { appName, allowAdmin, allowRegister, refresh }
  },
  {
    persist: true,
  },
)
