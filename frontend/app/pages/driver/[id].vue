<script setup lang="ts">
definePageMeta({ middleware: ['auth'] })

import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const route = useRoute()
const auth = useAuthStore()
const { apiFetch } = useApi()

const delivery = ref<any>(null)
const loading = ref(true)
const locating = ref(false)
const completing = ref(false)
const locMsg = ref('')
let map: any = null
let marker: any = null

onMounted(async () => {
  auth.init()
  await loadDelivery()
  await nextTick()
  await initMap()
})

async function loadDelivery() {
  loading.value = true
  try {
    delivery.value = await apiFetch<any>(`/api/delivery/${route.params.id}`)
  } finally {
    loading.value = false
  }
}

async function initMap() {
  if (!import.meta.client || !delivery.value) return
  const L = (await import('leaflet')).default
  const container = document.getElementById('driver-map')
  if (!container || map) return

  map = L.map('driver-map').setView([43.238, 76.889], 13)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap',
  }).addTo(map)

  const lat = delivery.value.current_latitude
  const lng = delivery.value.current_longitude
  if (lat && lng) {
    const icon = L.divIcon({ html: '<div style="font-size:24px">🛵</div>', className: '', iconAnchor: [12, 12] })
    marker = L.marker([lat, lng], { icon }).addTo(map)
    map.setView([lat, lng], 14)
  }

  L.marker([43.238, 76.889]).bindPopup(delivery.value.delivery_address).addTo(map)
}

async function updateLocation() {
  locating.value = true
  locMsg.value = ''
  navigator.geolocation.getCurrentPosition(
    async (pos) => {
      try {
        await apiFetch('/api/delivery/location', {
          method: 'PATCH',
          body: JSON.stringify({
            driver_id: auth.driverId,
            latitude: pos.coords.latitude,
            longitude: pos.coords.longitude,
          }),
        })
        locMsg.value = `✓ Локация обновлена: ${pos.coords.latitude.toFixed(4)}, ${pos.coords.longitude.toFixed(4)}`
        if (marker && map) {
          marker.setLatLng([pos.coords.latitude, pos.coords.longitude])
          map.panTo([pos.coords.latitude, pos.coords.longitude])
        }
      } catch {
        locMsg.value = 'Ошибка обновления локации'
      } finally {
        locating.value = false
      }
    },
    () => {
      locMsg.value = 'Геолокация недоступна'
      locating.value = false
    }
  )
}

async function complete() {
  completing.value = true
  try {
    delivery.value = await apiFetch<any>(`/api/delivery/${route.params.id}/complete`, { method: 'POST' })
  } catch (e: any) {
    alert(e?.error ?? 'Ошибка завершения доставки')
  } finally {
    completing.value = false
  }
}

const statusLabel: Record<string, string> = {
  assigned: 'Назначен',
  in_transit: 'В пути',
  completed: 'Завершён',
  cancelled: 'Отменён',
}
</script>

<template>
  <div class="max-w-xl mx-auto">
    <NuxtLink to="/driver" class="text-sm text-brand-600 hover:underline">← Все доставки</NuxtLink>

    <div v-if="loading" class="text-center py-12 text-gray-400 mt-4">Загрузка...</div>

    <template v-else-if="delivery">
      <div class="bg-white rounded-xl shadow-sm p-5 mt-4 space-y-3">
        <div class="flex items-center justify-between">
          <h1 class="text-lg font-bold text-gray-900">Доставка #{{ delivery.delivery_id }}</h1>
          <span class="text-sm text-gray-500">Заказ #{{ delivery.order_id }}</span>
        </div>

        <div class="space-y-1 text-sm text-gray-700">
          <p>📍 <span class="font-medium">{{ delivery.delivery_address }}</span></p>
          <p>👤 {{ delivery.user_email }}</p>
          <p>
            Статус:
            <span class="font-medium">{{ statusLabel[delivery.status] ?? delivery.status }}</span>
          </p>
        </div>

        <div v-if="delivery.status !== 'completed' && delivery.status !== 'cancelled'" class="space-y-2 pt-2">
          <button
            @click="updateLocation"
            :disabled="locating"
            class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 rounded-lg transition-colors disabled:opacity-60"
          >
            {{ locating ? 'Определяем локацию...' : '📡 Обновить мою локацию' }}
          </button>
          <p v-if="locMsg" class="text-xs text-center" :class="locMsg.startsWith('✓') ? 'text-green-600' : 'text-red-500'">{{ locMsg }}</p>

          <button
            @click="complete"
            :disabled="completing"
            class="w-full bg-green-500 hover:bg-green-600 text-white font-medium py-2 rounded-lg transition-colors disabled:opacity-60"
          >
            {{ completing ? 'Завершаем...' : '✅ Завершить доставку' }}
          </button>
        </div>

        <div v-else class="pt-2 text-center text-sm text-green-600 font-medium">
          Доставка завершена
        </div>
      </div>

      <div class="mt-4 bg-white rounded-xl shadow-sm overflow-hidden">
        <div class="p-3 border-b text-sm font-medium text-gray-700">Карта</div>
        <ClientOnly>
          <div id="driver-map" class="w-full h-64" />
        </ClientOnly>
      </div>
    </template>
  </div>
</template>
