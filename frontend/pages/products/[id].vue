<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import ProductCard from '~/components/store/ProductCard.vue'
import { createBreadcrumbJsonLd, createProductJsonLd, useStoreSeo } from '~/composables/useStoreSeo'
import { fetchProductDetail, fetchProducts } from '~/composables/useStoreApi'
import { buildProductPath, parseProductIdFromRouteParam } from '~/composables/useProductSlug'

const tenantStore = useTenantStore()
await tenantStore.initTenant()

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()
const checkoutStore = useCheckoutStore()

const productId = parseProductIdFromRouteParam(route.params.id)
const product = await fetchProductDetail(productId)
const { products } = await fetchProducts(1, 20)

if (!product) {
  throw createError({
    statusCode: 404,
    statusMessage: '商品不存在',
  })
}

const tenantName = computed(() => tenantStore.currentTenant?.name ?? 'Vape Group 商城')
const canonicalPath = buildProductPath(product)

if (route.path !== canonicalPath) {
  await navigateTo(canonicalPath, { redirectCode: 301, replace: true })
}
const selectedOptions = ref<Record<string, string>>({})
const activeImage = ref(product.image)
const isImagePreviewOpen = ref(false)
const productIntro = computed(() => product.description.trim())
const productNote = computed(() => product.longDescription.trim())
const relatedProducts = computed(() => products.filter((item) => item.id !== product.id).slice(0, 4))
const selectedSkuVariant = computed(() => {
  if (!product.skuVariants.length) {
    return null
  }
  return product.skuVariants.find((variant) =>
    product.optionGroups.every((group) => selectedOptions.value[group.name] === variant.selections[group.name]),
  ) ?? null
})
const displayStock = computed(() => selectedSkuVariant.value?.stock ?? product.stock)
const displayPrice = computed(() => selectedSkuVariant.value?.price ?? product.salePrice ?? product.price)

watch(() => product.id, () => {
  activeImage.value = product.image
  isImagePreviewOpen.value = false
  selectedOptions.value = product.optionGroups.length
    ? Object.fromEntries(product.optionGroups.map((group) => [group.name, group.values[0] ?? '']))
    : {}
}, { immediate: true })

const openImagePreview = (image?: string) => {
  activeImage.value = image || activeImage.value || product.image
  isImagePreviewOpen.value = true
}

const closeImagePreview = () => {
  isImagePreviewOpen.value = false
}

const getCurrentPurchasePayload = () => {
  if (product.skuVariants.length && !selectedSkuVariant.value) {
    if (import.meta.client) {
      window.alert('請先選擇完整規格')
    }
    return null
  }

  const optionLabel = product.optionGroups
    .map((group) => `${group.name}：${selectedOptions.value[group.name]}`)
    .filter((item) => !item.endsWith('：'))
    .join(' / ')

  return {
    product,
    variant: selectedSkuVariant.value,
    optionLabel,
  }
}

const addCurrentProduct = async () => {
  const payload = getCurrentPurchasePayload()
  if (!payload) {
    return
  }
  cartStore.addItem(payload.product, payload.variant, payload.optionLabel)
  await router.push('/cart')
}

const buyNow = async () => {
  const payload = getCurrentPurchasePayload()
  if (!payload) {
    return
  }
  checkoutStore.startDirectCheckout(payload.product, payload.variant, payload.optionLabel)
  await router.push('/checkout')
}

const description =
  product.seoDescription ||
  product.description ||
  product.longDescription ||
  `${product.name} 商品詳情`
const title =
  product.seoTitle ||
  `${product.name} | ${product.category} | ${tenantName.value}`

useStoreSeo({
  title,
  description,
  image: product.image,
  canonicalPath,
  type: 'product',
  siteName: tenantName.value,
  locale: 'zh_TW',
  lang: 'zh-Hant',
  jsonLd: [
    createBreadcrumbJsonLd({
      items: [
        { name: '首頁', path: '/' },
        { name: '商品目錄', path: '/products' },
        { name: product.name, path: canonicalPath },
      ],
    }),
    createProductJsonLd({
      name: product.name,
      description,
      image: product.gallery.length ? product.gallery : [product.image],
      sku: product.sku,
      category: product.category,
      price: displayPrice.value,
      availability: displayStock.value > 0,
      rating: product.rating,
      reviews: product.reviews,
    }),
  ],
})
</script>

