<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useSeo } from '@/composables/useSeo'
import ProductCard from '@/components/ProductCard.vue'
import { useProductStore } from '@/stores/product'

const store = useProductStore()
const keyword = ref('')
const category = ref('')
const sortBy = ref('default')

const sourceProducts = computed(() => store.products)
const categories = computed(() => [...new Set(sourceProducts.value.map((item) => item.category))])

const filteredProducts = computed(() => {
  const base = sourceProducts.value.filter((item) => {
    const name = 'name' in item ? item.name : ''
    const description = 'description' in item ? item.description : ''
    const matchesKeyword = keyword.value
      ? name.toLowerCase().includes(keyword.value.toLowerCase()) ||
        description.toLowerCase().includes(keyword.value.toLowerCase())
      : true
    const matchesCategory = category.value ? item.category === category.value : true
    return matchesKeyword && matchesCategory
  })

  if (sortBy.value === 'price-asc') {
    return [...base].sort((a, b) => (a.salePrice ?? a.price) - (b.salePrice ?? b.price))
  }
  if (sortBy.value === 'price-desc') {
    return [...base].sort((a, b) => (b.salePrice ?? b.price) - (a.salePrice ?? a.price))
  }
  if (sortBy.value === 'rating') {
    return [...base].sort((a, b) => b.rating - a.rating)
  }
  return base
})

onMounted(() => {
  store.fetchProducts()
})

useSeo(computed(() => {
  const categoryLabel = category.value ? `${category.value} 分類商品` : '全站商品目錄'
  const keywordLabel = keyword.value ? `，關鍵字：${keyword.value}` : ''

  return {
    title: `${categoryLabel} | Vape Group 商城`,
    description: `瀏覽 ${categoryLabel}${keywordLabel}，支援商品搜尋、分類篩選與價格排序。`,
    image: filteredProducts.value[0]?.image,
    type: 'website' as const,
    canonicalPath: '/products',
  }
}))
</script>

<template>
  <section class="product-list-page">
    <div class="page-intro">
      <p class="breadcrumb">首頁 / 商店 / 商品目錄</p>
      <div class="intro-row">
        <div>
          <h1>商品目錄</h1>
          <p>提供搜尋、分類與排序，並支援從商品卡直接進入詳情與加入購物車。</p>
        </div>
        <div class="intro-summary panel">
          <span>Showing</span>
          <strong>{{ filteredProducts.length }} of {{ sourceProducts.length }} products</strong>
        </div>
      </div>
    </div>

    <div class="toolbar panel">
      <div class="filters">
        <input v-model="keyword" placeholder="搜尋商品、口味、型號..." />
        <select v-model="category">
          <option value="">全部分類</option>
          <option v-for="item in categories" :key="item" :value="item">{{ item }}</option>
        </select>
        <select v-model="sortBy">
          <option value="default">預設排序</option>
          <option value="price-asc">價格由低到高</option>
          <option value="price-desc">價格由高到低</option>
          <option value="rating">評分優先</option>
        </select>
      </div>
      <p class="toolbar-note">商品目錄現在直接來自後端 API，搜尋、分類與排序都基於真實資料。</p>
    </div>

    <div v-if="store.loading" class="loading-state panel">商品資料載入中...</div>

    <div v-else class="product-grid">
      <ProductCard v-for="product in filteredProducts" :key="product.id" :product="product" />
    </div>

    <div v-if="!store.loading && filteredProducts.length === 0" class="empty-state">
      未找到符合條件的商品，請試著清空搜尋或切換分類。
    </div>
  </section>
</template>

<style scoped>
.product-list-page {
  display: grid;
  gap: 1.25rem;
}

.page-intro {
  display: grid;
  gap: 0.75rem;
}

.breadcrumb,
.toolbar-note {
  color: var(--wp-text-muted);
  font-size: 0.8125rem;
}

.intro-row {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: end;
}

.intro-row h1 {
  color: var(--wp-heading);
  font-size: clamp(1.85rem, 3vw, 2.4rem);
  margin-bottom: 0.5rem;
}

.intro-row p {
  color: var(--wp-text-muted);
  max-width: 66ch;
}

.panel {
  background: var(--wp-surface);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
}

.intro-summary {
  min-width: 260px;
  padding: 1rem 1.25rem;
}

.intro-summary span {
  display: block;
  color: var(--wp-text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.intro-summary strong {
  color: var(--wp-heading);
  font-size: 1.15rem;
}

.toolbar {
  display: grid;
  gap: 0.875rem;
  padding: 1rem 1.25rem;
}

.filters {
  display: flex;
  gap: 0.875rem;
  flex-wrap: wrap;
}

.filters input,
.filters select {
  min-width: 220px;
  min-height: 2.6rem;
  padding: 0.65rem 0.8rem;
  border-radius: 0.375rem;
  border: 1px solid var(--wp-border-strong);
  color: var(--wp-text);
  background: #fff;
}

.loading-state,
.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--wp-text-muted);
}

.product-grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.empty-state {
  border: 1px dashed var(--wp-border-strong);
  border-radius: 0.5rem;
  background: var(--wp-surface);
  box-shadow: var(--wp-shadow);
}

@media (max-width: 1400px) {
  .product-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

</style>
