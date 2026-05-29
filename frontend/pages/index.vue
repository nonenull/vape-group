<script setup lang="ts">
import { computed } from 'vue'
import ProductCard from '~/components/store/ProductCard.vue'
import { createItemListJsonLd, createStoreJsonLd, useStoreSeo } from '~/composables/useStoreSeo'
import { fetchBrands, fetchCategories, fetchProducts } from '~/composables/useStoreApi'
import { buildCategoryPath } from '~/composables/useProductSlug'
import { buildProductPath } from '~/composables/useProductSlug'

const tenantStore = useTenantStore()
await tenantStore.initTenant()

const [{ products }, categories, brands] = await Promise.all([
  fetchProducts(1, 200),
  fetchCategories(),
  fetchBrands(),
])

const tenant = tenantStore.currentTenant
const tenantName = tenant?.name ?? 'Vape Group 商城'
const homeModuleDefaults = ['products', 'categories', 'brands'] as const
type HomeModuleKey = typeof homeModuleDefaults[number]
const homeModuleSet = new Set<HomeModuleKey>(homeModuleDefaults)

const primaryBrand = computed(() => {
  if (!tenant?.primaryBrandId) {
    return null
  }
  return brands.find((brand) => brand.id === tenant.primaryBrandId) ?? null
})

const primaryBrandProducts = computed(() => {
  if (!primaryBrand.value) {
    return []
  }
  return products.filter((product) => product.brand === primaryBrand.value?.name)
})

const categoryMap = new Map(categories.map((category) => [category.id, category]))

const configuredPrimaryCategoryCards = [
  { key: 'device', title: '煙桿', keywords: ['煙桿', '烟杆', '主機', '主机', '設備', '设备', 'pod 系統', '套裝', '套装', 'kit', 'device'] },
  { key: 'pod', title: '煙彈', keywords: ['煙彈', '烟弹', 'pod', '彈', '弹'] },
  { key: 'disposable', title: '拋棄式電子煙', keywords: ['拋棄式', '一次性', '電子煙', '电子烟', 'disposable'] },
] as const

function findMatchingCategoryId(keywords: readonly string[]) {
  const match = categories.find((category) => {
    const categoryName = category.name.trim().toLowerCase()
    return keywords.some((keyword) => categoryName.includes(keyword.toLowerCase()))
  })
  return match?.id ?? null
}

const primaryCategoryCards = computed(() =>
  configuredPrimaryCategoryCards.map((group) => {
    const categoryId = findMatchingCategoryId(group.keywords)
    const items = categoryId == null
      ? []
      : primaryBrandProducts.value
        .filter((product) => {
          const productCategoryIds = product.categoryIds?.length
            ? product.categoryIds
            : product.categoryId != null
              ? [product.categoryId]
              : []
          return productCategoryIds.includes(categoryId)
        })
        .slice(0, 4)
    return {
      key: group.key,
      title: group.title,
      categoryId,
      categoryName: categoryId != null ? categoryMap.get(categoryId)?.name ?? '' : '',
      items,
    }
  }),
)

const featuredCategoryIds = tenantStore.platformConfig.featuredCategoryIds
const categoryCountMap = products.reduce((map, product) => {
  const productCategoryIds = product.categoryIds?.length
    ? product.categoryIds
    : product.categoryId != null
      ? [product.categoryId]
      : []
  for (const categoryId of productCategoryIds) {
    map.set(categoryId, (map.get(categoryId) ?? 0) + 1)
  }
  return map
}, new Map<number, number>())

const featuredCategories = computed(() => {
  const configured = featuredCategoryIds.filter((id) => categoryMap.has(id))
  const fallback = categories
    .filter((category) => (categoryCountMap.get(category.id) ?? 0) > 0)
    .sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id)
    .map((category) => category.id)

  return (configured.length ? configured : fallback)
    .slice(0, 6)
    .map((id) => {
      const category = categoryMap.get(id)
      if (!category) {
        return null
      }
      return {
        id: category.id,
        name: category.name,
        count: categoryCountMap.get(category.id) ?? 0,
      }
    })
    .filter((item): item is NonNullable<typeof item> => item !== null)
})

