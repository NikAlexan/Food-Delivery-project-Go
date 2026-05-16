export default defineNuxtConfig({
  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt',
    '@nuxt/icon',
    '@vueuse/nuxt',
  ],

  routeRules: {
    '/api/**': { proxy: 'http://localhost:8080/api/**' },
  },

  vite: {
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  },

  runtimeConfig: {
    public: {
      apiBase: '/api',
    },
  },

  app: {
    head: {
      title: 'Food Delivery',
      meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }],
      link: [
        {
          rel: 'stylesheet',
          href: 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css',
        },
      ],
    },
  },

  compatibilityDate: '2024-11-01',
})