<template>
  <section class="product-detail-page">
    <p class="breadcrumb">首頁 / 商品目錄 / {{ product.name }}</p>

    <div class="detail-layout panel">
      <div class="media-panel">
        <button type="button" class="image-stage" @click="openImagePreview(activeImage || product.image)">
          <img :src="activeImage || product.image" :alt="product.name">
          <span v-if="product.badge" class="badge">{{ product.badge }}</span>
          <span class="image-stage-tip">點擊查看大圖</span>
        </button>
        <div v-if="product.gallery.length" class="thumb-grid">
          <button
            v-for="(image, index) in product.gallery"
            :key="`${image}-${index}`"
            type="button"
            class="thumb-button"
            :class="{ active: (activeImage || product.image) === image }"
            @click="activeImage = image"
            @dblclick="openImagePreview(image)"
          >
            <img :src="image" :alt="`${product.name}-thumb-${index}`">
          </button>
        </div>
      </div>

      <div class="content-panel">
        <h1>{{ product.name }}</h1>
        <div v-if="productIntro" class="intro-block">
          <p class="description">{{ productIntro }}</p>
        </div>
        <div class="price-row">
          <strong>NT$ {{ displayPrice.toFixed(2) }}</strong>
          <span v-if="product.salePrice">NT$ {{ product.price.toFixed(2) }}</span>
        </div>
        <div v-if="product.optionGroups.length" class="variant-picker">
          <div v-for="group in product.optionGroups" :key="group.name" class="variant-group">
            <span>{{ group.name }}</span>
            <div class="variant-options">
              <button
                v-for="value in group.values"
                :key="`${group.name}-${value}`"
                type="button"
                class="variant-button"
                :class="{ active: selectedOptions[group.name] === value }"
                @click="selectedOptions[group.name] = value"
              >
                {{ value }}
              </button>
            </div>
          </div>
          <small>所選 SKU：{{ selectedSkuVariant?.sku ?? product.sku }}</small>
        </div>
        <div class="actions">
          <button class="primary" type="button" @click="addCurrentProduct">加入購物車</button>
          <button class="secondary" type="button" @click="buyNow">直接下單</button>
        </div>
      </div>
    </div>

    <div class="detail-sections">
      <div class="detail-main-column">
        <article class="panel">
          <h2>商品規格</h2>
          <div class="spec-list">
            <div v-for="spec in product.specs" :key="spec.label" class="spec-row">
              <span>{{ spec.label }}</span>
              <strong>{{ spec.value }}</strong>
            </div>
          </div>
        </article>

        <article v-if="productNote" class="panel">
          <h2>商品說明</h2>
          <p class="detail-copy">{{ productNote }}</p>
        </article>

        <article v-if="product.detailImages.length" class="panel">
          <h2>商品詳情圖</h2>
          <div class="detail-images">
            <img
              v-for="(image, index) in product.detailImages"
              :key="`${image}-${index}`"
              :src="image"
              :alt="`${product.name}-detail-${index}`"
            >
          </div>
        </article>
      </div>

      <aside class="detail-side-column">
        <article class="panel related-panel">
          <div class="section-head">
            <h2>推薦商品</h2>
            <NuxtLink to="/products">瀏覽更多</NuxtLink>
          </div>
          <div class="related-list">
            <ProductCard v-for="item in relatedProducts" :key="item.id" :product="item" :show-detail-button="false" />
          </div>
        </article>
      </aside>
    </div>

    <div v-if="isImagePreviewOpen" class="image-preview-overlay" @click="closeImagePreview">
      <button type="button" class="image-preview-close" aria-label="關閉圖片預覽" @click.stop="closeImagePreview">
        關閉
      </button>
      <div class="image-preview-dialog" @click.stop>
        <img :src="activeImage || product.image" :alt="`${product.name}-preview`">
      </div>
    </div>
  </section>
</template>

<style scoped>
.product-detail-page {
  display: grid;
  gap: 1rem;
}

.breadcrumb,
.description,
.spec-row span,
.detail-copy,
.variant-picker small {
  color: var(--wp-text-muted);
}

.detail-copy {
  line-height: 1.8;
  white-space: pre-line;
}

