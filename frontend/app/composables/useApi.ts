import { useAuthStore } from '~/stores/auth'

export function useApi() {
  const auth = useAuthStore()

  async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string> ?? {}),
    }

    if (auth.accessToken) {
      headers['Authorization'] = `Bearer ${auth.accessToken}`
    }

    const res = await fetch(path, { ...options, headers })

    if (res.status === 401 && auth.refreshToken) {
      const refreshed = await tryRefresh()
      if (refreshed) {
        headers['Authorization'] = `Bearer ${auth.accessToken}`
        const retry = await fetch(path, { ...options, headers })
        if (!retry.ok) throw await retry.json()
        return retry.json() as Promise<T>
      } else {
        auth.logout()
        navigateTo('/auth/login')
        throw new Error('Session expired')
      }
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw err
    }

    const text = await res.text()
    return text ? JSON.parse(text) : ({} as T)
  }

  async function tryRefresh(): Promise<boolean> {
    try {
      const res = await fetch('/api/users/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: auth.refreshToken }),
      })
      if (!res.ok) return false
      const data = await res.json()
      auth.setTokens(data.access_token, data.refresh_token)
      return true
    } catch {
      return false
    }
  }

  return { apiFetch }
}
