import { FetchError } from 'ofetch'
import { defineStore } from 'pinia'
import type { TokenUserResponse, User } from '~/types'

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

    function logout() {
      token.value = ''
      user.value = undefined
    }

    return { login, logout, loggedIn, user, token }
  },
)
