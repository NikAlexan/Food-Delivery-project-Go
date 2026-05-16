<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useApi } from '~/composables/useApi'

const { apiFetch } = useApi()
const orders = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const data = await apiFetch<{ orders: any[] }>('/api/orders')
    orders.value = (data.orders ?? []).filter((o) => ['pending', 'paid'].includes(o.status))
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-bold text-gray-900">Активные заказы</h1>
      <NuxtLink to="/orders/history" class="text-sm text-brand-600 hover:underline">История →</NuxtLink>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Загрузка...</div>
    <div v-else-if="orders.length === 0" class="text-center py-12 text-gray-400">
      <p class="text-4xl mb-3">📦</p>
      <p>Нет активных заказов</p>
      <NuxtLink to="/restaurants" class="mt-3 text-brand-600 hover:underline text-sm inline-block">Сделать заказ</NuxtLink>
    </div>
    <div v-else class="space-y-3">
      <OrderCard v-for="o in orders" :key="o.id" :order="o" />
    </div>
  </div>
</template>
