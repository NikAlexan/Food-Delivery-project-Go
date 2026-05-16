<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const auth = useAuthStore()
const { apiFetch } = useApi()

const deliveries = ref<any[]>([])
const loading = ref(true)
const driverIdInput = ref('')

onMounted(async () => {
  auth.init()
  driverIdInput.value = auth.driverId ? String(auth.driverId) : ''
  if (auth.driverId) await load()
  else loading.value = false
})

async function load() {
  loading.value = true
  try {
    const data = await apiFetch<{ deliveries: any[] }>(`/api/delivery/drivers/${auth.driverId}`)
    deliveries.value = (data.deliveries ?? []).filter((d) => ['assigned', 'in_transit'].includes(d.status))
  } finally {
    loading.value = false
  }
}

function setDriver() {
  const id = Number(driverIdInput.value)
  if (!id) return
  auth.setRole('driver', id, null)
  load()
}

const statusLabel: Record<string, string> = {
  assigned: 'Назначен',
  in_transit: 'В пути',
  completed: 'Завершён',
  cancelled: 'Отменён',
}

const statusColor: Record<string, string> = {
  assigned: 'bg-yellow-100 text-yellow-700',
  in_transit: 'bg-blue-100 text-blue-700',
  completed: 'bg-green-100 text-green-700',
  cancelled: 'bg-red-100 text-red-600',
}
</script>

<template>
  <div>
    <h1 class="text-xl font-bold text-gray-900 mb-6">Мои доставки</h1>

    <div v-if="!auth.driverId" class="max-w-sm mx-auto bg-white rounded-xl shadow-sm p-6 space-y-4">
      <p class="text-sm text-gray-600">Введите ваш Driver ID для просмотра доставок</p>
      <input
        v-model="driverIdInput"
        type="number"
        placeholder="Driver ID"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
      />
      <button
        @click="setDriver"
        class="w-full bg-brand-500 hover:bg-brand-600 text-white font-medium py-2 rounded-lg transition-colors"
      >Продолжить</button>
    </div>

    <template v-else>
      <div v-if="loading" class="text-center py-12 text-gray-400">Загрузка...</div>

      <div v-else-if="deliveries.length === 0" class="text-center py-12 text-gray-400">
        <p class="text-4xl mb-3">🛵</p>
        <p>Нет активных доставок</p>
      </div>

      <div v-else class="space-y-3">
        <NuxtLink
          v-for="d in deliveries"
          :key="d.delivery_id"
          :to="`/driver/${d.delivery_id}`"
          class="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow block"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="font-medium text-gray-900">Заказ #{{ d.order_id }}</span>
            <span
              class="text-xs px-2 py-0.5 rounded-full font-medium"
              :class="statusColor[d.status] ?? 'bg-gray-100 text-gray-600'"
            >{{ statusLabel[d.status] ?? d.status }}</span>
          </div>
          <p class="text-sm text-gray-600">📍 {{ d.delivery_address }}</p>
          <p class="text-xs text-gray-400 mt-1">{{ d.user_email }}</p>
        </NuxtLink>
      </div>
    </template>
  </div>
</template>
