<script setup lang="ts">
defineProps<{
  restaurant: {
    restaurant_id: number
    name: string
    description: string
    address: string
    category_name: string
    rating: number
    image_url: string
    is_active: boolean
  }
}>()
</script>

<template>
  <NuxtLink
    :to="`/restaurants/${restaurant.restaurant_id}`"
    class="bg-white rounded-xl shadow-sm hover:shadow-md transition-shadow overflow-hidden block"
  >
    <div class="h-36 bg-gray-100 flex items-center justify-center text-4xl">
      {{ restaurant.image_url ? '' : '🍽️' }}
      <img v-if="restaurant.image_url" :src="restaurant.image_url" class="w-full h-full object-cover" alt="" />
    </div>
    <div class="p-4">
      <div class="flex items-start justify-between">
        <h3 class="font-semibold text-gray-900 text-base">{{ restaurant.name }}</h3>
        <span class="text-sm text-yellow-500 font-medium ml-2">★ {{ restaurant.rating?.toFixed(1) ?? '—' }}</span>
      </div>
      <p class="text-xs text-gray-500 mt-1">{{ restaurant.category_name }}</p>
      <p class="text-sm text-gray-600 mt-1 line-clamp-2">{{ restaurant.description }}</p>
      <p class="text-xs text-gray-400 mt-2 truncate">📍 {{ restaurant.address }}</p>
    </div>
  </NuxtLink>
</template>
