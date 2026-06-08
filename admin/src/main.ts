import './assets/main.css'
import 'element-plus/dist/index.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAdminStore } from './stores/admin'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const adminStore = useAdminStore(pinia)

router.beforeEach(async (to) => {
  if (!adminStore.authReady) {
    try {
      await adminStore.hydrateAuth()
    } catch {
      // Route guard will redirect unauthenticated users below.
    }
  }

  if (to.meta.requiresAuth !== false && !adminStore.isAuthenticated) {
    return {
      name: 'login',
      query: to.fullPath && to.fullPath !== '/' ? { redirect: to.fullPath } : {},
    }
  }

  if (to.name === 'login' && adminStore.isAuthenticated) {
    const redirect = typeof to.query.redirect === 'string' && to.query.redirect ? to.query.redirect : '/'
    return redirect
  }
})

app.mount('#app')