function getStableBrandRandomScore(brandId: number) {
  const baseSeed = `${tenant?.id ?? 0}-${tenant?.primaryBrandId ?? 0}-${brandId}`
  let hash = 0
  for (let index = 0; index < baseSeed.length; index += 1) {
    hash = (hash * 31 + baseSeed.charCodeAt(index)) >>> 0
  }
  return hash
}

const featuredBrands = computed(() =>
  brands
    .filter((brand) => brand.id !== primaryBrand.value?.id)
    .sort((left, right) => {
      const scoreDifference = getStableBrandRandomScore(left.id) - getStableBrandRandomScore(right.id)
      if (scoreDifference !== 0) {
        return scoreDifference
      }
      return left.id - right.id
    })
    .slice(0, 6),
)

function getSeasonalProductScore(product: (typeof products)[number]) {
  const saleWeight = product.salePrice ? 50 : 0
  const badgeWeight = product.badge?.trim() ? 20 : 0
  const reviewWeight = Math.min(product.reviews, 200)
  const ratingWeight = product.rating * 10
  const stockWeight = product.stock > 0 ? 5 : 0
  return saleWeight + badgeWeight + reviewWeight + ratingWeight + stockWeight
}

const seasonalProducts = computed(() =>
  [...products]
    .sort((left, right) => {
      const scoreDifference = getSeasonalProductScore(right) - getSeasonalProductScore(left)
      if (scoreDifference !== 0) {
        return scoreDifference
      }
      return right.id - left.id
    })
    .slice(0, 8),
)

const orderedHomeModules = computed<HomeModuleKey[]>(() => {
  const configured = (tenant?.homeModuleOrder ?? [])
    .filter((item): item is HomeModuleKey => homeModuleSet.has(item as HomeModuleKey))
    .filter((item, index, array) => array.indexOf(item) === index)

  return [
    ...configured,
    ...homeModuleDefaults.filter((item) => !configured.includes(item)),
  ]
})

const seoTitle = computed(() => {
  if (tenant?.seoTitle?.trim()) {
    return tenant.seoTitle.trim()
  }
  const primaryBrandName = primaryBrand.value?.name?.trim()
  return primaryBrandName ? `${primaryBrandName} | ${tenantName}` : `${tenantName} | 商品首頁`
})

const seoDescription = computed(() => {
  if (tenant?.seoDescription?.trim()) {
    return tenant.seoDescription.trim()
  }

  const primaryBrandName = primaryBrand.value?.name?.trim()
  const categorySummary = primaryCategoryCards.value
    .map((card) => `${card.title}${card.items.length ? `${card.items.length}款` : ''}`)
    .join('、')
  const hotCategoryNames = featuredCategories.value.slice(0, 3).map((item) => item.name).join('、')
  const otherBrandNames = featuredBrands.value.slice(0, 3).map((item) => item.name).join('、')

  const segments = [
    primaryBrandName ? `${primaryBrandName}主打${categorySummary || '商品專區'}` : `${tenantName}首頁商品專區`,
    hotCategoryNames ? `熱門分類包含${hotCategoryNames}` : '',
    otherBrandNames ? `其他品牌有${otherBrandNames}` : '',
  ].filter(Boolean)

  return segments.join('，') || tenant?.tagline || `${tenantName} 精選商品與品牌內容總覽。`
})

const convenienceTools = [
  {
    key: 'store-locator',
    title: '7-11 門市查詢',
    description: '查詢可取貨或寄件的 7-ELEVEN 門市資訊。',
    href: 'https://emap.pcsc.com.tw/',
  },
  {
    key: 'parcel-tracking',
    title: '7-11 貨態查詢',
    description: '輸入交貨便單號，快速查看包裹目前配送狀態。',
    href: 'https://eservice.7-11.com.tw/E-Tracking/search.aspx',
  },
] as const

