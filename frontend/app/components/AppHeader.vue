<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useCartStore } from '~/stores/cart'

const auth = useAuthStore()
const cart = useCartStore()

onMounted(() => {
  auth.init()
  cart.init()
})

function logout() {
  auth.logout()
  cart.clear()
  navigateTo('/auth/login')
}
</script>

<template>
  <header class="bg-white shadow-sm sticky top-0 z-50">
    <div class="container mx-auto px-4 h-14 flex items-center justify-between">
      <NuxtLink to="/restaurants" class="text-brand-600 font-bold text-xl tracking-tight">
        🍔 FoodDelivery
      </NuxtLink>

      <nav class="flex items-center gap-4 text-sm font-medium text-gray-600">
        <NuxtLink to="/restaurants" class="hover:text-brand-600 transition-colors">Рестораны</NuxtLink>

        <template v-if="auth.isLoggedIn">
          <template v-if="auth.role === 'driver'">
            <NuxtLink to="/driver" class="hover:text-brand-600 transition-colors">Мои доставки</NuxtLink>
          </template>

          <template v-else-if="auth.role === 'manager'">
            <NuxtLink to="/restaurant" class="hover:text-brand-600 transition-colors">Мой ресторан</NuxtLink>
          </template>

          <template v-else>
            <NuxtLink to="/orders" class="hover:text-brand-600 transition-colors">Заказы</NuxtLink>
            <NuxtLink to="/cart" class="relative hover:text-brand-600 transition-colors">
              🛒
              <span
                v-if="cart.count > 0"
                class="absolute -top-2 -right-2 bg-brand-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center"
              >{{ cart.count }}</span>
            </NuxtLink>
          </template>

          <NuxtLink to="/profile" class="hover:text-brand-600 transition-colors">Профиль</NuxtLink>
          <button @click="logout" class="text-gray-400 hover:text-red-500 transition-colors">Выйти</button>
        </template>

        <template v-else>
          <NuxtLink to="/auth/login" class="hover:text-brand-600 transition-colors">Войти</NuxtLink>
          <NuxtLink
            to="/auth/register"
            class="bg-brand-500 text-white px-3 py-1.5 rounded-lg hover:bg-brand-600 transition-colors"
          >Регистрация</NuxtLink>
        </template>
      </nav>
    </div>
  </header>
</template>
