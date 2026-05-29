<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import ProductCard from '~/components/store/ProductCard.vue'
import { createBreadcrumbJsonLd, createProductJsonLd, sanitizeProductMetaDescription, useStoreSeo } from '~/composables/useStoreSeo'
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
const previewScale = ref(1)
const previewOffsetX = ref(0)
const previewOffsetY = ref(0)
const isPreviewDragging = ref(false)
const didPreviewDrag = ref(false)
const previewDragStartX = ref(0)
const previewDragStartY = ref(0)
const previewDragOriginX = ref(0)
const previewDragOriginY = ref(0)
const previewViewportRef = ref<HTMLElement | null>(null)
const productIntro = computed(() => product.description.trim())
const productNote = computed(() => product.longDescription.trim())
const productSpecificationHtml = computed(() => product.specificationHtml.trim())
const productFaqHtml = computed(() => tenantStore.platformConfig.faqHtml.trim())
const relatedProducts = computed(() => products.filter((item) => item.id !== product.id).slice(0, 4))
const selectedSkuVariant = computed(() => {
  if (!product.skuVariants.length) {
    return null
  }
  return product.skuVariants.find((variant) =>
    product.optionGroups.every((group) => selectedOptions.value[group.name] === variant.selections[group.name]),
  ) ?? null
})
const displayStock = computed(() => {
  const variantStock = selectedSkuVariant.value?.stock
  return variantStock == null ? product.stock : variantStock
})
const displayPrice = computed(() => selectedSkuVariant.value?.price ?? product.salePrice ?? product.price)
const basicInfoSpecs = computed(() =>
  product.specs.map((spec) => {
    if (spec.label === 'SKU') {
      return {
        ...spec,
        value: selectedSkuVariant.value?.sku ?? product.sku,
      }
    }
    if (spec.label === '庫存') {
      return {
        ...spec,
        value: String(displayStock.value),
      }
    }
    return spec
  }),
)
const basicInfoBadges = computed(() => {
  const items: Array<{ label: string; value: string }> = []

  if (product.category && product.category !== '未分類') {
    items.push({ label: '分類', value: product.category })
  }

  if (product.brand) {
    items.push({ label: '品牌', value: product.brand })
  }

  if (product.optionGroups.length) {
    items.push({ label: '可選規格', value: `${product.optionGroups.length} 組` })
  }

  return items
})
const stockState = computed(() => {
  if (displayStock.value <= 0) {
    return {
      label: '已售罄',
      tone: 'soldout',
    }
  }

  if (displayStock.value <= 10) {
    return {
      label: `庫存 ${displayStock.value}`,
      tone: 'low',
    }
  }

  return {
    label: `庫存 ${displayStock.value}`,
    tone: 'ready',
  }
})

const hasIntroHtml = computed(() => /<[^>]+>/.test(productIntro.value))
const hasNoteHtml = computed(() => /<[^>]+>/.test(productNote.value))
const hasSpecificationHtml = computed(() => /<[^>]+>/.test(productSpecificationHtml.value))

watch(() => product.id, () => {
  activeImage.value = product.image
  isImagePreviewOpen.value = false
  previewScale.value = 1
  previewOffsetX.value = 0
  previewOffsetY.value = 0
  isPreviewDragging.value = false
  selectedOptions.value = product.optionGroups.length
    ? Object.fromEntries(product.optionGroups.map((group) => [group.name, group.values[0] ?? '']))
    : {}
}, { immediate: true })

const resetPreviewTransform = () => {
  previewScale.value = 1
  previewOffsetX.value = 0
  previewOffsetY.value = 0
  isPreviewDragging.value = false
  didPreviewDrag.value = false
}

const openImagePreview = (image?: string) => {
  activeImage.value = image || activeImage.value || product.image
  resetPreviewTransform()
  isImagePreviewOpen.value = true
}

const closeImagePreview = () => {
  isImagePreviewOpen.value = false
  resetPreviewTransform()
}

