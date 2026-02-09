import type { LoginResponse, PatchOperation, RefreshResponse, User } from '~/types'
import { FetchError } from 'ofetch'
import { defineStore } from 'pinia'
import { apiRawFetch } from '~/composables/apiFetch'
import { useAppStore } from './app'

export type LoginError = {
  statusCode: 401
} | {
  statusCode: 429
  retryAfter: string
}

export const useAuthStore = defineStore(
  'auth',
  () => {
    const user = ref<User>()
    const accessToken = ref('')
    const loggedIn = computed(() => accessToken.value !== '')

    async function login(credentials: { username: string, password: string }): Promise<true | LoginError> {
      accessToken.value = ''
      user.value = undefined

      try {
        const response = await apiRawFetch<LoginResponse>('/auth/login', {
          body: credentials,
          method: 'POST',
          credentials: 'include', // Include cookies for refresh token
        })

        accessToken.value = response.accessToken
        user.value = response.user

        // Reload app data to get version info now that user is logged in
        const appStore = useAppStore()
        await appStore.refresh()
      } catch (err) {
        if (err instanceof FetchError && err.statusCode === 401) {
          return { statusCode: 401 }
        }
        if (err instanceof FetchError && err.statusCode === 429) {
          const retryAfter = err.response?.headers.get('Retry-After') ?? '1'
          return {
            statusCode: 429,
            retryAfter,
          }
        }
        throw err
      }

      return true
    }

    // Track if we're currently refreshing to prevent multiple refresh calls
    let refreshPromise: Promise<void> | null = null

    async function refreshAccessToken(): Promise<boolean> {
      // If already refreshing, wait for that to complete
      if (refreshPromise) {
        await refreshPromise
        return loggedIn.value
      }

      refreshPromise = (async () => {
        try {
          const response = await apiRawFetch<RefreshResponse>('/auth/refresh', {
            method: 'POST',
          })

          accessToken.value = response.accessToken
          user.value = response.user
        } catch {
          // Refresh failed, clear auth state
          accessToken.value = ''
          user.value = undefined
        } finally {
          refreshPromise = null
        }
      })()

      await refreshPromise
      return loggedIn.value
    }

    async function logout() {
      // Call logout API to revoke refresh token
      try {
        await apiRawFetch('/auth/logout', {
          method: 'POST',
        })
      } catch {
        // Ignore errors during logout
      }

      accessToken.value = ''
      user.value = undefined

      // Reload app data to remove version info now that user is logged out
      const appStore = useAppStore()
      await appStore.refresh()
    }

    async function updateMe(newMe: { displayName: string }) {
      if (!user.value) {
        throw new Error('not logged in')
      }
      const ops: PatchOperation[] = [
        { op: 'replace', path: '/displayName', value: newMe.displayName },
      ]
      await apiFetch(`/auth/users/${user.value.username}`, {
        method: 'PATCH',
        body: ops,
      })
      user.value.displayName = newMe.displayName
    }

    async function changePassword(currentPassword: string, newPassword: string) {
      if (!user.value) {
        throw new Error('not logged in')
      }
      await apiFetch(`/auth/users/${user.value.username}/password`, {
        method: 'POST',
        body: { currentPassword, newPassword },
      })
    }

    async function deleteMe() {
      if (!user.value) {
        throw new Error('not logged in')
      }
      await apiFetch(`/auth/users/${user.value.username}`, { method: 'DELETE' })
      await logout()
    }

    return {
      login,
      logout,
      loggedIn,
      user,
      accessToken,
      updateMe,
      changePassword,
      deleteMe,
      refreshAccessToken,
    }
  },
  {
    persist: {
      // Only persist accessToken and user
      pick: ['accessToken', 'user'],
    },
  },
)
