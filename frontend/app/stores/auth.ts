import { defineStore } from 'pinia'

interface UserProfile {
  user_id: number
  email: string
  name: string
  phone: string
  created_at?: string
}

type Role = 'customer' | 'driver' | 'manager' | null

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: null as string | null,
    refreshToken: null as string | null,
    user: null as UserProfile | null,
    role: null as Role,
    driverId: null as number | null,
    restaurantId: null as number | null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.accessToken,
  },

  actions: {
    init() {
      if (!import.meta.client) return
      this.accessToken = localStorage.getItem('access_token')
      this.refreshToken = localStorage.getItem('refresh_token')
      const u = localStorage.getItem('user')
      if (u) this.user = JSON.parse(u)
      this.role = (localStorage.getItem('role') as Role) ?? null
      const did = localStorage.getItem('driver_id')
      this.driverId = did ? Number(did) : null
      const rid = localStorage.getItem('restaurant_id')
      this.restaurantId = rid ? Number(rid) : null
    },

    setTokens(access: string, refresh: string) {
      this.accessToken = access
      this.refreshToken = refresh
      if (import.meta.client) {
        localStorage.setItem('access_token', access)
        localStorage.setItem('refresh_token', refresh)
      }
    },

    setUser(user: UserProfile) {
      this.user = user
      if (import.meta.client) {
        localStorage.setItem('user', JSON.stringify(user))
      }
    },

    setRole(role: Role, driverId?: number | null, restaurantId?: number | null) {
      this.role = role
      this.driverId = driverId ?? null
      this.restaurantId = restaurantId ?? null
      if (import.meta.client) {
        if (role) localStorage.setItem('role', role)
        else localStorage.removeItem('role')
        if (driverId) localStorage.setItem('driver_id', String(driverId))
        else localStorage.removeItem('driver_id')
        if (restaurantId) localStorage.setItem('restaurant_id', String(restaurantId))
        else localStorage.removeItem('restaurant_id')
      }
    },

    logout() {
      this.accessToken = null
      this.refreshToken = null
      this.user = null
      this.role = null
      this.driverId = null
      this.restaurantId = null
      if (import.meta.client) {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user')
        localStorage.removeItem('role')
        localStorage.removeItem('driver_id')
        localStorage.removeItem('restaurant_id')
      }
    },
  },
})
