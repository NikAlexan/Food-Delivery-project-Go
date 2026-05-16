<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

definePageMeta({ middleware: [] })

const auth = useAuthStore()
const form = reactive({ email: '', password: '' })
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await $fetch<{ access_token: string; refresh_token: string }>('/api/users/login', {
      method: 'POST',
      body: form,
    })
    auth.setTokens(res.access_token, res.refresh_token)

    const profile = await $fetch<any>('/api/users/profile', {
      headers: { Authorization: `Bearer ${res.access_token}` },
    })
    auth.setUser(profile)

    navigateTo('/restaurants')
  } catch (e: any) {
    error.value = e?.error ?? 'Неверный email или пароль'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-sm mx-auto mt-16">
    <h1 class="text-2xl font-bold text-gray-900 text-center mb-6">Вход</h1>

    <form @submit.prevent="submit" class="bg-white rounded-xl shadow-sm p-6 space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input
          v-model="form.email"
          type="email"
          required
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
        <input
          v-model="form.password"
          type="password"
          required
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
        />
      </div>

      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>

      <button
        type="submit"
        :disabled="loading"
        class="w-full bg-brand-500 hover:bg-brand-600 text-white font-medium py-2 rounded-lg transition-colors disabled:opacity-60"
      >
        {{ loading ? 'Вход...' : 'Войти' }}
      </button>

      <p class="text-sm text-center text-gray-500">
        Нет аккаунта?
        <NuxtLink to="/auth/register" class="text-brand-600 hover:underline">Зарегистрироваться</NuxtLink>
      </p>
    </form>
  </div>
</template>
