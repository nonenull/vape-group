<script setup lang="ts">
import { computed } from 'vue'
import ProductCard from '~/components/store/ProductCard.vue'
import { fetchBrands, fetchCategories, fetchProducts } from '~/composables/useStoreApi'
import { buildCategoryPath, buildProductPath } from '~/composables/useProductSlug'
import { createItemListJsonLd, createStoreJsonLd, useStoreSeo } from '~/composables/useStoreSeo'
import type { HomeSectionConfig } from '~/types/store'

const generatedBannerImage = '/generated/ig_04c0f9fe9013d072016a1936ee03348191bae85cbc14a15435.png'

const tenantStore = useTenantStore()
await tenantStore.initTenant()

const [{ products }, categories, brands] = await Promise.all([
  fetchProducts(1, 200),
  fetchCategories(),
  fetchBrands(),
])

const tenant = tenantStore.currentTenant
const tenantName = tenant?.name ?? 'Vape Group 商城'
const categoryMap = new Map(categories.map((category) => [category.id, category]))

const primaryBrand = computed(() => {
  if (!tenant?.primaryBrandId) {
    return null
  }
  return brands.find((brand) => brand.id === tenant.primaryBrandId) ?? null
})

function getCategoryIds(product: (typeof products)[number]) {
  return product.categoryIds?.length
    ? product.categoryIds
    : product.categoryId != null
      ? [product.categoryId]
      : []
}

const primaryBrandProducts = computed(() => {
  if (!primaryBrand.value) {
    return []
  }
  return products.filter((product) => product.brand === primaryBrand.value?.name)
})

function getHotProductScore(product: (typeof products)[number]) {
  const badgeWeight = product.badge?.trim() ? 18 : 0
  const reviewWeight = Math.min(product.reviews, 220)
  const ratingWeight = product.rating * 14
  const stockWeight = product.stock > 0 ? 8 : 0
  return badgeWeight + reviewWeight + ratingWeight + stockWeight
}

const hotProducts = computed(() =>
  [...products]
    .sort((left, right) => {
      const scoreDifference = getHotProductScore(right) - getHotProductScore(left)
      if (scoreDifference !== 0) {
        return scoreDifference
      }
      return right.id - left.id
    })
    .slice(0, 8),
)

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
    .slice(0, 8),
)

const categoryCountMap = products.reduce((map, product) => {
  for (const categoryId of getCategoryIds(product)) {
    map.set(categoryId, (map.get(categoryId) ?? 0) + 1)
  }
  return map
}, new Map<number, number>())

const featuredCategories = computed(() => {
  const configured = tenantStore.platformConfig.featuredCategoryIds.filter((id) => categoryMap.has(id))
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

const configuredPrimaryCategoryCards = [
  { key: 'device', title: '設備', keywords: ['煙桿', '烟杆', '主機', '主机', '設備', 'device', 'kit'] },
  { key: 'pod', title: '煙彈', keywords: ['煙彈', '烟弹', 'pod', '彈', '弹'] },
  { key: 'disposable', title: '拋棄式', keywords: ['拋棄式', '一次性', '電子煙', 'disposable'] },
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
        .filter((product) => getCategoryIds(product).includes(categoryId))
        .slice(0, 4)
    return {
      key: group.key,
      title: group.title,
      categoryId,
      items,
    }
  }),
)

const defaultSections: HomeSectionConfig[] = [
  { id: 'brand-categories', type: 'brand_categories', enabled: true, title: '品牌分類', limit: 3 },
  { id: 'hot-products', type: 'hot_products', enabled: true, title: '最近熱賣', limit: 8 },
  { id: 'other-brands', type: 'other_brands', enabled: true, title: '其他品牌', limit: 8 },
  { id: 'featured-categories', type: 'featured_categories', enabled: true, title: '熱門分類', limit: 6 },
]

const homeSections = computed(() => {
  const configured = (tenant?.homeSections ?? []).filter((section) => section.enabled !== false && section.type)
  return configured.length ? configured : defaultSections
})

