<script setup lang="ts">
import { useCartStore } from '~/stores/cart'

const props = defineProps<{
  item: {
    item_id: number
    name: string
    description: string
    price: number
    category: string
    image_url: string
    is_available: boolean
  }
  restaurantId: number
  restaurantName: string
}>()

const cart = useCartStore()
const added = ref(false)
const conflict = ref(false)

function addToCart() {
  conflict.value = false
  const ok = cart.addItem({
    menu_item_id: props.item.item_id,
    name: props.item.name,
    quantity: 1,
    price: props.item.price,
    restaurant_id: props.restaurantId,
    restaurant_name: props.restaurantName,
  })
  if (!ok) {
    conflict.value = true
    return
  }
  added.value = true
  setTimeout(() => (added.value = false), 1500)
}

function replaceCart() {
  cart.clear()
  conflict.value = false
  addToCart()
}
</script>

<template>
  <div
    class="bg-white rounded-xl p-4 shadow-sm flex gap-4"
    :class="{ 'opacity-60': !item.is_available }"
  >
    <div class="w-20 h-20 flex-shrink-0 rounded-lg bg-gray-100 overflow-hidden flex items-center justify-center text-2xl">
      <img v-if="item.image_url" :src="item.image_url" class="w-full h-full object-cover" alt="" />
      <span v-else>🍴</span>
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex items-start justify-between gap-2">
        <div>
          <h4 class="font-medium text-gray-900">{{ item.name }}</h4>
          <p class="text-xs text-gray-500 mt-0.5">{{ item.category }}</p>
          <p class="text-sm text-gray-600 mt-1 line-clamp-2">{{ item.description }}</p>
        </div>
        <span class="font-semibold text-brand-600 whitespace-nowrap">{{ item.price.toFixed(2) }} ₸</span>
      </div>

      <div class="mt-3">
        <div v-if="conflict" class="text-xs text-red-500 mb-2">
          В корзине товары из другого ресторана.
          <button class="underline ml-1" @click="replaceCart">Заменить корзину</button>
        </div>
        <button
          :disabled="!item.is_available"
          @click="addToCart"
          class="text-sm px-4 py-1.5 rounded-lg transition-colors"
          :class="added
            ? 'bg-green-500 text-white'
            : 'bg-brand-500 text-white hover:bg-brand-600 disabled:bg-gray-200 disabled:text-gray-400'"
        >
          {{ !item.is_available ? 'Недоступно' : added ? '✓ Добавлено' : 'В корзину' }}
        </button>
      </div>
    </div>
  </div>
</template>