const setPreviewScale = (nextScale: number) => {
  previewScale.value = Math.min(3, Math.max(1, Number(nextScale.toFixed(2))))
  if (previewScale.value === 1) {
    previewOffsetX.value = 0
    previewOffsetY.value = 0
    isPreviewDragging.value = false
  }
}

const zoomPreview = (delta: number) => {
  setPreviewScale(previewScale.value + delta)
}

const togglePreviewZoom = () => {
  if (previewScale.value > 1) {
    resetPreviewTransform()
    return
  }
  setPreviewScale(2)
}

const zoomToPreviewPoint = (clientX: number, clientY: number, nextScale: number) => {
  const viewport = previewViewportRef.value
  if (!viewport) {
    setPreviewScale(nextScale)
    return
  }
  const rect = viewport.getBoundingClientRect()
  const localX = clientX - rect.left - rect.width / 2
  const localY = clientY - rect.top - rect.height / 2
  const scaleRatio = nextScale / previewScale.value
  previewOffsetX.value = previewOffsetX.value * scaleRatio - localX * (scaleRatio - 1)
  previewOffsetY.value = previewOffsetY.value * scaleRatio - localY * (scaleRatio - 1)
  setPreviewScale(nextScale)
}

const handlePreviewImageClick = (event: MouseEvent) => {
  if (didPreviewDrag.value) {
    didPreviewDrag.value = false
    return
  }
  if (previewScale.value > 1) {
    resetPreviewTransform()
    return
  }
  zoomToPreviewPoint(event.clientX, event.clientY, 2)
}

const handlePreviewWheel = (event: WheelEvent) => {
  event.preventDefault()
  zoomPreview(event.deltaY < 0 ? 0.2 : -0.2)
}

const handleRichCopyClick = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof HTMLImageElement)) {
    return
  }

  const imageSrc = target.currentSrc || target.src
  if (!imageSrc) {
    return
  }

  openImagePreview(imageSrc)
}

const handlePreviewPointerDown = (event: PointerEvent) => {
  if (previewScale.value <= 1) {
    return
  }
  isPreviewDragging.value = true
  didPreviewDrag.value = false
  previewDragStartX.value = event.clientX
  previewDragStartY.value = event.clientY
  previewDragOriginX.value = previewOffsetX.value
  previewDragOriginY.value = previewOffsetY.value
  ;(event.currentTarget as HTMLElement | null)?.setPointerCapture?.(event.pointerId)
  event.preventDefault()
}

const handlePreviewPointerMove = (event: PointerEvent) => {
  if (!isPreviewDragging.value || previewScale.value <= 1) {
    return
  }
  const deltaX = event.clientX - previewDragStartX.value
  const deltaY = event.clientY - previewDragStartY.value
  if (Math.abs(deltaX) > 3 || Math.abs(deltaY) > 3) {
    didPreviewDrag.value = true
  }
  previewOffsetX.value = previewDragOriginX.value + deltaX
  previewOffsetY.value = previewDragOriginY.value + deltaY
}

const stopPreviewDragging = (event?: PointerEvent) => {
  if (event) {
    ;(event.currentTarget as HTMLElement | null)?.releasePointerCapture?.(event.pointerId)
  }
  isPreviewDragging.value = false
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
}

const buyNow = async () => {
  const payload = getCurrentPurchasePayload()
  if (!payload) {
    return
  }
  checkoutStore.startDirectCheckout(payload.product, payload.variant, payload.optionLabel)
  await router.push('/checkout')
}

