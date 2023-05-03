import { useAuthStore } from '~/store/auth'

export default defineNuxtRouteMiddleware((to, _from) => {
  const auth = useAuthStore()
  if (!auth.loggedIn) {
    return navigateTo(`/_login?returnTo=${encodeURIComponent(to.fullPath)}`)
  }
})
