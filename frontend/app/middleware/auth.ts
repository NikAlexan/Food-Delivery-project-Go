export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const publicPaths = ['/auth/login', '/auth/register', '/restaurants']
  const isPublic = publicPaths.some((p) => to.path === p || to.path.startsWith('/restaurants/'))

  if (isPublic) return

  const token = localStorage.getItem('access_token')
  if (!token) {
    return navigateTo('/auth/login')
  }
})