useStoreSeo({
  title: seoTitle.value,
  description: seoDescription.value,
  image: tenant?.previewImage || tenant?.logoImage || products[0]?.image,
  type: 'website',
  canonicalPath: '/',
  siteName: tenantName,
  locale: 'zh_TW',
  lang: 'zh-Hant',
  jsonLd: [
    createStoreJsonLd({
      name: tenantName,
      description: seoDescription.value,
    }),
    createItemListJsonLd({
      name: `${tenantName} 精選商品`,
      description: seoDescription.value,
      items: products.slice(0, 12).map((product) => ({
        name: product.name,
        path: buildProductPath(product),
        image: product.image,
      })),
    }),
  ],
})
</script>

<template>
  <section class="home-page">
    <section v-for="card in primaryCategoryCards" :key="card.key" class="panel product-section">
      <div class="section-heading">
        <div>
          <p class="section-label">Primary brand</p>
          <h2>{{ primaryBrand?.name ? `${primaryBrand.name} ${card.title}` : card.title }}</h2>
        </div>
      </div>

      <div v-if="card.items.length" class="product-grid">
        <ProductCard
          v-for="product in card.items"
          :key="product.id"
          :product="product"
          :show-detail-button="false"
        />
      </div>
      <div v-else class="empty-state">
        {{ card.categoryId == null ? `请先为 ${card.title} 配置分类。` : `主品牌下暂无对应 ${card.title} 商品。` }}
      </div>
    </section>

    <template v-for="moduleKey in orderedHomeModules" :key="moduleKey">
      <section v-if="moduleKey === 'products'" class="panel compact-section">
        <div class="section-heading">
          <div>
            <p class="section-label">Seasonal picks</p>
            <h2>本季热卖</h2>
          </div>
          <NuxtLink class="link-button" to="/products">查看全部</NuxtLink>
        </div>

        <div v-if="seasonalProducts.length" class="product-grid">
          <ProductCard
            v-for="product in seasonalProducts"
            :key="product.id"
            :product="product"
            :show-detail-button="false"
          />
        </div>
        <div v-else class="empty-state">暂无热卖商品可展示。</div>
      </section>

      <section v-else-if="moduleKey === 'categories'" class="panel compact-section">
        <div class="section-heading">
          <div>
            <p class="section-label">Browse categories</p>
            <h2>热门分类</h2>
          </div>
          <NuxtLink class="link-button" to="/products">查看全部</NuxtLink>
        </div>

        <div v-if="featuredCategories.length" class="compact-grid">
          <NuxtLink
            v-for="category in featuredCategories"
            :key="category.id"
            class="compact-card"
            :to="buildCategoryPath(category)"
          >
            <strong>{{ category.name }}</strong>
            <span>{{ category.count }} 件</span>
          </NuxtLink>
        </div>
        <div v-else class="empty-state">暂无热门分类。</div>
      </section>

      <section v-else-if="moduleKey === 'brands'" class="panel compact-section">
        <div class="section-heading">
          <div>
            <p class="section-label">Other brands</p>
            <h2>其他品牌</h2>
          </div>
        </div>

        <div v-if="featuredBrands.length" class="brand-grid">
          <NuxtLink
            v-for="brand in featuredBrands"
            :key="brand.id"
            class="brand-card"
            :to="`/products?brand=${encodeURIComponent(brand.name)}`"
          >
            <div class="brand-logo-shell">
              <img v-if="brand.logoUrl" :src="brand.logoUrl" :alt="brand.name" class="brand-logo" />
              <span v-else class="brand-placeholder">{{ brand.name.slice(0, 1) }}</span>
            </div>
            <h3>{{ brand.name }}</h3>
          </NuxtLink>
        </div>
        <div v-else class="empty-state">暂无其他品牌可展示。</div>
      </section>
    </template>

    <section class="panel compact-section convenience-section">
      <div class="section-heading">
        <div>
          <p class="section-label">Convenience tools</p>
          <h2>超商常用服務</h2>
        </div>
      </div>

      <div class="convenience-grid">
        <a
          v-for="item in convenienceTools"
          :key="item.key"
          class="convenience-card"
          :href="item.href"
          target="_blank"
          rel="noreferrer"
        >
          <strong>{{ item.title }}</strong>
          <span>{{ item.description }}</span>
        </a>
      </div>
    </section>
  </section>
