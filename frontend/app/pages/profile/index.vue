<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const auth = useAuthStore()
const { apiFetch } = useApi()

const form = reactive({ name: '', phone: '' })
const loading = ref(false)
const saved = ref(false)
const error = ref('')

onMounted(async () => {
  auth.init()
  try {
    const profile = await apiFetch<any>('/api/users/profile')
    auth.setUser(profile)
    form.name = profile.name ?? ''
    form.phone = profile.phone ?? ''
  } catch {}
})

async function save() {
  loading.value = true
  error.value = ''
  saved.value = false
  try {
    const res = await apiFetch<any>('/api/users/profile', {
      method: 'PUT',
      body: JSON.stringify(form),
    })
    if (auth.user) auth.setUser({ ...auth.user, ...res })
    saved.value = true
    setTimeout(() => (saved.value = false), 2000)
  } catch (e: any) {
    error.value = e?.error ?? 'Ошибка сохранения'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-bold text-gray-900">Профиль</h1>
      <NuxtLink to="/profile/addresses" class="text-sm text-brand-600 hover:underline">Адреса →</NuxtLink>
    </div>

    <div class="bg-white rounded-xl shadow-sm p-5 space-y-4">
      <div>
        <p class="text-xs text-gray-400 mb-1">Email</p>
        <p class="text-sm text-gray-700">{{ auth.user?.email }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Имя</label>
        <input
          v-model="form.name"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Телефон</label>
        <input
          v-model="form.phone"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
        />
      </div>

      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
      <p v-if="saved" class="text-sm text-green-600">Сохранено ✓</p>

      <button
        @click="save"
        :disabled="loading"
        class="w-full bg-brand-500 hover:bg-brand-600 text-white font-medium py-2 rounded-lg transition-colors disabled:opacity-60"
      >
        {{ loading ? 'Сохранение...' : 'Сохранить' }}
      </button>
    </div>
  </div>
</template>
