import { FetchError } from 'ofetch'
import type { NitroFetchOptions, NitroFetchRequest } from 'nitropack'
import { useAuthStore } from '~/store/auth'

export async function apiFetch<T>(request: string, opts?: NitroFetchOptions<NitroFetchRequest>) {
  const { loggedIn, token, logout } = useAuthStore()
  const route = useRoute()

  const headers = { ...opts?.headers, authorization: loggedIn ? `Bearer ${token}` : '' }

  try {
    return await $fetch<T>(request, {
      ...opts,
      headers,
      baseURL: opts?.baseURL || '/_api',
    })
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 401) {
      if (loggedIn) {
        await logout()
      }

      if (route.path !== '/_login') {
        await navigateTo({ path: '/_login', query: { returnTo: route.fullPath } })
        throw new Error('redirected to /_login')
      }
    }

    throw err
  }
}
