import { FetchError } from 'ofetch'
import { defineStore } from 'pinia'
import type { PatchOperation, TokenUserResponse, User } from '~/types'

const minRemainingTokenValidity = 6 * 24 * 60 * 60 // in seconds

export const useAuthStore = defineStore(
  'auth',
  () => {
    const user = ref<User>()
    const token = ref('')
    const loggedIn = computed(() => token.value !== '')

    async function login(credentials: { username: string; password: string }): Promise<boolean> {
      token.value = ''
      user.value = undefined

      try {
        const response = await apiFetch<TokenUserResponse>('/auth/login', { body: credentials, method: 'POST' })

        token.value = response.token
        user.value = response.user
      } catch (err) {
        if (err instanceof FetchError && err.statusCode === 401) {
          return false
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

    function logout() {
      token.value = ''
      user.value = undefined

      // Refresh current page
      const router = useRouter()
      router.go(0)
    }

    async function updateMe(newMe: { displayName: string; password: string }) {
      if (!user.value) {
        throw new Error('not logged in')
      }
      const ops: PatchOperation[] = [
        { op: 'replace', path: '/displayName', value: newMe.displayName },
      ]
      if (newMe.password) {
        ops.push({ op: 'replace', path: '/password', value: newMe.password })
      }
      await apiFetch(`/auth/users/${user.value.username}`, {
        method: 'PATCH',
        body: ops,
      })
      user.value.displayName = newMe.displayName
    }

    async function deleteMe() {
      if (!user.value) {
        throw new Error('not logged in')
      }
      await apiFetch(`/auth/users/${user.value.username}`, { method: 'DELETE' })
      logout()
    }

    /**
     * Parses token and returns expiration time or 0 if token is not set
     */
    const tokenExpiration = computed(() => {
      if (!token.value) {
        return 0
      }

      const jsonStr = window.atob(token.value.split('.')[1])

      const payload = JSON.parse(jsonStr) as unknown
      if (typeof payload === 'object' && payload !== null && 'exp' in payload && typeof payload.exp === 'number') {
        return payload.exp
      }
      throw new Error('invalid token payload format')
    })

    const now = useTimestamp({ interval: 1000 })

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

    return { login, logout, loggedIn, user, token, updateMe, deleteMe }
  },
  {
    persist: true,
  },
)
