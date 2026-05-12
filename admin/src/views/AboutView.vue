<template>
  <section class="about-page">
    <article class="panel">
      <p class="eyebrow">About the admin</p>
      <h2>這是統一商品與租戶內容的後台工作台。</h2>
      <p>
        目前已補上租戶中心、全域商品編輯、租戶商品覆寫與訂單追蹤，讓 admin 更接近 `product.md` 定義的多租戶管理後台。
      </p>
    </article>

    <div class="about-grid">
      <article class="panel">
        <h3>当前能力</h3>
        <ul>
          <li>集中管理價格、庫存、圖片等關鍵商品資訊</li>
          <li>支援不同租戶網域的差異化內容對應</li>
          <li>提供租戶層級的 SEO 與商品覆寫入口</li>
        </ul>
      </article>
      <article class="panel">
        <h3>改版方向</h3>
        <ul>
          <li>深色左側導覽 + 白色內容區</li>
          <li>統計卡、表格、編輯表單統一為 Woo 風格</li>
          <li>減少「展示頁」感，強化可操作的工作流</li>
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
  padding: 1.25rem;
}

.eyebrow {
  color: var(--wp-blue);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: 0.75rem;
  font-weight: 700;
  margin-bottom: 0.45rem;
}

.panel h2,
.panel h3 {
  margin-bottom: 0.75rem;
}

.panel p,
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

@media (max-width: 820px) {
  .about-grid {
    grid-template-columns: 1fr;
  }
}
</style>
