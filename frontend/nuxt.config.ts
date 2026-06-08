export default defineNuxtConfig({
  ssr: true,
  buildDir: process.env.NUXT_BUILD_DIR?.trim() || '.nuxt',
  vite: {
    server: {
      allowedHosts: true,
    },
  },
  devtools: {
    enabled: true,
  },
  modules: ['@pinia/nuxt'],
  css: ['~/assets/styles/main.css'],
  app: {
    head: {
      htmlAttrs: {
        lang: 'zh-Hant',
      },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      ],
      link: [
        { rel: 'icon', href: '/favicon.ico' },
      ],
    },
  },
  nitro: {
    routeRules: {
      '/cart': {
        headers: {
          'x-robots-tag': 'noindex, nofollow',
        },
      },
      '/checkout': {
        headers: {
          'x-robots-tag': 'noindex, nofollow',
        },
      },
    },
  },
  runtimeConfig: {
    serverApiBase: process.env.NUXT_SERVER_API_BASE?.trim() || 'http://backend:8088',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE?.trim() || '',
      assetBase: process.env.NUXT_PUBLIC_ASSET_BASE?.trim() || '',
    },
  },
  compatibilityDate: '2026-05-18',
  typescript: {
    strict: true,
    typeCheck: false,
  },
})
