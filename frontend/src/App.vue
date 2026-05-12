<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { useTenantStore } from '@/stores/tenant'

const productStore = useProductStore()
const tenantStore = useTenantStore()

const tenantName = computed(() => tenantStore.currentTenant?.name ?? 'Vape Group 商城')
const announcement = computed(() => tenantStore.currentTenant?.announcement ?? '多站點商品內容統一管理')
const supportText = computed(() => tenantStore.currentTenant?.supportText ?? '支援訂單追蹤與商品諮詢')
const tickerText = computed(() => announcement.value || '歡迎來到本站，最新活動與優惠資訊將在此更新。')

onMounted(() => {
  tenantStore.initTenant()
  productStore.fetchProducts()
})
</script>

<template>
  <div class="store-shell">
    <header class="store-header">
      <div class="header-inner">
        <RouterLink to="/" class="brand">
          <div class="brand-mark">
            <img alt="Vape Group" class="logo" src="@/assets/logo.svg" width="42" height="42" />
          </div>
          <div>
            <strong>{{ tenantName }}</strong>
            <p>{{ supportText }}</p>
          </div>
        </RouterLink>

        <nav class="primary-nav" aria-label="Store Navigation">
          <RouterLink to="/">首頁</RouterLink>
          <RouterLink to="/products">商品目錄</RouterLink>
          <RouterLink to="/cart">購物車</RouterLink>
          <RouterLink to="/about">店鋪說明</RouterLink>
        </nav>

        <div class="header-tools" aria-label="網站公告">
          <div class="ticker-window">
            <div class="ticker-track">
              <span>{{ tickerText }}</span>
              <span aria-hidden="true">{{ tickerText }}</span>
            </div>
          </div>
        </div>
      </div>
    </header>

    <main class="store-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.store-shell {
  min-height: 100vh;
}

.notice-bar {
  background: var(--wp-admin-dark);
  color: #f0f0f1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.notice-inner,
.header-inner,
.store-main {
  width: min(calc(100% - 32px), var(--content-width));
  margin: 0 auto;
}

.notice-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1.25rem;
  min-height: 2.5rem;
  color: #c3c4c7;
  font-size: 0.8125rem;
}

.store-header {
  position: sticky;
  top: 0;
  z-index: 20;
  background: rgba(246, 247, 247, 0.92);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--wp-border);
}

.header-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1.5rem;
  min-height: 5.5rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  color: var(--wp-heading);
  min-width: 240px;
}

.brand-mark {
  width: 3rem;
  height: 3rem;
  border-radius: 0.75rem;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #ffffff, #eaf3fb);
  border: 1px solid var(--wp-border);
  box-shadow: var(--wp-shadow);
}

.brand strong {
  display: block;
  font-size: 1.05rem;
}

.brand p {
  color: var(--wp-text-muted);
  font-size: 0.8125rem;
}

.primary-nav {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem;
  border: 1px solid var(--wp-border);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: var(--wp-shadow);
}

.primary-nav a {
  padding: 0.65rem 1rem;
  border-radius: 999px;
  color: var(--wp-admin-muted);
  font-weight: 600;
}

.primary-nav a.router-link-active {
  color: var(--tenant-accent, var(--wp-blue));
  background: var(--wp-blue-soft);
}

.header-tools {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: min(30vw, 420px);
  padding: 0.55rem 0.8rem;
  border: 1px solid var(--wp-border);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: var(--wp-shadow);
}

.ticker-window {
  position: relative;
  overflow: hidden;
  flex: 1;
  white-space: nowrap;
}

.ticker-track {
  display: inline-flex;
  align-items: center;
  gap: 2.5rem;
  min-width: max-content;
  color: var(--wp-text-muted);
  font-size: 0.82rem;
  animation: ticker-scroll 18s linear infinite;
}

.ticker-track span {
  display: inline-block;
}

@keyframes ticker-scroll {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(calc(-50% - 1.25rem));
  }
}

.store-main {
  padding: 2rem 0 3rem;
}

@media (max-width: 960px) {
  .notice-inner,
  .header-inner {
    flex-wrap: wrap;
    padding: 0.75rem 0;
  }

  .header-inner {
    justify-content: center;
  }

  .primary-nav {
    order: 3;
    width: 100%;
    justify-content: center;
    border-radius: 1rem;
  }

  .header-tools {
    width: 100%;
    justify-content: center;
    min-width: 0;
  }
}

@media (max-width: 640px) {
  .notice-inner {
    flex-wrap: wrap;
    justify-content: center;
  }

  .primary-nav {
    flex-wrap: wrap;
  }

  .primary-nav a {
    flex: 1 1 calc(50% - 0.5rem);
    text-align: center;
  }

  .header-tools {
    padding: 0.5rem 0.7rem;
  }

  .ticker-track {
    font-size: 0.78rem;
  }
}
</style>
