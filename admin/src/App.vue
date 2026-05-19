<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const router = useRouter()
const store = useAdminStore()

onMounted(() => {
  store.bootstrap()
})
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="sidebar-brand">
        <div class="brand-mark">
          <img alt="Vape Group Admin" class="logo" src="@/assets/logo.svg" width="34" height="34" />
        </div>
        <div>
          <strong>Vape Group</strong>
          <p>Woo-style admin</p>
        </div>
      </div>

      <nav class="sidebar-nav" aria-label="Admin Navigation">
        <RouterLink to="/">儀表板</RouterLink>
        <RouterLink to="/products">商品</RouterLink>
        <RouterLink to="/categories">分類</RouterLink>
        <RouterLink to="/brands">品牌</RouterLink>
        <RouterLink to="/tenants">租戶</RouterLink>
        <RouterLink to="/platform-settings">平台配置</RouterLink>
        <RouterLink to="/orders">訂單</RouterLink>
      </nav>

      <div class="sidebar-footer">
        <span>Store Ops</span>
        <p>統一管理商品、庫存、租戶與站點內容。</p>
      </div>
    </aside>

    <div class="admin-content">
      <header class="admin-topbar">
        <div>
          <p class="topbar-label">WordPress inspired workspace</p>
          <h1>管理後台</h1>
        </div>
        <div class="topbar-actions">
          <button class="ghost-button" type="button" @click="router.push('/categories')">管理分類</button>
          <button class="primary-button" type="button" @click="router.push('/products')">新增商品</button>
        </div>
      </header>

      <main class="admin-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: var(--sidebar-width) minmax(0, 1fr);
}

.admin-sidebar {
  background: linear-gradient(180deg, var(--wp-admin-dark) 0%, #2c3338 100%);
  color: #f0f0f1;
  padding: 1rem 0.875rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  border-right: 1px solid rgba(255, 255, 255, 0.08);
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.5rem;
}

.brand-mark {
  width: 2.75rem;
  height: 2.75rem;
  display: grid;
  place-items: center;
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.08);
}

.sidebar-brand strong {
  display: block;
}

.sidebar-brand p {
  color: #c3c4c7;
  font-size: 0.8125rem;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar-nav a {
  color: #c3c4c7;
  font-weight: 600;
  padding: 0.75rem 0.875rem;
  border-radius: 0.375rem;
}

.sidebar-nav a.router-link-active {
  background: rgba(34, 113, 177, 0.18);
  color: #fff;
}

.sidebar-footer {
  margin-top: auto;
  padding: 1rem;
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.04);
}

.sidebar-footer span {
  display: block;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #8c8f94;
  margin-bottom: 0.35rem;
}

.sidebar-footer p {
  color: #dcdcde;
}

.admin-content {
  min-width: 0;
}

.admin-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: #fff;
  border-bottom: 1px solid var(--wp-border);
}

.topbar-label {
  color: var(--wp-blue);
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.admin-topbar h1 {
  font-size: 1.5rem;
}

.topbar-actions {
  display: flex;
  gap: 0.75rem;
}

.ghost-button,
.primary-button {
  min-height: 2.5rem;
  padding: 0.65rem 0.95rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid var(--wp-border-strong);
  background: #fff;
  color: var(--wp-text);
}

.primary-button {
  background: var(--wp-blue);
  color: #fff;
  border-color: var(--wp-blue);
}

.admin-main {
  padding: 1.5rem;
}

@media (max-width: 880px) {
  .admin-shell {
    grid-template-columns: 1fr;
  }

  .admin-sidebar {
    border-right: 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .sidebar-nav {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .admin-topbar {
    flex-wrap: wrap;
  }
}
</style>
