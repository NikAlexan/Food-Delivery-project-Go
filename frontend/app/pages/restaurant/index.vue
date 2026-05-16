<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const auth = useAuthStore()
const { apiFetch } = useApi()

const form = reactive({
  name: '',
  description: '',
  address: '',
  phone: '',
  category_id: 1,
  image_url: '',
})
const loading = ref(false)
const error = ref('')

onMounted(() => {
  auth.init()
  if (auth.restaurantId) navigateTo(`/restaurant/${auth.restaurantId}`)
})

async function create() {
  loading.value = true
  error.value = ''
  try {
    const res = await apiFetch<any>('/api/restaurants', {
      method: 'POST',
      body: JSON.stringify({ ...form, category_id: Number(form.category_id) }),
    })
    auth.setRole('manager', null, res.restaurant_id)
    navigateTo(`/restaurant/${res.restaurant_id}`)
  } catch (e: any) {
    error.value = e?.error ?? 'Ошибка создания'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto">
    <h1 class="text-xl font-bold text-gray-900 mb-6">Создать ресторан</h1>

    <form @submit.prevent="create" class="bg-white rounded-xl shadow-sm p-5 space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Название *</label>
        <input v-model="form.name" required class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
        <textarea v-model="form.description" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Адрес *</label>
        <input v-model="form.address" required class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Телефон</label>
        <input v-model="form.phone" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Category ID</label>
        <input v-model="form.category_id" type="number" min="1" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">URL изображения</label>
        <input v-model="form.image_url" type="url" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>

      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>

      <button
        type="submit"
        :disabled="loading"
        class="w-full bg-brand-500 hover:bg-brand-600 text-white font-medium py-2 rounded-lg transition-colors disabled:opacity-60"
      >
        {{ loading ? 'Создание...' : 'Создать ресторан' }}
      </button>
    </form>
  </div>
</template>
