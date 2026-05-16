<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useCartStore } from '~/stores/cart'
import { useApi } from '~/composables/useApi'

const cart = useCartStore()
const { apiFetch } = useApi()

const totals = ref<{ subtotal: number; delivery_fee: number; total: number } | null>(null)
const loading = ref(false)

onMounted(async () => {
  cart.init()
  if (cart.items.length > 0) await calcTotal()
})

async function calcTotal() {
  loading.value = true
  try {
    totals.value = await apiFetch('/api/orders/calculate', {
      method: 'POST',
      body: JSON.stringify({
        items: cart.items.map((i) => ({
          menu_item_id: i.menu_item_id,
          name: i.name,
          quantity: i.quantity,
          price: i.price,
        })),
      }),
    })
  } catch {
    totals.value = null
  } finally {
    loading.value = false
  }
}

function remove(id: number) {
  cart.removeItem(id)
  if (cart.items.length > 0) calcTotal()
  else totals.value = null
}

function changeQty(id: number, delta: number) {
  const item = cart.items.find((i) => i.menu_item_id === id)
  if (!item) return
  cart.updateQty(id, item.quantity + delta)
  if (cart.items.length > 0) calcTotal()
  else totals.value = null
}
</script>

<template>
  <div class="max-w-xl mx-auto">
    <h1 class="text-xl font-bold text-gray-900 mb-6">Корзина</h1>

    <div v-if="cart.items.length === 0" class="text-center py-16 text-gray-400">
      <p class="text-4xl mb-3">🛒</p>
      <p>Корзина пуста</p>
      <NuxtLink to="/restaurants" class="mt-3 text-brand-600 hover:underline text-sm inline-block">Выбрать ресторан</NuxtLink>
    </div>

    <template v-else>
      <p class="text-sm text-gray-500 mb-4">{{ cart.restaurant_name }}</p>

      <div class="space-y-3 mb-6">
        <div
          v-for="item in cart.items"
          :key="item.menu_item_id"
          class="bg-white rounded-xl p-4 shadow-sm flex items-center gap-4"
        >
          <div class="flex-1">
            <p class="font-medium text-gray-900 text-sm">{{ item.name }}</p>
            <p class="text-brand-600 text-sm mt-0.5">{{ (item.price * item.quantity).toFixed(2) }} ₸</p>
          </div>
          <div class="flex items-center gap-2">
            <button @click="changeQty(item.menu_item_id, -1)" class="w-7 h-7 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center text-lg leading-none">−</button>
            <span class="w-5 text-center text-sm font-medium">{{ item.quantity }}</span>
            <button @click="changeQty(item.menu_item_id, 1)" class="w-7 h-7 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center text-lg leading-none">+</button>
          </div>
          <button @click="remove(item.menu_item_id)" class="text-gray-300 hover:text-red-400 transition-colors ml-2">✕</button>
        </div>
      </div>

      <div class="bg-white rounded-xl p-4 shadow-sm mb-4 space-y-2">
        <div v-if="loading" class="text-sm text-gray-400">Пересчёт...</div>
        <template v-else-if="totals">
          <div class="flex justify-between text-sm text-gray-600">
            <span>Сумма</span><span>{{ totals.subtotal.toFixed(2) }} ₸</span>
          </div>
          <div class="flex justify-between text-sm text-gray-600">
            <span>Доставка</span><span>{{ totals.delivery_fee.toFixed(2) }} ₸</span>
          </div>
          <div class="flex justify-between font-semibold text-gray-900 pt-2 border-t">
            <span>Итого</span><span>{{ totals.total.toFixed(2) }} ₸</span>
          </div>
        </template>
      </div>

      <NuxtLink
        to="/checkout"
        class="block w-full bg-brand-500 hover:bg-brand-600 text-white text-center font-medium py-3 rounded-xl transition-colors"
      >
        Оформить заказ
      </NuxtLink>
    </template>
  </div>
</template>
