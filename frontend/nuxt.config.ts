// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  ssr: false,

  modules: ['@pinia/nuxt', '@nuxtjs/tailwindcss'],

  css: [
    '~/assets/styles/variables.css',
    '~/assets/styles/global.css',
    'video.js/dist/video-js.css',
  ],

  devServer: {
    port: 3000,
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
  nitro: {
    devProxy: {
      '/api/**': 'http://localhost:8080',
    },
  },

  runtimeConfig: {
    public: {
      apiBase: '',
    },
  },

  typescript: {
    strict: true,
  },
})
