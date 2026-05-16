<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useApi } from '~/composables/useApi'

const { apiFetch } = useApi()

const addresses = ref<any[]>([])
const form = reactive({ street: '', city: '', zip: '', is_default: false })
const loading = ref(false)
const showForm = ref(false)
const error = ref('')

onMounted(load)

async function load() {
  try {
    const data = await apiFetch<{ addresses: any[] }>('/api/users/addresses')
    addresses.value = data.addresses ?? []
  } catch {}
}

async function add() {
  loading.value = true
  error.value = ''
  try {
    await apiFetch('/api/users/addresses', {
      method: 'POST',
      body: JSON.stringify(form),
    })
    form.street = ''
    form.city = ''
    form.zip = ''
    form.is_default = false
    showForm.value = false
    await load()
  } catch (e: any) {
    error.value = e?.error ?? 'Ошибка'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <NuxtLink to="/profile" class="text-sm text-brand-600 hover:underline">← Профиль</NuxtLink>
      <h1 class="text-xl font-bold text-gray-900">Адреса</h1>
    </div>

    <div class="space-y-3 mb-4">
      <div
        v-for="addr in addresses"
        :key="addr.address_id"
        class="bg-white rounded-xl p-4 shadow-sm"
      >
        <div class="flex items-start justify-between">
          <div>
            <p class="text-sm font-medium text-gray-900">{{ addr.street }}</p>
            <p class="text-xs text-gray-500 mt-0.5">{{ addr.city }}<span v-if="addr.zip">, {{ addr.zip }}</span></p>
          </div>
          <span v-if="addr.is_default" class="text-xs bg-brand-100 text-brand-700 px-2 py-0.5 rounded-full">по умолчанию</span>
        </div>
      </div>
      <p v-if="addresses.length === 0" class="text-sm text-gray-400 text-center py-4">Нет добавленных адресов</p>
    </div>

    <button
      v-if="!showForm"
      @click="showForm = true"
      class="w-full border-2 border-dashed border-gray-300 text-gray-500 rounded-xl py-3 text-sm hover:border-brand-400 hover:text-brand-600 transition-colors"
    >
      + Добавить адрес
    </button>

    <form v-else @submit.prevent="add" class="bg-white rounded-xl p-4 shadow-sm space-y-3">
      <h2 class="font-medium text-gray-900">Новый адрес</h2>
      <input v-model="form.street" placeholder="Улица и дом" required class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      <input v-model="form.city" placeholder="Город" required class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      <input v-model="form.zip" placeholder="Индекс (опционально)" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
      <label class="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
        <input type="checkbox" v-model="form.is_default" class="text-brand-500" />
        Сделать основным
      </label>
      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
      <div class="flex gap-2">
        <button type="button" @click="showForm = false" class="flex-1 border border-gray-300 text-gray-600 py-2 rounded-lg text-sm hover:bg-gray-50">Отмена</button>
        <button type="submit" :disabled="loading" class="flex-1 bg-brand-500 text-white py-2 rounded-lg text-sm hover:bg-brand-600 disabled:opacity-60">
          {{ loading ? 'Сохранение...' : 'Сохранить' }}
        </button>
      </div>
    </form>
  </div>
</template>