const description = sanitizeProductMetaDescription({
  description:
    product.seoDescription ||
    product.description ||
    product.longDescription ||
    `${product.name} 商品詳情`,
  productName: product.name,
})
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
        <div
          v-if="product.gallery.length"
          class="thumb-grid"
          :class="{ 'thumb-grid-single': product.gallery.length === 1 }"
        >
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
          <div v-if="hasIntroHtml" class="rich-copy" v-html="productIntro"></div>
          <p v-else class="description">{{ productIntro }}</p>
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
        </div>
        <div class="actions">
          <button class="primary" type="button" @click="addCurrentProduct">加入購物車</button>
          <button class="secondary" type="button" @click="buyNow">直接下單</button>
        </div>
      </div>
    </div>

    <div class="detail-sections">
      <div class="detail-main-column">
        <article class="panel basic-info-panel">
          <div class="basic-info-header">
            <div>
              <p class="section-kicker">Quick Overview</p>
              <h2>基本資訊</h2>
              <p class="basic-info-caption">先把關鍵資料整理好，價格、分類、品牌和目前庫存一眼就能看懂。</p>
            </div>
            <span class="stock-indicator" :class="`stock-indicator-${stockState.tone}`">
              {{ stockState.label }}
            </span>
          </div>
          <div v-if="basicInfoBadges.length" class="basic-info-pills">
            <span
              v-for="badge in basicInfoBadges"
              :key="`${badge.label}-${badge.value}`"
              class="basic-info-pill"
            >
              <strong>{{ badge.label }}</strong>
              <span>{{ badge.value }}</span>
            </span>
          </div>
          <div class="basic-info-grid">
            <div
              v-for="spec in basicInfoSpecs"
              :key="spec.label"
              class="basic-info-item"
            >
              <span class="basic-info-label">{{ spec.label }}</span>
              <strong class="basic-info-value">{{ spec.value || '未提供' }}</strong>
            </div>
          </div>
        </article>

        <article v-if="productNote" class="panel">
          <h2>產品說明</h2>
          <div v-if="hasNoteHtml" class="rich-copy detail-copy" v-html="productNote" @click="handleRichCopyClick"></div>
          <p v-else class="detail-copy">{{ productNote }}</p>
        </article>

        <article v-if="productSpecificationHtml" class="panel">
          <h2>產品規格</h2>
          <div
            v-if="hasSpecificationHtml"
            class="rich-copy detail-copy"
            v-html="productSpecificationHtml"
            @click="handleRichCopyClick"
          ></div>
          <p v-else class="detail-copy">{{ productSpecificationHtml }}</p>
        </article>

        <article v-if="productFaqHtml" class="panel">
          <h2>常見問題</h2>
          <div class="rich-copy detail-copy" v-html="productFaqHtml" @click="handleRichCopyClick"></div>
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
      <div class="image-preview-toolbar" @click.stop>
        <button type="button" class="image-preview-action" aria-label="缩小预览图" @click="zoomPreview(-0.2)">
          -
        </button>
        <span class="image-preview-scale">{{ Math.round(previewScale * 100) }}%</span>
        <button type="button" class="image-preview-action" aria-label="放大预览图" @click="zoomPreview(0.2)">
          +
        </button>
        <button type="button" class="image-preview-action" aria-label="重置预览图缩放" @click="resetPreviewTransform">
          重置
        </button>
        <button type="button" class="image-preview-close" aria-label="關閉圖片預覽" @click="closeImagePreview">
          關閉
        </button>
      </div>
      <div
        ref="previewViewportRef"
        class="image-preview-dialog"
        :class="{ dragging: isPreviewDragging, zoomed: previewScale > 1 }"
        @click.stop
        @wheel="handlePreviewWheel"
        @pointerdown.stop="handlePreviewPointerDown"
        @pointermove.stop="handlePreviewPointerMove"
        @pointerup.stop="stopPreviewDragging"
        @pointercancel.stop="stopPreviewDragging"
      >
        <img
          :src="activeImage || product.image"
          :alt="`${product.name}-preview`"
          :style="{ transform: `translate(${previewOffsetX}px, ${previewOffsetY}px) scale(${previewScale})` }"
          :class="{ zoomed: previewScale > 1, dragging: isPreviewDragging }"
          @click.stop="handlePreviewImageClick"
        >
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
.detail-copy {
  color: var(--wp-text-muted);
}

