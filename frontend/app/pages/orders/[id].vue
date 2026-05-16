<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useApi } from '~/composables/useApi'

const route = useRoute()
const { apiFetch } = useApi()

const order = ref<any>(null)
const delivery = ref<any>(null)
const loading = ref(true)

const statusLabel: Record<string, string> = {
  pending: 'Ожидает оплаты',
  paid: 'Оплачен',
  cancelled: 'Отменён',
  delivered: 'Доставлен',
}

const statusColor: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  paid: 'bg-blue-100 text-blue-700',
  cancelled: 'bg-red-100 text-red-600',
  delivered: 'bg-green-100 text-green-700',
}

onMounted(async () => {
  try {
    order.value = await apiFetch<any>(`/api/orders/${route.params.id}`)
    if (['paid', 'delivered'].includes(order.value?.status)) {
      try {
        const hist = await apiFetch<{ deliveries: any[] }>('/api/delivery/history')
        delivery.value = (hist.deliveries ?? []).find((d: any) => d.order_id === order.value.id) ?? null
      } catch {}
    }
  } finally {
    loading.value = false
  }
})

async function cancel() {
  try {
    order.value = await apiFetch<any>(`/api/orders/${route.params.id}/cancel`, { method: 'POST' })
  } catch (e: any) {
    alert(e?.error ?? 'Не удалось отменить заказ')
  }
}
</script>

<template>
  <div class="max-w-xl mx-auto">
    <NuxtLink to="/orders" class="text-sm text-brand-600 hover:underline">← Заказы</NuxtLink>

    <div v-if="loading" class="text-center py-12 text-gray-400 mt-4">Загрузка...</div>

    <template v-else-if="order">
      <div class="bg-white rounded-xl shadow-sm p-5 mt-4">
        <div class="flex items-center justify-between">
          <h1 class="text-lg font-bold text-gray-900">Заказ #{{ order.id }}</h1>
          <span
            class="text-xs px-2 py-1 rounded-full font-medium"
            :class="statusColor[order.status] ?? 'bg-gray-100 text-gray-600'"
          >{{ statusLabel[order.status] ?? order.status }}</span>
        </div>
        <p class="text-xs text-gray-400 mt-1">{{ new Date(order.created_at).toLocaleString('ru') }}</p>

        <div class="mt-4 space-y-2">
          <div
            v-for="item in order.items"
            :key="item.menu_item_id"
            class="flex justify-between text-sm text-gray-700"
          >
            <span>{{ item.name }} × {{ item.quantity }}</span>
            <span>{{ (item.price * item.quantity).toFixed(2) }} ₸</span>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t flex justify-between font-semibold text-gray-900">
          <span>Итого</span>
          <span>{{ order.total.toFixed(2) }} ₸</span>
        </div>

        <button
          v-if="order.status === 'pending'"
          @click="cancel"
          class="mt-4 text-sm text-red-500 hover:underline"
        >Отменить заказ</button>
      </div>

      <div v-if="delivery" class="mt-4">
        <ClientOnly>
          <DeliveryTracker :delivery-id="delivery.delivery_id" />
        </ClientOnly>
      </div>
    </template>
  </div>
</template>
