<script setup lang="ts">
import { createBreadcrumbJsonLd, useStoreSeo } from '~/composables/useStoreSeo'

const tenantStore = useTenantStore()
await tenantStore.initTenant()

const tenantName = tenantStore.currentTenant?.name ?? 'Vape Group 商城'

useStoreSeo({
  title: `店鋪說明 | ${tenantName}`,
  description: '瞭解多租戶商城前台的商品、購物車、訂單與店鋪內容能力。',
  canonicalPath: '/about',
  type: 'website',
  siteName: tenantName,
  locale: 'zh_TW',
  lang: 'zh-Hant',
  jsonLd: createBreadcrumbJsonLd({
    items: [
      { name: '首頁', path: '/' },
      { name: '店鋪說明', path: '/about' },
    ],
  }),
})
</script>

<template>
  <section class="about-page">
    <article class="panel lead-card">
      <p class="eyebrow">About the storefront</p>
      <h1>這個前台已經不只是展示頁，而是一個可演示的商城骨架。</h1>
      <p>
        現在前台除了首頁與列表，還補上了商品詳情、購物車、租戶店鋪資訊與 API 串接，並用 Nuxt 3 SSR 輸出關鍵落地頁。
      </p>
    </article>

    <div class="about-grid">
      <article class="panel">
        <h2>目前支援</h2>
        <ul>
          <li>支援不同域名映射不同店鋪內容與主題色</li>
          <li>首頁、商品目錄、商品詳情與 about 頁面採用 SSR</li>
          <li>前台可直接接後端 API 取得商品列表、詳情與庫存</li>
        </ul>
      </article>
      <article class="panel">
        <h2>這次補強</h2>
        <ul>
          <li>把關鍵 SEO 落地頁改成 Nuxt 3 SSR，而非只靠 SPA 注入 meta</li>
          <li>保留購物車與直接下單互動，避免商城只剩靜態展示</li>
          <li>為後續 sitemap、快取與部署收斂出一致的前台架構</li>
        </ul>
      </article>
    </div>
  </section>
</template>

<style scoped>
.about-page {
  display: grid;
  gap: 1rem;
}

.eyebrow {
  color: var(--tenant-accent, var(--wp-blue));
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: 0.75rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.lead-card h1,
.panel h2 {
  color: var(--wp-heading);
  margin-bottom: 0.75rem;
}

.lead-card p:last-child,
.panel li {
  color: var(--wp-text-muted);
}

.about-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, 1fr);
}

ul {
  padding-left: 1.2rem;
}

@media (max-width: 800px) {
  .about-grid {
    grid-template-columns: 1fr;
  }
}
</style>
