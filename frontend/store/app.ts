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

    const appTitle = computed(() => data.value?.appTitle ?? 'PlainPage')
    const allowAdmin = computed(() => data.value?.allowAdmin || false)
    const allowRegister = computed(() => (data.value?.allowRegister ?? false) || (data.value?.setupMode ?? false))
    const version = computed(() => data.value?.version ?? '')

    return {
      appTitle,
      allowAdmin,
      allowRegister,
      version,
      refresh,
    }
  },
  {
    persist: true,
  },
)