.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(320px, 0.95fr);
  gap: 1.5rem;
}

.media-panel,
.content-panel,
.detail-sections,
.detail-main-column,
.thumb-grid,
.detail-images {
  display: grid;
  gap: 1rem;
}

.related-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.related-list :deep(.btn-secondary) {
  display: none;
}

.related-list :deep(.card-actions .btn-primary) {
  width: 100%;
}

.image-stage {
  position: relative;
  display: block;
  width: 100%;
  padding: 1.4rem;
  border-radius: 0.75rem;
  background: linear-gradient(180deg, #f8fbfe 0%, #eef3f7 100%);
  border: 1px solid var(--wp-border);
  cursor: zoom-in;
}

.image-stage img {
  width: 100%;
  max-height: 520px;
  object-fit: contain;
}

.image-stage-tip {
  position: absolute;
  right: 1rem;
  bottom: 1rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  background: rgba(16, 21, 23, 0.68);
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 600;
}

.badge {
  position: absolute;
  top: 1rem;
  left: 1rem;
  padding: 0.28rem 0.6rem;
  background: var(--wp-green);
  color: #ffffff;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.thumb-grid {
  grid-template-columns: repeat(auto-fit, minmax(82px, 1fr));
}

.thumb-button,
.variant-button {
  border: 1px solid var(--wp-border);
  background: #ffffff;
  cursor: pointer;
}

.thumb-button {
  padding: 0.45rem;
  border-radius: 0.75rem;
}

.thumb-button.active,
.variant-button.active {
  border-color: var(--tenant-accent, var(--wp-blue));
  box-shadow: 0 0 0 2px rgba(34, 113, 177, 0.12);
}

.image-preview-overlay {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background: rgba(16, 21, 23, 0.88);
}

.image-preview-dialog {
  max-width: min(92vw, 1320px);
  max-height: 88vh;
  display: grid;
  place-items: center;
}

.image-preview-dialog img {
  max-width: 100%;
  max-height: 88vh;
  object-fit: contain;
  border-radius: 1rem;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35);
}

.image-preview-close {
  position: absolute;
  top: 1.25rem;
  right: 1.25rem;
  padding: 0.7rem 1rem;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.24);
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
  cursor: pointer;
}

.content-panel h1,
.panel h2 {
  color: var(--wp-heading);
}

.content-panel {
  align-content: start;
}

.price-row,
.actions,
.section-head,
.spec-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.price-row strong {
  color: var(--wp-heading);
  font-size: 1.8rem;
}

.price-row span {
  color: var(--wp-text-muted);
  text-decoration: line-through;
}

.variant-picker,
.variant-group,
.variant-options,
.spec-list {
  display: grid;
  gap: 0.85rem;
}

.variant-options {
  grid-template-columns: repeat(auto-fit, minmax(98px, 1fr));
}

.variant-button {
  border-radius: 999px;
  padding: 0.75rem 0.9rem;
}

.detail-sections {
  grid-template-columns: minmax(0, 1fr) 360px;
  align-items: start;
}

.detail-images img {
  border-radius: 0.75rem;
  border: 1px solid var(--wp-border);
}

.spec-row {
  padding: 0.8rem 0;
  border-bottom: 1px solid var(--wp-border);
}

.spec-row:last-child {
  border-bottom: none;
}

@media (min-width: 1081px) {
  .spec-list {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 1.25rem;
    row-gap: 0.2rem;
  }

  .spec-row {
    display: grid;
    grid-template-columns: 72px minmax(0, 1fr);
    justify-content: start;
    align-items: start;
    padding: 0.55rem 0;
    gap: 0.35rem;
  }

  .detail-main-column > .panel:first-child {
    padding: 1rem 1.15rem;
  }

  .detail-main-column > .panel:first-child h2 {
    margin-bottom: 0.4rem;
    font-size: 1.05rem;
  }

  .spec-row span,
  .spec-row strong {
    font-size: 0.87rem;
    line-height: 1.35;
  }

  .spec-row strong {
    text-align: left;
  }
}

@media (max-width: 1080px) {
  .detail-layout,
  .detail-sections {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .related-list :deep(.card-actions .btn-primary) {
    min-height: 36px;
    padding: 0.55rem 0.75rem;
    font-size: 0.82rem;
  }
}
</style>
