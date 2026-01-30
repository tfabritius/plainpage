import type { PatchOperation, TokenUserResponse, User } from '~/types'
import { FetchError } from 'ofetch'
import { defineStore } from 'pinia'
import { useAppStore } from './app'

const hour = 60 * 60 // in seconds
const minRemainingTokenValidity = 5 * 30 * 24 * hour // 5 months

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
    const token = ref('')
    const loggedIn = computed(() => token.value !== '')

    async function login(credentials: { username: string, password: string }): Promise<true | LoginError> {
      token.value = ''
      user.value = undefined

      try {
        const response = await apiFetch<TokenUserResponse>('/auth/login', { body: credentials, method: 'POST' })

        token.value = response.token
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

    async function renewToken() {
      if (!loggedIn.value) {
        throw new Error('not logged in')
      }

      const response = await apiFetch<TokenUserResponse>('/auth/refresh', { method: 'POST' })

      token.value = response.token
      user.value = response.user
    }

    async function logout() {
      token.value = ''
      user.value = undefined

      // Reload app data to remove version info now that user is logged out
      const appStore = useAppStore()
      await appStore.refresh()

      // Run middlewares of current page again
      const router = useRouter()
      await router.replace({ path: router.currentRoute.value.fullPath, force: true })
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
      await apiFetch(`/auth/users/_me/password`, {
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

    /**
     * Parses token and returns expiration time or 0 if token is not set
     */
    const tokenExpiration = computed(() => {
      if (!token.value) {
        return 0
      }

      const jsonStr = window.atob(token.value.split('.')[1] ?? '')

      const payload = JSON.parse(jsonStr) as unknown
      if (typeof payload === 'object'
        && payload !== null
        && 'exp' in payload
        && typeof payload.exp === 'number'
      ) {
        return payload.exp
      }
      throw new Error('invalid token payload format')
    })

    // Interval limits the changes to 1/min,
    // limiting the frequency of refreshing the token
    const now = useTimestamp({ interval: 60 * 1000 })

    const tokenRemainingSeconds = computed(() => {
      return Math.max(0, tokenExpiration.value - Math.floor(now.value / 1000))
    })

    /**
     * Watches token expiration and renews it if needed
     */
    watch(tokenRemainingSeconds, async (time) => {
      if (loggedIn.value && time < minRemainingTokenValidity) {
        await renewToken()
      }
    })

    return { login, logout, loggedIn, user, token, updateMe, changePassword, deleteMe }
  },
  {
    persist: true,
  },
)
