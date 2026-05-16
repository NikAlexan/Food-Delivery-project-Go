<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useApi } from '~/composables/useApi'

const route = useRoute()
const { apiFetch } = useApi()
const rid = route.params.id

const items = ref<any[]>([])
const loading = ref(true)
const showAdd = ref(false)
const editingId = ref<number | null>(null)

const addForm = reactive({ name: '', description: '', price: 0, category: '', image_url: '' })
const editForm = reactive({ name: '', description: '', price: 0, category: '', image_url: '', is_available: true })
const addLoading = ref(false)
const editLoading = ref(false)
const addError = ref('')

onMounted(load)

async function load() {
  loading.value = true
  try {
    const data = await apiFetch<{ items: any[] }>(`/api/restaurants/${rid}/menu`)
    items.value = data.items ?? []
  } finally {
    loading.value = false
  }
}

async function addItem() {
  addLoading.value = true
  addError.value = ''
  try {
    await apiFetch(`/api/restaurants/${rid}/menu`, {
      method: 'POST',
      body: JSON.stringify({ ...addForm, price: Number(addForm.price), restaurant_id: Number(rid) }),
    })
    showAdd.value = false
    Object.assign(addForm, { name: '', description: '', price: 0, category: '', image_url: '' })
    await load()
  } catch (e: any) {
    addError.value = e?.error ?? 'Ошибка'
  } finally {
    addLoading.value = false
  }
}

function startEdit(item: any) {
  editingId.value = item.item_id
  editForm.name = item.name
  editForm.description = item.description
  editForm.price = item.price
  editForm.category = item.category
  editForm.image_url = item.image_url
  editForm.is_available = item.is_available
}

async function saveEdit(itemId: number) {
  editLoading.value = true
  try {
    await apiFetch(`/api/restaurants/${rid}/menu/${itemId}`, {
      method: 'PUT',
      body: JSON.stringify({ ...editForm, price: Number(editForm.price), item_id: itemId, restaurant_id: Number(rid) }),
    })
    editingId.value = null
    await load()
  } finally {
    editLoading.value = false
  }
}

async function toggleAvail(item: any) {
  await apiFetch(`/api/restaurants/${rid}/menu/${item.item_id}`, {
    method: 'PUT',
    body: JSON.stringify({
      item_id: item.item_id,
      restaurant_id: Number(rid),
      name: item.name,
      description: item.description,
      price: item.price,
      category: item.category,
      image_url: item.image_url,
      is_available: !item.is_available,
    }),
  })
  await load()
}

async function deleteItem(itemId: number) {
  if (!confirm('Удалить позицию?')) return
  await apiFetch(`/api/restaurants/${rid}/menu/${itemId}`, { method: 'DELETE' })
  await load()
}
</script>

<template>
  <div>
    <div class="flex items-center gap-4 mb-6">
      <NuxtLink :to="`/restaurant/${rid}`" class="text-sm text-brand-600 hover:underline">← Ресторан</NuxtLink>
      <h1 class="text-xl font-bold text-gray-900 flex-1">Управление меню</h1>
      <button
        @click="showAdd = !showAdd"
        class="text-sm bg-brand-500 text-white px-3 py-1.5 rounded-lg hover:bg-brand-600 transition-colors"
      >+ Добавить</button>
    </div>

    <!-- Форма добавления -->
    <div v-if="showAdd" class="bg-white rounded-xl shadow-sm p-4 mb-4 space-y-3">
      <h2 class="font-medium text-gray-900">Новое блюдо</h2>
      <div class="grid grid-cols-2 gap-3">
        <div class="col-span-2">
          <input v-model="addForm.name" placeholder="Название *" required class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
        </div>
        <input v-model="addForm.price" type="number" step="0.01" placeholder="Цена *" class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
        <input v-model="addForm.category" placeholder="Категория" class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
        <div class="col-span-2">
          <textarea v-model="addForm.description" placeholder="Описание" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
        </div>
        <div class="col-span-2">
          <input v-model="addForm.image_url" type="url" placeholder="URL изображения" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
        </div>
      </div>
      <p v-if="addError" class="text-sm text-red-500">{{ addError }}</p>
      <div class="flex gap-2">
        <button @click="showAdd = false" class="flex-1 border border-gray-300 text-gray-600 py-2 rounded-lg text-sm hover:bg-gray-50">Отмена</button>
        <button @click="addItem" :disabled="addLoading" class="flex-1 bg-brand-500 text-white py-2 rounded-lg text-sm hover:bg-brand-600 disabled:opacity-60">
          {{ addLoading ? 'Сохранение...' : 'Добавить' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Загрузка...</div>
    <p v-else-if="items.length === 0" class="text-center py-8 text-gray-400">Меню пустое</p>

    <div v-else class="space-y-3">
      <div v-for="item in items" :key="item.item_id" class="bg-white rounded-xl shadow-sm overflow-hidden">
        <!-- Просмотр -->
        <div v-if="editingId !== item.item_id" class="p-4 flex items-center gap-4">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-medium text-gray-900">{{ item.name }}</span>
              <span class="text-xs text-gray-400">{{ item.category }}</span>
              <span
                class="text-xs px-1.5 py-0.5 rounded"
                :class="item.is_available ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'"
              >{{ item.is_available ? 'Доступно' : 'Скрыто' }}</span>
            </div>
            <p class="text-sm text-gray-500 mt-0.5 truncate">{{ item.description }}</p>
          </div>
          <span class="font-semibold text-brand-600 whitespace-nowrap">{{ item.price.toFixed(2) }} ₸</span>
          <div class="flex gap-1">
            <button @click="toggleAvail(item)" class="p-1.5 rounded hover:bg-gray-100 text-gray-400 hover:text-gray-700 transition-colors text-sm" title="Переключить доступность">👁</button>
            <button @click="startEdit(item)" class="p-1.5 rounded hover:bg-gray-100 text-gray-400 hover:text-blue-600 transition-colors text-sm" title="Редактировать">✏️</button>
            <button @click="deleteItem(item.item_id)" class="p-1.5 rounded hover:bg-gray-100 text-gray-400 hover:text-red-500 transition-colors text-sm" title="Удалить">🗑</button>
          </div>
        </div>

        <!-- Редактирование inline -->
        <div v-else class="p-4 space-y-3 bg-blue-50">
          <div class="grid grid-cols-2 gap-2">
            <div class="col-span-2">
              <input v-model="editForm.name" placeholder="Название" class="w-full border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
            </div>
            <input v-model="editForm.price" type="number" step="0.01" placeholder="Цена" class="border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
            <input v-model="editForm.category" placeholder="Категория" class="border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
            <div class="col-span-2">
              <textarea v-model="editForm.description" rows="2" placeholder="Описание" class="w-full border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
            </div>
            <div class="col-span-2">
              <input v-model="editForm.image_url" type="url" placeholder="URL изображения" class="w-full border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500" />
            </div>
          </div>
          <label class="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
            <input type="checkbox" v-model="editForm.is_available" />
            Доступно для заказа
          </label>
          <div class="flex gap-2">
            <button @click="editingId = null" class="flex-1 border border-gray-300 text-gray-600 py-1.5 rounded-lg text-sm hover:bg-white">Отмена</button>
            <button @click="saveEdit(item.item_id)" :disabled="editLoading" class="flex-1 bg-brand-500 text-white py-1.5 rounded-lg text-sm hover:bg-brand-600 disabled:opacity-60">Сохранить</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