.section-kicker {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.45rem;
  color: var(--tenant-accent, var(--wp-blue));
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.detail-copy {
  line-height: 1.8;
  white-space: pre-line;
}

.rich-copy {
  color: var(--wp-text-muted);
  line-height: 1.8;
}

.rich-copy :deep(h2),
.rich-copy :deep(h3),
.rich-copy :deep(h4) {
  margin: 1.25rem 0 0.7rem;
  color: var(--wp-heading);
  line-height: 1.35;
}

.rich-copy :deep(h2:first-child),
.rich-copy :deep(h3:first-child),
.rich-copy :deep(h4:first-child) {
  margin-top: 0;
}

.rich-copy :deep(h2) {
  font-size: 1.28rem;
}

.rich-copy :deep(h3) {
  font-size: 1.05rem;
}

.rich-copy :deep(p) {
  margin: 0 0 0.9rem;
}

.rich-copy :deep(p:last-child) {
  margin-bottom: 0;
}

.rich-copy :deep(br) {
  content: '';
}

.rich-copy :deep(ul),
.rich-copy :deep(ol) {
  margin: 0 0 1rem;
  padding-left: 1.15rem;
}

.rich-copy :deep(li + li) {
  margin-top: 0.45rem;
}

.rich-copy :deep(strong),
.rich-copy :deep(b) {
  color: var(--wp-heading);
}

.rich-copy :deep(a) {
  color: var(--tenant-accent, var(--wp-blue));
  text-decoration: underline;
}

.rich-copy :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.75rem;
  cursor: zoom-in;
}

.rich-copy :deep(.whit) {
  white-space: nowrap;
  text-align: right;
}

.rich-copy :deep(table) {
  width: 100%;
  margin: 1rem 0 0;
  border-collapse: collapse;
  overflow: hidden;
  border: 1px solid var(--wp-border);
  border-radius: 0.8rem;
  background: #fff;
  font-size: 0.92rem;
}

.rich-copy :deep(table.canshu) {
  border-collapse: collapse;
}

.rich-copy :deep(table.canshu td),
.rich-copy :deep(table.canshu th),
.rich-copy :deep(table td),
.rich-copy :deep(table th) {
  padding: 0.7rem 0.85rem;
  vertical-align: top;
  border-bottom: 1px solid var(--wp-border);
}

.rich-copy :deep(table tr:last-child td),
.rich-copy :deep(table tr:last-child th) {
  border-bottom: 0;
}

.rich-copy :deep(table td:first-child),
.rich-copy :deep(table th:first-child) {
  width: 28%;
  color: var(--wp-heading);
  font-weight: 700;
  background: #f8fbfd;
}

.rich-copy :deep(table td) {
  overflow-wrap: anywhere;
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

.thumb-grid-single {
  grid-template-columns: 88px;
  justify-content: start;
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

.thumb-button img {
  display: block;
  width: 100%;
  aspect-ratio: 1 / 1;
  object-fit: contain;
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
  overflow: auto;
  touch-action: none;
}

.image-preview-dialog img {
  max-width: 100%;
  max-height: 88vh;
  object-fit: contain;
  border-radius: 1rem;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35);
  cursor: zoom-in;
  transform-origin: center center;
  transition: transform 0.18s ease;
  user-select: none;
}

.image-preview-dialog.zoomed {
  cursor: grab;
}

.image-preview-dialog.dragging {
  cursor: grabbing;
}

.image-preview-dialog img.zoomed {
  cursor: grab;
}

.image-preview-dialog img.dragging {
  cursor: grabbing;
}

.image-preview-toolbar {
  position: absolute;
  top: 1.25rem;
  right: 1.25rem;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.image-preview-action,
.image-preview-close {
  padding: 0.7rem 1rem;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.24);
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
  cursor: pointer;
}

.image-preview-action {
  min-width: 2.75rem;
}

.image-preview-scale {
  min-width: 3.8rem;
  text-align: center;
  color: #ffffff;
  font-size: 0.9rem;
  font-weight: 600;
}

.image-preview-close {
  position: static;
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

.basic-info-panel {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--tenant-accent, var(--wp-blue)) 12%, transparent), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(246, 250, 253, 0.98));
}

