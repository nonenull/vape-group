<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import ProductCard from '~/components/store/ProductCard.vue'
import { createBreadcrumbJsonLd, createItemListJsonLd, useStoreSeo } from '~/composables/useStoreSeo'
import { fetchBrands, fetchCategories, fetchProducts } from '~/composables/useStoreApi'
import { buildCategoryPath, buildProductPath } from '~/composables/useProductSlug'
import type { Product } from '~/types/store'

const props = withDefaults(defineProps<{
  initialCategoryId?: number | null
  redirectLegacyQuery?: boolean
}>(), {
  initialCategoryId: null,
  redirectLegacyQuery: false,
})

const tenantStore = useTenantStore()
await tenantStore.initTenant()

const route = useRoute()
const router = useRouter()

const initialKeyword = typeof route.query.keyword === 'string' ? route.query.keyword : ''
const initialCategory = typeof route.query.category === 'string'
  ? route.query.category
  : props.initialCategoryId != null
    ? String(props.initialCategoryId)
    : ''
const initialBrand = typeof route.query.brand === 'string' ? route.query.brand : ''
const initialSort = typeof route.query.sort === 'string' ? route.query.sort : 'default'
const initialPage = typeof route.query.page === 'string'
  ? Math.max(Number.parseInt(route.query.page, 10) || 1, 1)
  : 1
const initialLimit = typeof route.query.limit === 'string'
  ? Math.max(Number.parseInt(route.query.limit, 10) || 20, 1)
  : 20

const [initialProductResponse, categories, brands] = await Promise.all([
  fetchProducts(initialPage, initialLimit, {
    keyword: initialKeyword,
    category: initialCategory,
    brand: initialBrand,
    sort: initialSort,
  }),
  fetchCategories(),
  fetchBrands(),
])

const keyword = ref(initialKeyword)
const category = ref(initialCategory)
const brand = ref(initialBrand)
const sortBy = ref(initialSort)
const currentPage = ref(initialPage)
const pageSize = ref(initialLimit)
const products = ref<Product[]>(initialProductResponse.products)
const totalProducts = ref(initialProductResponse.total)
const tenantName = computed(() => tenantStore.currentTenant?.name ?? 'Vape Group 商城')
const isApplyingHistoryState = ref(false)

const categoryProductCount = computed(() => {
  const counts = new Map<number, number>()
  for (const product of products.value) {
    const productCategoryIds = product.categoryIds?.length
      ? product.categoryIds
      : product.categoryId != null
        ? [product.categoryId]
        : []
    for (const categoryId of productCategoryIds) {
      counts.set(categoryId, (counts.get(categoryId) ?? 0) + 1)
    }
  }
  return counts
})

const categoryDescendantIds = computed(() => {
  const byParent = new Map<number | null, number[]>()
  for (const item of categories) {
    const key = item.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(item.id)
    byParent.set(key, bucket)
  }

  const result = new Map<number, Set<number>>()
  const collect = (categoryId: number): Set<number> => {
    const cached = result.get(categoryId)
    if (cached) {
      return cached
    }

    const values = new Set<number>([categoryId])
    for (const childId of byParent.get(categoryId) ?? []) {
      for (const nestedId of collect(childId)) {
        values.add(nestedId)
      }
    }
    result.set(categoryId, values)
    return values
  }

  for (const item of categories) {
    collect(item.id)
  }

  return result
})

const selectedCategoryId = computed(() => {
  const value = Number(category.value)
  return Number.isFinite(value) && value > 0 ? value : null
})

const selectedCategoryRecord = computed(() => {
  if (!selectedCategoryId.value) {
    return null
  }
  return categories.find((item) => item.id === selectedCategoryId.value) ?? null
})

const selectedCategoryName = computed(() => selectedCategoryRecord.value?.name ?? '')

const categoryOptions = computed(() => {
  const byParent = new Map<number | null, typeof categories[number][]>()
  const sortedCategories = [...categories].sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id)

  for (const item of sortedCategories) {
    const key = item.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(item)
    byParent.set(key, bucket)
  }

  const options: Array<{ id: number; label: string }> = []
  const descendantCounts = categoryDescendantIds.value
  const walk = (parentId: number | null, depth: number, ancestors: string[]) => {
    const items = byParent.get(parentId) ?? []
    for (const item of items) {
      const path = [...ancestors, item.name]
      const descendantCount = Array.from(descendantCounts.get(item.id) ?? [])
        .reduce((sum, id) => sum + (categoryProductCount.value.get(id) ?? 0), 0)
      const prefix = depth === 0 ? '● ' : `${'　'.repeat(Math.max(depth - 1, 0))}└ `
      options.push({
        id: item.id,
        label: `${prefix}${path.join(' / ')} (${descendantCount})`,
      })
      walk(item.id, depth + 1, path)
    }
  }

  walk(null, 0, [])
  return options
})

const brandOptions = computed(() => {
  return brands
    .map((item) => item.name)
    .filter(Boolean)
    .sort((a, b) => a.localeCompare(b, 'zh-Hant'))
})

