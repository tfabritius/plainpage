import type { NitroFetchOptions, NitroFetchRequest } from 'nitropack'
import { useAuthStore } from '~/store/auth'

export function apiFetch<T>(request: string, opts?: NitroFetchOptions<NitroFetchRequest>) {
  const { loggedIn, token } = useAuthStore()

  const headers = { ...opts?.headers, authorization: loggedIn ? `Bearer ${token}` : '' }

  return $fetch<T>(request, {
    ...opts,
    headers,
    baseURL: opts?.baseURL || '/_api',
  })
}
