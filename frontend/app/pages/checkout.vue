<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useCartStore } from '~/stores/cart'
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const cart = useCartStore()
const auth = useAuthStore()
const { apiFetch } = useApi()

const addresses = ref<any[]>([])
const selectedAddress = ref('')
const payMethod = ref('card')
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  cart.init()
  auth.init()
  if (cart.items.length === 0) navigateTo('/cart')
  try {
    const data = await apiFetch<{ addresses: any[] }>('/api/users/addresses')
    addresses.value = data.addresses ?? []
    const def = addresses.value.find((a) => a.is_default)
    if (def) selectedAddress.value = `${def.street}, ${def.city}`
  } catch {}
})

const DELIVERY_FEE = 500

const subtotal = computed(() => cart.items.reduce((s, i) => s + i.price * i.quantity, 0))
const total = computed(() => subtotal.value + DELIVERY_FEE)

async function place() {
  if (!selectedAddress.value.trim()) {
    error.value = 'Укажите адрес доставки'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const order = await apiFetch<any>('/api/orders', {
      method: 'POST',
      body: JSON.stringify({
        restaurant_id: cart.restaurant_id,
        items: cart.items.map((i) => ({
          menu_item_id: i.menu_item_id,
          name: i.name,
          quantity: i.quantity,
          price: i.price,
        })),
        delivery_address: selectedAddress.value,
        user_email: auth.email ?? '',
      }),
    })

    await apiFetch(`/api/orders/${order.id}/payment`, {
      method: 'POST',
      body: JSON.stringify({ method: payMethod.value, amount: order.total }),
    })

    cart.clear()
    navigateTo(`/orders/${order.id}`)
  } catch (e: any) {
    error.value = e?.error ?? 'Не удалось оформить заказ'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto">
    <h1 class="text-xl font-bold text-gray-900 mb-6">Оформление заказа</h1>

    <div class="space-y-4">
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <h2 class="font-medium text-gray-900 mb-3">Адрес доставки</h2>
        <div v-if="addresses.length" class="space-y-2 mb-3">
          <label
            v-for="addr in addresses"
            :key="addr.address_id"
            class="flex items-center gap-2 cursor-pointer"
          >
            <input
              type="radio"
              :value="`${addr.street}, ${addr.city}`"
              v-model="selectedAddress"
              class="text-brand-500"
            />
            <span class="text-sm text-gray-700">{{ addr.street }}, {{ addr.city }}</span>
            <span v-if="addr.is_default" class="text-xs text-brand-600">по умолчанию</span>
          </label>
        </div>
        <input
          v-model="selectedAddress"
          placeholder="Или введите адрес вручную"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
        />
      </div>

      <div class="bg-white rounded-xl p-4 shadow-sm">
        <h2 class="font-medium text-gray-900 mb-3">Способ оплаты</h2>
        <div class="space-y-2">
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="radio" value="card" v-model="payMethod" class="text-brand-500" />
            <span class="text-sm text-gray-700">💳 Банковская карта</span>
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="radio" value="cash" v-model="payMethod" class="text-brand-500" />
            <span class="text-sm text-gray-700">💵 Наличными курьеру</span>
          </label>
        </div>
      </div>

      <div class="bg-white rounded-xl p-4 shadow-sm">
        <h2 class="font-medium text-gray-900 mb-3">Состав заказа</h2>
        <ul class="space-y-1 text-sm text-gray-600">
          <li v-for="item in cart.items" :key="item.menu_item_id" class="flex justify-between">
            <span>{{ item.name }} × {{ item.quantity }}</span>
            <span>{{ (item.price * item.quantity).toFixed(2) }} ₸</span>
          </li>
        </ul>
        <div class="mt-3 pt-2 flex justify-between text-sm text-gray-500">
          <span>Сумма заказа</span>
          <span>{{ subtotal.toFixed(2) }} ₸</span>
        </div>
        <div class="flex justify-between text-sm text-gray-500">
          <span>Доставка</span>
          <span>{{ DELIVERY_FEE.toFixed(2) }} ₸</span>
        </div>
        <div class="mt-2 pt-2 border-t flex justify-between font-semibold text-gray-900">
          <span>Итого</span>
          <span>{{ total.toFixed(2) }} ₸</span>
        </div>
      </div>

      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>

      <button
        @click="place"
        :disabled="loading"
        class="w-full bg-brand-500 hover:bg-brand-600 text-white font-medium py-3 rounded-xl transition-colors disabled:opacity-60"
      >
        {{ loading ? 'Оформляем...' : 'Подтвердить и оплатить' }}
      </button>
    </div>
  </div>
</template>