const filteredProducts = computed(() => products.value)
const totalPages = computed(() => Math.max(Math.ceil(totalProducts.value / pageSize.value), 1))
const paginatedProducts = computed(() => products.value)
const pageStart = computed(() => totalProducts.value === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1)
const pageEnd = computed(() => Math.min(currentPage.value * pageSize.value, totalProducts.value))
const visiblePageNumbers = computed(() => {
  const total = totalPages.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, index) => index + 1)
  }

  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(total, currentPage.value + 2)
  const pages = new Set<number>([1, total])
  for (let page = start; page <= end; page += 1) {
    pages.add(page)
  }

  return [...pages].sort((a, b) => a - b)
})

const syncQuery = async (mode: 'push' | 'replace' = 'push') => {
  const query = {
    ...(keyword.value ? { keyword: keyword.value } : {}),
    ...(brand.value ? { brand: brand.value } : {}),
    ...(sortBy.value !== 'default' ? { sort: sortBy.value } : {}),
    ...(currentPage.value > 1 ? { page: String(currentPage.value) } : {}),
    ...(pageSize.value !== 20 ? { limit: String(pageSize.value) } : {}),
  }

  if (selectedCategoryRecord.value) {
    await router[mode]({
      path: buildCategoryPath(selectedCategoryRecord.value),
      query,
    })
    return
  }

  await router[mode]({
    path: '/products',
    query,
  })
}

async function reloadProducts() {
  const response = await fetchProducts(currentPage.value, pageSize.value, {
    keyword: keyword.value,
    category: category.value,
    brand: brand.value,
    sort: sortBy.value,
  })
  products.value = response.products
  totalProducts.value = response.total
}

watch([keyword, category, brand, sortBy], () => {
  currentPage.value = 1
})

watch([keyword, category, brand, sortBy, currentPage, pageSize], async () => {
  await reloadProducts()
  if (import.meta.client) {
    await syncQuery(isApplyingHistoryState.value ? 'replace' : 'push')
  }
}, { immediate: false })

