<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import ProductCard from '@/components/ProductCard.vue'
import { createStoreJsonLd, useSeo } from '@/composables/useSeo'
import { useProductStore } from '@/stores/product'
import { useTenantStore } from '@/stores/tenant'

const productStore = useProductStore()
const tenantStore = useTenantStore()
const currentTenant = computed(() => tenantStore.currentTenant)
const featuredProducts = computed(() => productStore.products.slice(0, 5))
const productCategories = computed(() => {
  const names = [...new Set(productStore.products.map((product) => product.category).filter(Boolean))]
  return names.slice(0, 3).map((name) => ({
    name,
    description: `租戶站點中的 ${name} 商品分類，可直接從商品目錄進一步篩選。`,
  }))
})

useSeo(computed(() => {
  const tenant = currentTenant.value
  const title = tenant?.seoTitle || `${tenant?.name ?? 'Vape Group 商城'} | 精選電子煙商品與多租戶店鋪`
  const description =
    tenant?.seoDescription ||
    tenant?.tagline ||
    '瀏覽電子煙商品目錄、熱銷推薦與多租戶店鋪內容，快速進入商品詳情與購物流程。'

  return {
    title,
    description,
    image: tenant?.previewImage || tenant?.logoImage || featuredProducts.value[0]?.image,
    type: 'website' as const,
    canonicalPath: '/',
    jsonLd: createStoreJsonLd({
      name: tenant?.name ?? 'Vape Group 商城',
      description,
      url: window.location.origin,
    }),
  }
}))
</script>

<template>
  <section class="home-page">
    <section class="category-panel panel">
      <div class="section-heading">
        <div>
          <p class="section-label">Browse categories</p>
          <h2>熱門分類</h2>
        </div>
        <RouterLink class="link-button" to="/products">前往完整商品目錄</RouterLink>
      </div>
      <div class="category-grid">
        <article v-for="category in productCategories" :key="category.name" class="category-card">
          <span class="category-tag">Collection</span>
          <h3>{{ category.name }}</h3>
          <p>{{ category.description }}</p>
        </article>
      </div>
    </section>

    <section class="featured-panel panel">
      <div class="section-heading">
        <div>
          <p class="section-label">Top products</p>
          <h2>本季熱賣</h2>
        </div>
        <RouterLink class="link-button" to="/products">瀏覽全部商品</RouterLink>
      </div>

      <div class="product-grid">
        <ProductCard v-for="product in featuredProducts" :key="product.id" :product="product" />
      </div>
    </section>
  </section>
</template>

<style scoped>
.home-page {
  display: grid;
  gap: 1.5rem;
}

.panel {
  background: var(--wp-surface);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1.5rem;
}

.section-label,
.category-tag {
  font-weight: 700;
}

.section-label {
  color: var(--tenant-accent, var(--wp-blue));
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.category-grid,
.product-grid {
  display: grid;
  gap: 1rem;
}

.category-card {
  padding: 1rem;
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
  border: 1px solid var(--wp-border);
}

.category-grid {
  grid-template-columns: repeat(3, 1fr);
}

.product-grid {
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

@media (max-width: 1400px) {
  .product-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 1.25rem;
  gap: 1rem;
}

.link-button {
  color: var(--tenant-accent, var(--wp-blue));
  font-weight: 600;
}

.category-card h3 {
  margin: 0.75rem 0 0.5rem;
}

.category-card p {
  color: var(--wp-text-muted);
}

.category-tag {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 0.625rem;
  border-radius: 999px;
  background: var(--tenant-surface, var(--wp-blue-soft));
  color: var(--tenant-accent, var(--wp-blue));
  font-size: 0.75rem;
}

@media (max-width: 960px) {
  .category-grid {
    grid-template-columns: 1fr;
  }

  .product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
