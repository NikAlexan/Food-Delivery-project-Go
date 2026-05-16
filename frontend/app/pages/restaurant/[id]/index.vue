<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useApi } from '~/composables/useApi'

const route = useRoute()
const { apiFetch } = useApi()

const form = reactive({
  name: '',
  description: '',
  address: '',
  phone: '',
  image_url: '',
  is_active: true,
})
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const error = ref('')

onMounted(async () => {
  try {
    const r = await apiFetch<any>(`/api/restaurants/${route.params.id}`)
    form.name = r.name ?? ''
    form.description = r.description ?? ''
    form.address = r.address ?? ''
    form.phone = r.phone ?? ''
    form.image_url = r.image_url ?? ''
    form.is_active = r.is_active ?? true
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    await apiFetch(`/api/restaurants/${route.params.id}`, {
      method: 'PUT',
      body: JSON.stringify({ ...form, restaurant_id: Number(route.params.id) }),
    })
    saved.value = true
    setTimeout(() => (saved.value = false), 2000)
  } catch (e: any) {
    error.value = e?.error ?? 'Ошибка сохранения'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto">
    <div class="flex items-center gap-4 mb-6">
      <h1 class="text-xl font-bold text-gray-900 flex-1">Мой ресторан</h1>
      <NuxtLink
        :to="`/restaurant/${route.params.id}/menu`"
        class="text-sm bg-brand-500 text-white px-3 py-1.5 rounded-lg hover:bg-brand-600 transition-colors"
      >Управление меню</NuxtLink>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Загрузка...</div>

    <form v-else @submit.prevent="save" class="bg-white rounded-xl shadow-sm p-5 space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Название</label>
        <input v-model="form.name" required class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
        <textarea v-model="form.description" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Адрес</label>
        <input v-model="form.address" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Телефон</label>
        <input v-model="form.phone" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">URL изображения</label>
        <input v-model="form.image_url" type="url" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      </div>
      <label class="flex items-center gap-2 cursor-pointer">
        <input type="checkbox" v-model="form.is_active" class="rounded" />
        <span class="text-sm text-gray-700">Ресторан активен (виден в списке)</span>
      </label>

      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
      <p v-if="saved" class="text-sm text-green-600">Сохранено ✓</p>

      <button
        type="submit"
        :disabled="saving"
        class="w-full bg-brand-500 hover:bg-brand-600 text-white font-medium py-2 rounded-lg transition-colors disabled:opacity-60"
      >
        {{ saving ? 'Сохранение...' : 'Сохранить изменения' }}
      </button>
    </form>
  </div>
</template>