watch(totalPages, (value) => {
  if (currentPage.value > value) {
    currentPage.value = value
  }
})

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) {
    return
  }
  currentPage.value = page
  if (import.meta.client) {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

watch(
  () => route.fullPath,
  async () => {
    if (!import.meta.client) {
      return
    }

    const nextKeyword = typeof route.query.keyword === 'string' ? route.query.keyword : ''
    const nextCategory = typeof route.query.category === 'string'
      ? route.query.category
      : props.initialCategoryId != null && route.path.startsWith('/products/category/')
        ? String(props.initialCategoryId)
        : ''
    const nextBrand = typeof route.query.brand === 'string' ? route.query.brand : ''
    const nextSort = typeof route.query.sort === 'string' ? route.query.sort : 'default'
    const nextPage = typeof route.query.page === 'string'
      ? Math.max(Number.parseInt(route.query.page, 10) || 1, 1)
      : 1
    const nextLimit = typeof route.query.limit === 'string'
      ? Math.max(Number.parseInt(route.query.limit, 10) || 20, 1)
      : 20

    if (
      nextKeyword === keyword.value &&
      nextCategory === category.value &&
      nextBrand === brand.value &&
      nextSort === sortBy.value &&
      nextPage === currentPage.value &&
      nextLimit === pageSize.value
    ) {
      return
    }

    isApplyingHistoryState.value = true
    keyword.value = nextKeyword
    category.value = nextCategory
    brand.value = nextBrand
    sortBy.value = nextSort
    currentPage.value = nextPage
    pageSize.value = nextLimit
    await reloadProducts()
    isApplyingHistoryState.value = false
  },
)

const categoryLabel = computed(() => selectedCategoryName.value ? `${selectedCategoryName.value} 分類商品` : '全站商品目錄')
const brandLabel = computed(() => brand.value ? `${brand.value} 品牌商品` : '')
const seoTitle = computed(() => `${brandLabel.value || categoryLabel.value} | ${tenantName.value}`)
const description = computed(() => {
  const keywordLabel = keyword.value ? `，搜尋關鍵字為「${keyword.value}」` : ''
  const categorySegment = selectedCategoryName.value ? `分類鎖定 ${selectedCategoryName.value}` : ''
  const brandSegment = brand.value ? `品牌為 ${brand.value}` : ''
  const sortSegment = sortBy.value === 'price-asc'
    ? '目前依價格由低到高排序'
    : sortBy.value === 'price-desc'
      ? '目前依價格由高到低排序'
      : sortBy.value === 'rating'
        ? '目前依評分排序'
        : ''

  const segments = [
    `${tenantName.value} 商品目錄目前顯示 ${totalProducts.value} 款商品`,
    categorySegment,
    brandSegment,
    sortSegment,
  ].filter(Boolean)

  return `${segments.join('，')}${keywordLabel}。`
})

useStoreSeo({
  title: seoTitle.value,
  description: description.value,
  image: filteredProducts.value[0]?.image,
  type: 'website',
  canonicalPath: selectedCategoryRecord.value ? buildCategoryPath(selectedCategoryRecord.value) : '/products',
  siteName: tenantName.value,
  locale: 'zh_TW',
  lang: 'zh-Hant',
  jsonLd: [
    createBreadcrumbJsonLd({
      items: [
        { name: '首頁', path: '/' },
        { name: '商品目錄', path: '/products' },
        ...(selectedCategoryRecord.value ? [{ name: selectedCategoryRecord.value.name, path: buildCategoryPath(selectedCategoryRecord.value) }] : []),
      ],
    }),
    createItemListJsonLd({
      name: `${tenantName.value} 商品目錄`,
      description: description.value,
      items: paginatedProducts.value.slice(0, 12).map((product) => ({
        name: product.name,
        path: buildProductPath(product),
        image: product.image,
      })),
    }),
  ],
})

if (props.redirectLegacyQuery && typeof route.query.category === 'string' && selectedCategoryRecord.value) {
  await navigateTo({
    path: buildCategoryPath(selectedCategoryRecord.value),
    query: {
      ...(keyword.value ? { keyword: keyword.value } : {}),
      ...(brand.value ? { brand: brand.value } : {}),
      ...(sortBy.value !== 'default' ? { sort: sortBy.value } : {}),
      ...(currentPage.value > 1 ? { page: String(currentPage.value) } : {}),
      ...(pageSize.value !== 20 ? { limit: String(pageSize.value) } : {}),
    },
  }, { redirectCode: 301, replace: true })
}
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
          <strong>{{ pageStart }}-{{ pageEnd }} / {{ totalProducts }} 件商品</strong>
        </div>
      </div>
    </div>

    <div class="toolbar panel">
      <div class="filters">
        <input v-model="keyword" placeholder="搜尋商品、口味、型號...">
        <select v-model="category">
          <option value="">全部分類</option>
          <option v-for="item in categoryOptions" :key="item.id" :value="String(item.id)">{{ item.label }}</option>
        </select>
        <select v-model="brand">
          <option value="">全部品牌</option>
          <option v-for="item in brandOptions" :key="item" :value="item">{{ item }}</option>
        </select>
        <select v-model="sortBy">
          <option value="default">預設排序</option>
          <option value="price-asc">價格由低到高</option>
          <option value="price-desc">價格由高到低</option>
          <option value="rating">評分優先</option>
        </select>
        <select v-model.number="pageSize">
          <option :value="20">每頁 20 筆</option>
          <option :value="40">每頁 40 筆</option>
          <option :value="60">每頁 60 筆</option>
        </select>
      </div>
    </div>

    <div class="product-grid">
      <ProductCard
        v-for="product in paginatedProducts"
        :key="product.id"
        :product="product"
        :show-detail-button="false"
      />
    </div>

    <div v-if="filteredProducts.length === 0" class="empty-state panel">
      未找到符合條件的商品，請試著清空搜尋或切換分類。
    </div>

    <div v-else-if="totalPages > 1" class="pagination panel">
      <button class="page-button" type="button" :disabled="currentPage === 1" @click="goToPage(currentPage - 1)">
        上一頁
      </button>
      <div class="page-number-list">
        <template v-for="(page, index) in visiblePageNumbers" :key="`page-${page}`">
          <span
            v-if="index > 0 && page - visiblePageNumbers[index - 1] > 1"
            class="page-ellipsis"
          >
            ...
          </span>
          <button
            class="page-button"
            :class="{ active: page === currentPage }"
            type="button"
            @click="goToPage(page)"
          >
            {{ page }}
          </button>
        </template>
      </div>
      <button class="page-button" type="button" :disabled="currentPage === totalPages" @click="goToPage(currentPage + 1)">
        下一頁
      </button>
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

.breadcrumb {
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

.intro-row p,
.empty-state {
  color: var(--wp-text-muted);
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
  gap: 0.85rem;
}

.filters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.filters input,
.filters select {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
  padding: 0.8rem 0.9rem;
  color: var(--wp-text);
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 1rem;
}

.empty-state {
  text-align: center;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.page-number-list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.page-button {
  min-width: 2.5rem;
  min-height: 2.5rem;
  padding: 0.55rem 0.85rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.55rem;
  background: #fff;
  color: var(--wp-text);
  font-weight: 600;
}

.page-button.active {
  border-color: var(--tenant-accent, var(--wp-blue));
  background: color-mix(in srgb, var(--tenant-accent, var(--wp-blue)) 12%, #fff);
  color: var(--tenant-accent, var(--wp-blue));
}

.page-button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.page-ellipsis {
  color: var(--wp-text-muted);
  font-weight: 700;
}

@media (max-width: 1280px) {
  .product-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .filters {
    grid-template-columns: 1fr;
  }

  .product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .pagination {
    justify-content: stretch;
  }

  .page-number-list {
    justify-content: center;
  }
}
</style>
