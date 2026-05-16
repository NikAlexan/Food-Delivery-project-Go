<script setup lang="ts">
import { useApi } from '~/composables/useApi'

const props = defineProps<{ deliveryId: number }>()

const { apiFetch } = useApi()

interface TrackData {
  delivery_id: number
  status: string
  driver_name: string
  current_latitude: number
  current_longitude: number
  updated_at: string
}

const track = ref<TrackData | null>(null)
const error = ref('')
let map: any = null
let marker: any = null
let timer: ReturnType<typeof setInterval> | null = null

const statusLabel: Record<string, string> = {
  assigned: 'Курьер назначен',
  in_transit: 'В пути',
  completed: 'Доставлен',
  cancelled: 'Отменён',
}

async function fetchTrack() {
  try {
    track.value = await apiFetch<TrackData>(`/api/delivery/${props.deliveryId}/track`)
    updateMap()
  } catch {
    error.value = 'Данные о доставке недоступны'
  }
}

function updateMap() {
  if (!track.value || !import.meta.client) return
  const lat = track.value.current_latitude
  const lng = track.value.current_longitude
  if (!lat && !lng) return

  if (map && marker) {
    marker.setLatLng([lat, lng])
    map.panTo([lat, lng])
  }
}

async function initMap() {
  if (!import.meta.client) return
  const L = (await import('leaflet')).default

  const container = document.getElementById('delivery-map')
  if (!container || map) return

  const lat = track.value?.current_latitude ?? 51.18
  const lng = track.value?.current_longitude ?? 71.44

  map = L.map('delivery-map').setView([lat, lng], 14)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap',
  }).addTo(map)

  const icon = L.divIcon({
    html: '<div style="font-size:24px">🛵</div>',
    className: '',
    iconAnchor: [12, 12],
  })
  marker = L.marker([lat, lng], { icon }).addTo(map)
}

onMounted(async () => {
  await fetchTrack()
  await nextTick()
  await initMap()
  timer = setInterval(fetchTrack, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (map) map.remove()
})
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm overflow-hidden">
    <div class="p-4 border-b">
      <h3 class="font-semibold text-gray-900">Отслеживание доставки</h3>
      <div v-if="track" class="mt-2 flex flex-wrap gap-4 text-sm text-gray-600">
        <span>🚴 {{ track.driver_name || 'Курьер' }}</span>
        <span
          class="px-2 py-0.5 rounded-full text-xs font-medium"
          :class="{
            'bg-blue-100 text-blue-700': track.status === 'in_transit',
            'bg-green-100 text-green-700': track.status === 'completed',
            'bg-yellow-100 text-yellow-700': track.status === 'assigned',
          }"
        >{{ statusLabel[track.status] ?? track.status }}</span>
        <span v-if="track.current_latitude" class="text-gray-400 text-xs">
          {{ track.current_latitude.toFixed(4) }}, {{ track.current_longitude.toFixed(4) }}
        </span>
      </div>
      <p v-if="error" class="text-sm text-red-500 mt-1">{{ error }}</p>
    </div>
    <div id="delivery-map" class="w-full h-64" />
  </div>
</template>
