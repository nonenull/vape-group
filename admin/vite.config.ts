import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig({
  base: process.env.VITE_APP_BASE?.trim() || '/fuck/',
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
    vueDevTools(),
  ],
  server: {
    host: '0.0.0.0',
    allowedHosts: true,
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return
          }
          if (id.includes('@element-plus/icons-vue')) {
            return 'element-plus-icons'
          }
          if (id.includes('element-plus')) {
            if (id.includes('/components/table') || id.includes('/components/table-column')) {
              return 'element-plus-table'
            }
            if (id.includes('/components/form') || id.includes('/components/input') || id.includes('/components/select') || id.includes('/components/input-number') || id.includes('/components/switch') || id.includes('/components/date-picker')) {
              return 'element-plus-form'
            }
            if (id.includes('/components/dialog') || id.includes('/components/message-box') || id.includes('/components/overlay') || id.includes('/components/focus-trap')) {
              return 'element-plus-dialog'
            }
            if (id.includes('/components/button') || id.includes('/components/tag') || id.includes('/components/icon') || id.includes('/components/card')) {
              return 'element-plus-display'
            }
            return 'element-plus-core'
          }
          if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router')) {
            return 'vue-vendor'
          }
        },
      },
    },
  },
})