</template>

<style scoped>
.home-page {
  display: grid;
  gap: 1.25rem;
}

.product-section,
.compact-section {
  display: grid;
  gap: 1rem;
}

.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
}

.section-label {
  color: var(--tenant-accent, var(--wp-blue));
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.section-heading h2 {
  margin: 0.3rem 0 0;
}

.product-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.compact-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.compact-card {
  display: grid;
  gap: 0.35rem;
  padding: 1rem;
  border: 1px solid var(--tenant-border, var(--wp-border));
  border-radius: 0.75rem;
  background: linear-gradient(180deg, #ffffff, var(--tenant-surface, #f6f7f7));
  color: inherit;
  text-decoration: none;
}

.compact-card strong {
  color: var(--tenant-text, var(--wp-heading));
}

.compact-card span {
  color: var(--tenant-muted, var(--wp-text-muted));
  font-size: 0.85rem;
}

.brand-grid {
  display: grid;
  gap: 0.6rem;
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.brand-card {
  display: grid;
  justify-items: center;
  gap: 0.5rem;
  padding: 0.75rem 0.55rem;
  border: 1px solid var(--tenant-border, var(--wp-border));
  border-radius: 0.7rem;
  background: linear-gradient(180deg, #ffffff, var(--tenant-surface, #f6f7f7));
  text-decoration: none;
  color: inherit;
}

.brand-logo-shell {
  display: grid;
  place-items: center;
  width: 100%;
  min-height: 60px;
  padding: 0.55rem;
  border-radius: 0.6rem;
  background: var(--tenant-card-bg, #ffffff);
}

.brand-logo {
  max-width: 100%;
  max-height: 34px;
  object-fit: contain;
}

.brand-placeholder {
  display: inline-grid;
  place-items: center;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: 999px;
  background: var(--tenant-tag-bg, var(--tenant-surface, #f6f7f7));
  color: var(--tenant-accent, var(--wp-blue));
  font-weight: 700;
  font-size: 0.9rem;
}

.brand-card h3 {
  margin: 0;
  font-size: 0.88rem;
  line-height: 1.3;
  text-align: center;
  color: var(--tenant-text, var(--wp-heading));
}

.convenience-section {
  gap: 0.9rem;
}

.convenience-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.convenience-card {
  display: grid;
  gap: 0.35rem;
  padding: 1rem 1.05rem;
  border: 1px solid var(--tenant-border, var(--wp-border));
  border-radius: 0.75rem;
  background: linear-gradient(180deg, #ffffff, var(--tenant-surface, #f6f7f7));
  color: inherit;
  text-decoration: none;
}

.convenience-card strong {
  color: var(--tenant-text, var(--wp-heading));
  font-size: 0.96rem;
}

.convenience-card span {
  color: var(--tenant-muted, var(--wp-text-muted));
  font-size: 0.84rem;
  line-height: 1.5;
}

.empty-state {
  color: var(--tenant-muted, var(--wp-text-muted));
}

@media (max-width: 1440px) {
  .product-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 1200px) {
  .product-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .brand-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .section-heading {
    flex-direction: column;
    align-items: flex-start;
  }

  .product-grid,
  .compact-grid,
  .convenience-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .brand-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .home-page {
    gap: 1rem;
  }

  .product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .compact-grid,
  .brand-grid,
  .convenience-grid {
    gap: 0.5rem;
  }

  .convenience-grid {
    grid-template-columns: 1fr;
  }

  .brand-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .compact-card,
  .brand-card {
    padding: 0.7rem 0.5rem;
  }

  .brand-logo-shell {
    min-height: 52px;
    padding: 0.45rem;
  }

  .brand-logo {
    max-height: 28px;
  }

  .brand-card h3,
  .compact-card strong {
    font-size: 0.85rem;
  }
}
</style>
