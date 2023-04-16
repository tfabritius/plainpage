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
    const allowRegister = computed(() => data.value?.allowRegister ?? false)

    return { appTitle, allowAdmin, allowRegister, refresh }
  },
  {
    persist: true,
  },
)
