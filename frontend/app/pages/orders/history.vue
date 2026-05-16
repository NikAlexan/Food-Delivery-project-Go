<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useApi } from '~/composables/useApi'

const { apiFetch } = useApi()
const orders = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const data = await apiFetch<{ orders: any[] }>('/api/orders/history')
    orders.value = data.orders ?? []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="flex items-center gap-3 mb-6">
      <NuxtLink to="/orders" class="text-sm text-brand-600 hover:underline">← Активные</NuxtLink>
      <h1 class="text-xl font-bold text-gray-900">История заказов</h1>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Загрузка...</div>
    <div v-else-if="orders.length === 0" class="text-center py-12 text-gray-400">История пуста</div>
    <div v-else class="space-y-3">
      <OrderCard v-for="o in orders" :key="o.id" :order="o" />
    </div>
  </div>
</template>
