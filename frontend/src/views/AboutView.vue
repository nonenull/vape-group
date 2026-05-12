<script setup lang="ts">
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: '店鋪說明 | Vape Group 商城',
  description: '瞭解多租戶商城前台的商品、購物車、訂單與店鋪內容能力。',
  canonicalPath: '/about',
  type: 'website',
})
</script>

<template>
  <section class="about-page">
    <article class="panel lead-card">
      <p class="eyebrow">About the storefront</p>
      <h1>這個前台已經不只是展示頁，而是一個可演示的商城骨架。</h1>
      <p>
        現在前台除了首頁與列表，還補上了商品詳情、購物車、租戶店鋪資訊與 API fallback，已經更接近可演示的多租戶商城。
      </p>
    </article>

    <div class="about-grid">
      <article class="panel">
        <h2>目前支援</h2>
        <ul>
          <li>支援不同域名映射不同店鋪內容與主題色</li>
          <li>前台可直接接後端 API 取得商品列表、詳情與庫存</li>
          <li>已具備商品詳情頁、購物車與站點公告流程</li>
        </ul>
      </article>
      <article class="panel">
        <h2>這次補強</h2>
        <ul>
          <li>補齊商城的主購物流程，而不是只停留在靜態展示</li>
          <li>統一成接近 WooCommerce 的按鈕、卡片、詳情與購物車版型</li>
          <li>為真實商城 API 接入保留穩定的前端骨架</li>
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

.panel {
  background: var(--wp-surface);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1.5rem;
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
