<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
auth.init()

if (auth.role === 'manager') {
  navigateTo('/restaurant')
} else if (auth.role === 'driver') {
  navigateTo('/driver')
} else {
  navigateTo('/restaurants')
}
</script>
<template><div /></template>
