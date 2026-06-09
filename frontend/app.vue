<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useCartStore } from '~/stores/cart'
import { useTenantStore } from '~/stores/tenant'

const tenantStore = useTenantStore()
const cartStore = useCartStore()

await tenantStore.initTenant()

const tenantName = computed(() => tenantStore.currentTenant?.name ?? 'Vape Group 商城')
const supportText = computed(() => tenantStore.currentTenant?.supportText ?? '支援訂單追蹤與商品諮詢')
const tickerText = computed(() => tenantStore.currentTenant?.announcement || '歡迎來到本站，最新活動與優惠資訊將在此更新。')
const tenantLogo = computed(() => tenantStore.currentTenant?.logoImage?.trim() || '/logo.svg')
const lineContactUrl = computed(() => tenantStore.platformConfig.lineContactUrl.trim())
const showCartShortcut = computed(() => cartStore.showCartShortcut && cartStore.itemCount > 0)
const headerKeyword = ref('')
const showBackToTop = ref(false)
const themeVars = computed(() => tenantStore.currentTenant ? tenantStore.getThemeVariables(tenantStore.currentTenant) : {})

useHead(() => ({
  link: [
    { rel: 'icon', href: tenantLogo.value || '/favicon.ico' },
  ],
}))

const syncBackToTopState = () => {
  if (!import.meta.client) {
    return
  }
  showBackToTop.value = window.scrollY > 280
}

const scrollToTop = () => {
  if (!import.meta.client) {
    return
  }
  window.scrollTo({
    top: 0,
    behavior: 'smooth',
  })
}

const submitHeaderSearch = async () => {
  const keyword = headerKeyword.value.trim()
  await navigateTo({
    path: '/products',
    query: keyword ? { keyword } : {},
  })
}

onMounted(async () => {
  syncBackToTopState()
  window.addEventListener('scroll', syncBackToTopState, { passive: true })
})

onUnmounted(() => {
  if (!import.meta.client) {
    return
  }
  window.removeEventListener('scroll', syncBackToTopState)
})
</script>

<template>
  <div class="store-shell" :style="themeVars">
    <header class="store-header">
      <div class="header-inner">
        <NuxtLink to="/" class="brand">
          <div class="brand-mark">
            <img :alt="tenantName" class="logo" :src="tenantLogo" width="42" height="42">
          </div>
          <div>
            <strong>{{ tenantName }}</strong>
            <p>{{ supportText }}</p>
          </div>
        </NuxtLink>

        <nav class="primary-nav" aria-label="Store Navigation">
          <NuxtLink to="/">首頁</NuxtLink>
          <NuxtLink to="/products">商品目錄</NuxtLink>
          <NuxtLink to="/cart">購物車<span v-if="cartStore.itemCount"> ({{ cartStore.itemCount }})</span></NuxtLink>
        </nav>

        <form class="header-search" role="search" @submit.prevent="submitHeaderSearch">
          <input
            v-model="headerKeyword"
            type="search"
            placeholder="搜尋商品、口味、型號"
            aria-label="搜尋商品"
          >
          <button type="submit" class="header-search-button">搜尋</button>
        </form>

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
      <NuxtPage />
    </main>

    <div class="floating-actions" aria-label="快捷操作">
      <Transition name="cart-shortcut">
        <NuxtLink
          v-if="showCartShortcut"
          to="/cart"
          class="floating-button floating-button-cart"
          aria-label="查看購物車"
          @click="cartStore.hideCartShortcut()"
        >
          <span class="floating-icon" aria-hidden="true">🛒</span>
          <span class="floating-copy">前往購物車</span>
        </NuxtLink>
      </Transition>
      <button
        v-show="showBackToTop"
        type="button"
        class="floating-button floating-button-top"
        aria-label="回到頂部"
        @click="scrollToTop"
      >
        <span class="floating-icon" aria-hidden="true">↑</span>
      </button>
      <a
        v-if="lineContactUrl"
        :href="lineContactUrl"
        class="floating-button floating-button-line"
        aria-label="客服 LINE"
        rel="noopener noreferrer"
        target="_blank"
      >
        <span class="floating-icon" aria-hidden="true">LINE</span>
      </a>
    </div>
  </div>
</template>

<style scoped>
.floating-actions {
  position: fixed;
  right: 1rem;
  bottom: 1rem;
  z-index: 40;
  display: grid;
  gap: 0.75rem;
  justify-items: end;
}

.header-search {
  display: flex;
  flex: 1 1 320px;
  align-items: center;
  gap: 0.55rem;
  min-width: 0;
  width: min(100%, 420px);
}

.header-search input {
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
  min-height: 42px;
  padding: 0.7rem 0.85rem;
  border: 1px solid var(--tenant-border, var(--wp-border));
  border-radius: 999px;
  background: #ffffff;
  color: var(--tenant-text, var(--wp-heading));
}

.header-search-button {
  min-height: 42px;
  padding: 0.7rem 1rem;
  border: 0;
  border-radius: 999px;
  background: var(--tenant-accent, var(--wp-blue));
  color: #ffffff;
  cursor: pointer;
  white-space: nowrap;
}

.floating-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
  padding: 0.75rem 0.95rem;
  border-radius: 999px;
  border: 1px solid var(--tenant-border, var(--wp-border));
  background: #ffffff;
  color: var(--tenant-text, var(--wp-heading));
  text-decoration: none;
  box-shadow: var(--wp-shadow);
}

.floating-button-top {
  width: 46px;
  padding: 0;
}

.floating-button-line {
  width: 46px;
  min-width: 46px;
  min-height: 46px;
  padding: 0;
  background: #06c755;
  border-color: #06c755;
  color: #ffffff;
}

.floating-button-line .floating-icon {
  font-size: 0.72rem;
  letter-spacing: 0.02em;
}

.floating-button-cart {
  gap: 0.5rem;
  padding-inline: 0.95rem 1.1rem;
  background: var(--tenant-accent, var(--wp-blue));
  border-color: var(--tenant-accent, var(--wp-blue));
  color: #ffffff;
}

.floating-icon {
  line-height: 1;
}

.floating-copy {
  white-space: nowrap;
}

.cart-shortcut-enter-active,
.cart-shortcut-leave-active {
  transition: transform 0.22s ease, opacity 0.22s ease;
}

.cart-shortcut-enter-from,
.cart-shortcut-leave-to {
  transform: translateX(18px);
  opacity: 0;
}

@media (max-width: 960px) {
  .header-search {
    width: 100%;
    flex-basis: 100%;
    order: 3;
  }
}

@media (max-width: 640px) {
  .header-search {
    min-width: 100%;
  }

  .floating-actions {
    right: 0.75rem;
    bottom: 0.75rem;
  }

  .floating-button {
    min-height: 42px;
    padding: 0.65rem 0.85rem;
  }

  .floating-button-top {
    width: 42px;
    padding: 0;
  }

  .floating-button-line {
    width: 42px;
    min-width: 42px;
    min-height: 42px;
  }
}
</style>
