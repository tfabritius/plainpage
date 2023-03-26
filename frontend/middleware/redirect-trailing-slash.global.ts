export default defineNuxtRouteMiddleware((to, _from) => {
  if (to.path.endsWith('/') && to.path !== '/') {
    const { path, query, hash } = to
    const newPath = path.replace(/\/+$/, '') || '/'
    const newRoute = { path: newPath, query, hash }
    return navigateTo(newRoute, { redirectCode: 301 })
  }
})
