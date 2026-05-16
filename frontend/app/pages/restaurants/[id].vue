<script setup lang="ts">
definePageMeta({ middleware: [] })

const route = useRoute()
const id = route.params.id

const restaurant = ref<any>(null)
const menu = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [r, m] = await Promise.all([
      $fetch<any>(`/api/restaurants/${id}`),
      $fetch<{ items: any[] }>(`/api/restaurants/${id}/menu`),
    ])
    restaurant.value = r
    menu.value = m.items ?? []
  } finally {
    loading.value = false
  }
})

const grouped = computed(() => {
  const map: Record<string, any[]> = {}
  for (const item of menu.value) {
    const cat = item.category || 'Прочее'
    if (!map[cat]) map[cat] = []
    map[cat].push(item)
  }
  return map
})
</script>

<template>
  <div>
    <NuxtLink to="/restaurants" class="text-sm text-brand-600 hover:underline">← Все рестораны</NuxtLink>

    <div v-if="loading" class="text-center py-12 text-gray-400 mt-4">Загрузка...</div>

    <template v-else-if="restaurant">
      <div class="mt-4 bg-white rounded-xl shadow-sm overflow-hidden">
        <div v-if="restaurant.image_url" class="h-48 bg-gray-100">
          <img :src="restaurant.image_url" class="w-full h-full object-cover" alt="" />
        </div>
        <div class="p-6">
          <div class="flex items-start justify-between">
            <div>
              <h1 class="text-2xl font-bold text-gray-900">{{ restaurant.name }}</h1>
              <p class="text-sm text-gray-500 mt-1">{{ restaurant.category_name }}</p>
            </div>
            <span class="text-yellow-500 font-medium text-lg">★ {{ restaurant.rating?.toFixed(1) ?? '—' }}</span>
          </div>
          <p class="text-gray-600 mt-2">{{ restaurant.description }}</p>
          <p class="text-sm text-gray-400 mt-1">📍 {{ restaurant.address }}</p>
        </div>
      </div>

      <div class="mt-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Меню</h2>
        <div v-if="menu.length === 0" class="text-gray-400 text-sm">Меню пока не добавлено</div>

        <div v-for="(items, category) in grouped" :key="category" class="mb-6">
          <h3 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">{{ category }}</h3>
          <div class="space-y-3">
            <MenuItemCard
              v-for="item in items"
              :key="item.item_id"
              :item="item"
              :restaurant-id="Number(id)"
              :restaurant-name="restaurant.name"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
