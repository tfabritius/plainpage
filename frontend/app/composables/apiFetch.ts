import type { NitroFetchOptions, NitroFetchRequest } from 'nitropack'
import { FetchError } from 'ofetch'
import { useAuthStore } from '~/store/auth'

// Refresh token if it will expire within this many seconds
const tokenExpirationBuffer = 5

/**
 * Raw API fetch without auth interceptor logic.
 */
export async function apiRawFetch<T>(request: string, opts?: NitroFetchOptions<NitroFetchRequest>): Promise<T> {
  return await $fetch<T>(request, {
    ...opts,
    baseURL: opts?.baseURL || '/_api',
    credentials: 'include',
  })
}

/**
 * Parses JWT and returns expiration timestamp, or 0 if invalid/not set
 */
function getTokenExpiration(token: string): number {
  if (!token) {
    return 0
  }

  const jsonStr = window.atob(token.split('.')[1] ?? '')
  const payload = JSON.parse(jsonStr) as unknown
  if (typeof payload === 'object'
    && payload !== null
    && 'exp' in payload
    && typeof payload.exp === 'number'
  ) {
    return payload.exp
  }

  return 0
}

/**
 * Checks if token is expired or will expire within the buffer time
 */
function isTokenExpiringSoon(token: string): boolean {
  const exp = getTokenExpiration(token)
  if (exp === 0) {
    return false
  }

  const now = Math.floor(Date.now() / 1000)
  return exp - now < tokenExpirationBuffer
}

export async function apiFetch<T>(request: string, opts?: NitroFetchOptions<NitroFetchRequest>): Promise<T> {
  const authStore = useAuthStore()
  const route = useRoute()

  // Proactively refresh token before making request if it's expired or expiring soon
  if (authStore.accessToken && isTokenExpiringSoon(authStore.accessToken)) {
    await authStore.refreshAccessToken()
  }

  const makeRequest = async (): Promise<T> => {
    const headers = {
      ...opts?.headers,
      authorization: authStore.loggedIn ? `Bearer ${authStore.accessToken}` : '',
    }

    return await $fetch<T>(request, {
      ...opts,
      headers,
      baseURL: opts?.baseURL || '/_api',
      credentials: 'include', // Always include cookies for refresh token
    })
  }

  try {
    return await makeRequest()
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 401) {
      // If we have a refresh token (indicated by being logged in previously),
      // try to refresh the access token
      if (authStore.loggedIn || authStore.accessToken) {
        const refreshed = await authStore.refreshAccessToken()
        if (refreshed) {
          // Retry the original request with new access token
          try {
            return await makeRequest()
          } catch (retryErr) {
            if (retryErr instanceof FetchError && retryErr.statusCode === 401) {
              // Even after refresh, still unauthorized
              await handleUnauthorized(authStore, route)
            }
            throw retryErr
          }
        }
      }

      // No refresh token or refresh failed
      await handleUnauthorized(authStore, route)
    }

    throw err
  }
}

async function handleUnauthorized(authStore: ReturnType<typeof useAuthStore>, route: ReturnType<typeof useRoute>) {
  if (authStore.loggedIn || authStore.accessToken) {
    // Clear the auth state without calling logout API (we're already unauthorized)
    authStore.accessToken = ''
    authStore.user = undefined
  }

  if (route.path !== '/_login') {
    await navigateTo({ path: '/_login', query: { returnTo: route.fullPath } })
    throw new Error('redirected to /_login')
  }
}
