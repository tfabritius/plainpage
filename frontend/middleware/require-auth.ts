import { useAuthStore } from '~/store/auth'

export default defineNuxtRouteMiddleware((_to, _from) => {
  const auth = useAuthStore()
  const route = useRoute()
  if (!auth.loggedIn) {
    return navigateTo(`/_login?returnTo=${encodeURIComponent(route.fullPath)}`)
  }
})
