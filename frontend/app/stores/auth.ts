import { defineStore } from 'pinia'

interface UserProfile {
  user_id: number
  email: string
  name: string
  phone: string
  created_at?: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: null as string | null,
    refreshToken: null as string | null,
    user: null as UserProfile | null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.accessToken,
  },

  actions: {
    init() {
      if (import.meta.client) {
        this.accessToken = localStorage.getItem('access_token')
        this.refreshToken = localStorage.getItem('refresh_token')
        const u = localStorage.getItem('user')
        if (u) this.user = JSON.parse(u)
      }
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

    logout() {
      this.accessToken = null
      this.refreshToken = null
      this.user = null
      if (import.meta.client) {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user')
      }
    },
  },
})
