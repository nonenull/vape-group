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
const lineContactUrl = computed(() => tenantStore.platformConfig.lineContactUrl.trim())
const showBackToTop = ref(false)
const themeVars = computed(() => tenantStore.currentTenant ? tenantStore.getThemeVariables(tenantStore.currentTenant) : {})

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
            <img :alt="tenantName" class="logo" src="/logo.svg" width="42" height="42">
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
