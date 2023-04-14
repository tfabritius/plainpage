import { FetchError } from 'ofetch'
import { defineStore } from 'pinia'
import type { PatchOperation, TokenUserResponse, User } from '~/types'

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

    return { login, logout, loggedIn, user, token, updateMe, deleteMe }
  },
  {
    persist: true,
  },
)
