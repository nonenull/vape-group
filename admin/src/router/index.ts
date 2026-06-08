import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: {
        requiresAuth: false,
      },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('../views/HomeView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/products',
      name: 'products',
      component: () => import('../views/ProductsView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/categories',
      name: 'categories',
      component: () => import('../views/CategoriesView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/brands',
      name: 'brands',
      component: () => import('../views/BrandsView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/domains',
      name: 'domains',
      component: () => import('../views/DomainsView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/tenants',
      name: 'tenants',
      component: () => import('../views/TenantsView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/platform-settings',
      name: 'platform-settings',
      component: () => import('../views/PlatformSettingsView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/orders',
      name: 'orders',
      component: () => import('../views/OrdersView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
      meta: {
        requiresAuth: true,
      },
    },
  ],
})

export default router