const banner = computed(() => {
  const configured = tenant?.homeBanner
  return {
    title: configured?.title?.trim() || tenant?.heroTitle?.trim() || `${primaryBrand.value?.name || tenantName}`,
    subtitle: configured?.subtitle?.trim() || tenant?.tagline?.trim() || '精選商品',
    image: configured?.image || generatedBannerImage,
    link: configured?.link?.trim() || '/products',
    buttonText: configured?.buttonText?.trim() || '立即選購',
  }
})

const heroFeaturedProducts = computed(() => hotProducts.value.slice(0, 4))

function getSectionTitle(section: HomeSectionConfig, fallback: string) {
  return section.title?.trim() || fallback
}

const seoTitle = computed(() => tenant?.seoTitle?.trim() || `${banner.value.title} | ${tenantName}`)
const seoDescription = computed(() => tenant?.seoDescription?.trim() || banner.value.subtitle)

useStoreSeo({
  title: seoTitle.value,
  description: seoDescription.value,
  image: banner.value.image,
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
    <section class="hero">
      <div class="hero-copy">
        <h1>{{ banner.title }}</h1>
        <p>{{ banner.subtitle }}</p>
      </div>

      <div class="hero-media">
        <img :src="banner.image" :alt="banner.title" class="hero-image">
      </div>
    </section>

    <section v-if="heroFeaturedProducts.length" class="featured-strip">
      <NuxtLink
        v-for="product in heroFeaturedProducts"
        :key="product.id"
        :to="buildProductPath(product)"
        class="featured-strip-item"
      >
        <img :src="product.image" :alt="product.name">
        <div>
          <strong>{{ product.name }}</strong>
          <span>NT$ {{ product.price.toFixed(2) }}</span>
        </div>
      </NuxtLink>
    </section>

    <template v-for="section in homeSections" :key="section.id">
      <section v-if="section.type === 'hot_products'" class="panel section-panel">
        <div class="section-heading">
          <h2>{{ getSectionTitle(section, '最近熱賣') }}</h2>
          <NuxtLink class="section-link" to="/products">更多</NuxtLink>
        </div>

        <div v-if="hotProducts.length" class="product-grid">
          <ProductCard
            v-for="product in hotProducts.slice(0, section.limit || 8)"
            :key="product.id"
            :product="product"
            :show-detail-button="false"
          />
        </div>
        <div v-else class="empty-state">暂无商品</div>
      </section>

      <section v-else-if="section.type === 'brand_categories'" class="panel section-panel">
        <div class="section-heading">
          <h2>{{ getSectionTitle(section, '品牌分類') }}</h2>
        </div>

        <div class="category-stack-grid">
          <article
            v-for="card in primaryCategoryCards.slice(0, section.limit || 3)"
            :key="card.key"
            class="category-stack"
          >
            <div class="category-stack-head">
              <strong class="stack-title">{{ card.title }}</strong>
              <span v-if="card.items.length" class="stack-count">{{ card.items.length }} 款</span>
            </div>

            <div v-if="card.items.length" class="mini-product-grid">
              <NuxtLink
                v-for="product in card.items"
                :key="product.id"
                :to="buildProductPath(product)"
                class="mini-product-card"
              >
                <img :src="product.image" :alt="product.name">
                <div class="mini-product-copy">
                  <span>{{ product.name }}</span>
                  <small>NT$ {{ product.price.toFixed(2) }}</small>
                </div>
              </NuxtLink>
            </div>
            <div v-else class="empty-state">暂无商品</div>
          </article>
        </div>
      </section>

      <section v-else-if="section.type === 'other_brands'" class="panel section-panel">
        <div class="section-heading">
          <h2>{{ getSectionTitle(section, '其他品牌') }}</h2>
        </div>

        <div v-if="featuredBrands.length" class="brand-grid">
          <NuxtLink
            v-for="brand in featuredBrands.slice(0, section.limit || 8)"
            :key="brand.id"
            class="brand-card"
            :to="`/products?brand=${encodeURIComponent(brand.name)}`"
          >
            <div class="brand-logo-shell" :class="{ 'has-logo': !!brand.logoUrl }">
              <img v-if="brand.logoUrl" :src="brand.logoUrl" :alt="brand.name" class="brand-logo">
              <span v-else class="brand-placeholder">{{ brand.name.slice(0, 1) }}</span>
            </div>
            <strong>{{ brand.name }}</strong>
          </NuxtLink>
        </div>
        <div v-else class="empty-state">暂无品牌</div>
      </section>

      <section v-else-if="section.type === 'featured_categories'" class="panel section-panel">
        <div class="section-heading">
          <h2>{{ getSectionTitle(section, '熱門分類') }}</h2>
          <NuxtLink class="section-link" to="/products">更多</NuxtLink>
        </div>

        <div v-if="featuredCategories.length" class="category-grid">
          <NuxtLink
            v-for="category in featuredCategories.slice(0, section.limit || 6)"
            :key="category.id"
            :to="buildCategoryPath(category)"
            class="category-card"
          >
            <strong>{{ category.name }}</strong>
            <small>{{ category.count }} 件</small>
          </NuxtLink>
        </div>
        <div v-else class="empty-state">暂无分类</div>
      </section>
    </template>
  </section>
</template>

<style scoped>
.home-page {
  display: grid;
  gap: 1rem;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  gap: 1rem;
  padding: 1.2rem;
  border-radius: 1.25rem;
  border: 1px solid color-mix(in srgb, var(--tenant-border, #dcdcde) 80%, #ffffff);
  background:
    radial-gradient(circle at top left, color-mix(in srgb, var(--tenant-accent, #2271b1) 14%, transparent), transparent 28%),
    linear-gradient(135deg, #f8fcff 0%, #f1f7fb 45%, #ffffff 100%);
}

.hero-copy {
  display: grid;
  align-content: center;
  gap: 0.85rem;
}

.hero-copy h1 {
  margin: 0;
  font-size: clamp(2rem, 3.6vw, 3.6rem);
  line-height: 0.98;
  color: #12212f;
}

.hero-copy p {
  max-width: 30rem;
  color: #5d7384;
  line-height: 1.6;
}

.hero-media {
  min-width: 0;
}

.hero-image {
  width: 100%;
  height: 100%;
  min-height: 320px;
  border-radius: 1rem;
  object-fit: cover;
}

.featured-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.featured-strip-item {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 0.7rem;
  align-items: center;
  padding: 0.75rem;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--tenant-border, #dcdcde) 80%, #ffffff);
  background: rgba(255, 255, 255, 0.92);
  color: inherit;
  text-decoration: none;
}

.featured-strip-item img {
  width: 72px;
  height: 72px;
  border-radius: 0.8rem;
  object-fit: cover;
  background: #f3f7fa;
}

.featured-strip-item strong {
  display: block;
  color: #153040;
  line-height: 1.35;
}

.featured-strip-item span {
  color: #6d8393;
  font-size: 0.82rem;
}

.section-panel {
  display: grid;
  gap: 0.9rem;
  padding: 1rem;
  border-radius: 1.1rem;
}

.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.8rem;
}

.section-heading h2 {
  margin: 0;
  font-size: 1.15rem;
  color: #112535;
}

.section-link {
  color: var(--tenant-accent, #2271b1);
  font-size: 0.84rem;
  font-weight: 700;
}

.product-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.category-stack-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.category-stack {
  display: grid;
  gap: 0.9rem;
  height: 100%;
  padding: 1rem;
  border-radius: 1rem;
  background:
    radial-gradient(circle at top left, color-mix(in srgb, var(--tenant-accent, #2271b1) 8%, transparent), transparent 28%),
    linear-gradient(180deg, #ffffff, #f8fbfd);
  border: 1px solid rgba(221, 229, 236, 0.95);
}

.category-stack-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.stack-title {
  color: #153040;
  font-size: 1rem;
}

.stack-count {
  flex: 0 0 auto;
  padding: 0.28rem 0.55rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--tenant-accent, #2271b1) 10%, #ffffff);
  color: var(--tenant-accent, #2271b1);
  font-size: 0.72rem;
  font-weight: 700;
}

.mini-product-grid {
  display: grid;
  gap: 0.8rem;
  align-content: start;
  min-height: 22rem;
}

.mini-product-card {
  display: grid;
  grid-template-columns: 60px minmax(0, 1fr);
  gap: 0.6rem;
  align-items: start;
  padding: 0.5rem 0;
  color: inherit;
  text-decoration: none;
  border-top: 1px solid rgba(230, 236, 241, 0.9);
}

.mini-product-card:first-child {
  padding-top: 0;
  border-top: 0;
}

.mini-product-card img {
  width: 60px;
  height: 60px;
  border-radius: 0.75rem;
  object-fit: cover;
  background: #f3f7fa;
}

.mini-product-copy {
  display: grid;
  gap: 0.24rem;
}

.mini-product-card span {
  color: #173142;
  font-size: 0.84rem;
  line-height: 1.4;
}

.mini-product-copy small {
  color: #718596;
  font-size: 0.74rem;
}

.brand-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.brand-card {
  display: grid;
  gap: 0.55rem;
  justify-items: center;
  padding: 0.85rem 0.7rem;
  border-radius: 1rem;
  border: 1px solid rgba(220, 230, 238, 0.95);
  background: linear-gradient(180deg, #ffffff, #f8fbfd);
  color: inherit;
  text-decoration: none;
}

.brand-logo-shell {
  display: grid;
  place-items: center;
  width: 100%;
  min-height: 88px;
  padding: 0.75rem;
  border-radius: 0.9rem;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(228, 235, 240, 0.9);
}

.brand-logo-shell.has-logo {
  background: linear-gradient(180deg, #ffffff, #fbfdff);
}

.brand-logo {
  max-width: 100%;
  max-height: 52px;
  width: auto;
  object-fit: contain;
}

.brand-placeholder {
  display: inline-grid;
  place-items: center;
  width: 2.3rem;
  height: 2.3rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--tenant-accent, #2271b1) 12%, #ffffff);
  color: var(--tenant-accent, #2271b1);
  font-weight: 800;
}

.brand-card strong {
  color: #173142;
  font-size: 0.88rem;
  text-align: center;
  line-height: 1.35;
}

.category-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.category-card {
  display: grid;
  gap: 0.3rem;
  padding: 0.9rem;
  border-radius: 1rem;
  border: 1px solid rgba(220, 230, 238, 0.95);
  background: linear-gradient(180deg, #ffffff, #f8fbfd);
  color: inherit;
  text-decoration: none;
}

.category-card strong {
  color: #173142;
}

.category-card small,
.empty-state {
  color: #718596;
}

.empty-state {
  min-height: 22rem;
  display: grid;
  align-content: start;
}

@media (max-width: 1180px) {
  .featured-strip,
  .product-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .hero {
    grid-template-columns: 1fr;
  }

  .featured-strip {
    display: flex;
    flex-direction: column;
  }

  .product-grid,
  .category-stack-grid,
  .brand-grid,
  .category-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .home-page {
    gap: 0.85rem;
  }

  .hero {
    gap: 0.85rem;
    padding: 0.9rem;
  }

  .hero-copy h1 {
    font-size: 1.75rem;
  }

  .hero-copy p {
    font-size: 0.88rem;
  }

  .hero-image {
    min-height: 220px;
  }

  .featured-strip {
    gap: 0.55rem;
  }

  .featured-strip-item {
    width: 100%;
    grid-template-columns: 56px minmax(0, 1fr);
    padding: 0.55rem;
  }

  .featured-strip-item img {
    width: 56px;
    height: 56px;
  }

  .featured-strip,
  .product-grid,
  .brand-grid,
  .category-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .category-stack-grid {
    grid-template-columns: 1fr;
  }

  .mini-product-grid,
  .empty-state {
    min-height: 0;
  }

  .section-panel {
    padding: 0.85rem;
  }

  .section-heading h2 {
    font-size: 1rem;
  }
}
</style>
