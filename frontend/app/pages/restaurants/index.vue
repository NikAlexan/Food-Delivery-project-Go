<script setup lang="ts">
definePageMeta({ middleware: [] })

interface Restaurant {
  restaurant_id: number
  name: string
  description: string
  address: string
  category_name: string
  rating: number
  image_url: string
  is_active: boolean
}

const query = ref('')
const restaurants = ref<Restaurant[]>([])
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const url = query.value.trim()
      ? `/api/restaurants/search?q=${encodeURIComponent(query.value)}`
      : '/api/restaurants'
    const data = await $fetch<{ restaurants: Restaurant[] }>(url)
    restaurants.value = data.restaurants ?? []
  } catch {
    error.value = 'Не удалось загрузить рестораны'
  } finally {
    loading.value = false
  }
}

const debouncedLoad = useDebounceFn(load, 400)

watch(query, debouncedLoad)
onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-6">
      <h1 class="text-xl font-bold text-gray-900 flex-1">Рестораны</h1>
      <input
        v-model="query"
        type="search"
        placeholder="Поиск ресторанов..."
        class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-brand-500"
      />
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Загрузка...</div>
    <p v-else-if="error" class="text-red-500 text-sm">{{ error }}</p>
    <p v-else-if="restaurants.length === 0" class="text-center py-12 text-gray-400">Ничего не найдено</p>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <RestaurantCard v-for="r in restaurants" :key="r.restaurant_id" :restaurant="r" />
    </div>
  </div>
</template>