.basic-info-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1rem;
}

.basic-info-caption {
  margin-top: 0.45rem;
  max-width: 36rem;
  color: var(--wp-text-muted);
  line-height: 1.7;
}

.stock-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 92px;
  padding: 0.6rem 0.9rem;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.stock-indicator-ready {
  background: rgba(0, 163, 42, 0.12);
  color: #0d6b27;
}

.stock-indicator-low {
  background: rgba(219, 166, 23, 0.14);
  color: #8a6708;
}

.stock-indicator-soldout {
  background: rgba(214, 54, 56, 0.12);
  color: #a12622;
}

.basic-info-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-bottom: 1rem;
}

.basic-info-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.55rem 0.85rem;
  border: 1px solid rgba(34, 113, 177, 0.12);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.82);
  color: var(--wp-heading);
}

.basic-info-pill strong {
  color: var(--tenant-accent, var(--wp-blue));
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.basic-info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
}

.basic-info-item {
  display: grid;
  gap: 0.45rem;
  min-height: 88px;
  padding: 0.95rem 1rem;
  border: 1px solid rgba(16, 21, 23, 0.08);
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 10px 25px rgba(16, 21, 23, 0.04);
}

.basic-info-label {
  color: var(--wp-text-muted);
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.basic-info-value {
  color: var(--wp-heading);
  font-size: 1.05rem;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.variant-options {
  grid-template-columns: repeat(auto-fit, minmax(120px, max-content));
  justify-content: start;
  align-items: stretch;
}

.variant-button {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  border-radius: 999px;
  padding: 0.55rem 0.85rem;
  min-height: 2.35rem;
  max-width: min(100%, 240px);
  font-size: 0.82rem;
  line-height: 1.35;
  white-space: normal;
  text-align: left;
  overflow-wrap: anywhere;
  word-break: break-word;
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
  .basic-info-header {
    align-items: center;
    gap: 0.55rem;
    margin-bottom: 0.75rem;
    text-align: center;
  }

  .basic-info-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.65rem;
  }

  .basic-info-item {
    min-height: 0;
    padding: 0.72rem 0.8rem;
    gap: 0.28rem;
  }

  .basic-info-value {
    font-size: 0.94rem;
    line-height: 1.3;
  }

  .basic-info-caption {
    display: none;
  }

  .basic-info-pills {
    gap: 0.45rem;
    margin-bottom: 0.8rem;
  }

  .basic-info-pill {
    padding: 0.42rem 0.68rem;
    gap: 0.35rem;
    font-size: 0.78rem;
  }

  .basic-info-pill strong {
    font-size: 0.64rem;
  }

  .stock-indicator {
    min-width: 0;
    padding: 0.45rem 0.75rem;
    font-size: 0.74rem;
  }

  .detail-main-column > .basic-info-panel {
    padding: 0.95rem;
  }

  .detail-main-column > .basic-info-panel h2 {
    font-size: 1rem;
  }

  .section-kicker {
    margin-bottom: 0.3rem;
    font-size: 0.65rem;
  }

  .detail-copy,
  .rich-copy {
    font-size: 0.97rem;
    line-height: 1.95;
  }

  .rich-copy :deep(h2) {
    font-size: 1.16rem;
    margin: 1rem 0 0.65rem;
  }

  .rich-copy :deep(h3),
  .rich-copy :deep(h4) {
    font-size: 1rem;
    margin: 1rem 0 0.55rem;
  }

  .rich-copy :deep(p) {
    margin: 0 0 1rem;
  }

  .rich-copy :deep(li + li) {
    margin-top: 0.6rem;
  }

  .rich-copy :deep(table) {
    font-size: 0.86rem;
  }

  .rich-copy :deep(table td),
  .rich-copy :deep(table th) {
    padding: 0.62rem 0.68rem;
  }

  .related-list :deep(.card-actions .btn-primary) {
    min-height: 36px;
    padding: 0.55rem 0.75rem;
    font-size: 0.82rem;
  }
}
</style>
