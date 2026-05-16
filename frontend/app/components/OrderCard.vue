<script setup lang="ts">
defineProps<{
  order: {
    id: number
    restaurant_id: number
    status: string
    total: number
    created_at: string
    items: { name: string; quantity: number; price: number }[]
  }
}>()

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
</script>

<template>
  <NuxtLink
    :to="`/orders/${order.id}`"
    class="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow block"
  >
    <div class="flex items-center justify-between">
      <span class="text-sm font-medium text-gray-900">Заказ #{{ order.id }}</span>
      <span
        class="text-xs px-2 py-0.5 rounded-full font-medium"
        :class="statusColor[order.status] ?? 'bg-gray-100 text-gray-600'"
      >{{ statusLabel[order.status] ?? order.status }}</span>
    </div>
    <p class="text-xs text-gray-400 mt-1">{{ new Date(order.created_at).toLocaleString('ru') }}</p>
    <ul class="mt-2 space-y-0.5">
      <li v-for="item in order.items" :key="item.name" class="text-sm text-gray-600">
        {{ item.name }} × {{ item.quantity }}
      </li>
    </ul>
    <div class="mt-3 flex items-center justify-between">
      <span class="text-sm text-gray-500">Итого</span>
      <span class="font-semibold text-gray-900">{{ order.total.toFixed(2) }} ₸</span>
    </div>
  </NuxtLink>
</template>
