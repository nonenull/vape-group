<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { createProductJsonLd, useSeo } from '@/composables/useSeo'
import { useCartStore } from '@/stores/cart'
import { useProductStore } from '@/stores/product'

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()
const cartStore = useCartStore()

const productId = computed(() => Number(route.params.id))
const product = computed(() => productStore.selectedProduct)
const selectedOptions = ref<Record<string, string>>({})
const selectedSkuVariant = computed(() => {
  if (!product.value?.skuVariants.length) {
    return null
  }
  return product.value.skuVariants.find((variant) =>
    product.value?.optionGroups.every((group) => selectedOptions.value[group.name] === variant.selections[group.name]),
  ) ?? null
})
const displayStock = computed(() => {
  if (!product.value) {
    return 0
  }
  return selectedSkuVariant.value?.stock ?? product.value.stock
})
const displayPrice = computed(() => {
  if (!product.value) {
    return 0
  }
  return selectedSkuVariant.value?.price ?? product.value.salePrice ?? product.value.price
})
const activeImage = ref('')

const relatedProducts = computed(() => {
  return productStore.products.filter((item) => item.id !== product.value?.id).slice(0, 3)
})

const addCurrentProduct = () => {
  if (product.value) {
    if (product.value.skuVariants.length && !selectedSkuVariant.value) {
      alert('請先選擇完整規格')
      return
    }
    const optionLabel = product.value.optionGroups
      .map((group) => `${group.name}：${selectedOptions.value[group.name]}`)
      .filter((item) => !item.endsWith('：'))
      .join(' / ')
    cartStore.addItem(product.value, selectedSkuVariant.value, optionLabel)
    router.push('/cart')
  }
}

const setActiveImage = (image: string) => {
  activeImage.value = image
}

onMounted(() => {
  productStore.fetchProductDetail(productId.value)
  if (!productStore.products.length) {
    productStore.fetchProducts()
  }
})

watch(product, (value) => {
  activeImage.value = value?.image ?? value?.gallery[0] ?? ''
  if (value?.optionGroups.length) {
    selectedOptions.value = Object.fromEntries(
      value.optionGroups.map((group) => [group.name, group.values[0] ?? '']),
    )
  } else {
    selectedOptions.value = {}
  }
}, { immediate: true })

watch(productId, (id) => {
  productStore.fetchProductDetail(id)
})

useSeo(computed(() => {
  if (!product.value) {
    return {
      title: '商品詳情 | Vape Group 商城',
      description: '檢視商品詳情、規格、價格與購買資訊。',
      canonicalPath: `/products/${productId.value}`,
      type: 'product' as const,
    }
  }

  const description = product.value.description || product.value.longDescription || `${product.value.name} 商品詳情`

  return {
    title: `${product.value.name} | ${product.value.category} | Vape Group 商城`,
    description,
    image: product.value.image,
    canonicalPath: `/products/${productId.value}`,
    type: 'product' as const,
    jsonLd: createProductJsonLd({
      name: product.value.name,
      description,
      image: product.value.gallery.length ? product.value.gallery : [product.value.image],
      sku: product.value.sku,
      category: product.value.category,
      price: displayPrice.value,
      availability: product.value.stock > 0,
      url: window.location.href,
      rating: product.value.rating,
      reviews: product.value.reviews,
    }),
  }
}))
</script>

<template>
  <section v-if="product" class="product-detail-page">
    <p class="breadcrumb">首頁 / 商品目錄 / {{ product.name }}</p>

    <div class="detail-layout panel">
      <div class="media-panel">
        <div class="image-stage">
          <img :src="activeImage || product.image" :alt="product.name" />
          <span v-if="product.badge" class="badge">{{ product.badge }}</span>
        </div>
        <div v-if="product.gallery.length" class="thumb-grid">
          <button
            v-for="(image, index) in product.gallery"
            :key="`${image}-${index}`"
            type="button"
            class="thumb-button"
            :class="{ active: (activeImage || product.image) === image }"
            @click="setActiveImage(image)"
          >
            <img :src="image" :alt="`${product.name}-thumb-${index}`" />
          </button>
        </div>
      </div>

      <div class="content-panel">
        <p class="category">{{ product.category }}</p>
        <h1>{{ product.name }}</h1>
        <p class="description">{{ product.description }}</p>
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
        <div class="meta-grid">
          <div>
            <span>SKU</span>
            <strong>{{ selectedSkuVariant?.sku ?? product.sku }}</strong>
          </div>
          <div>
            <span>現貨</span>
            <strong>{{ displayStock }}</strong>
          </div>
        </div>
        <p class="long-copy">{{ product.longDescription }}</p>

        <div v-if="product.flavors.length" class="chip-group">
          <span v-for="flavor in product.flavors" :key="flavor" class="chip">{{ flavor }}</span>
        </div>

        <div class="actions">
          <button class="primary" type="button" @click="addCurrentProduct">加入購物車</button>
          <RouterLink class="secondary" to="/products">回商品列表</RouterLink>
        </div>
      </div>
    </div>

    <div class="detail-sections">
      <article class="panel">
        <h2>商品規格</h2>
        <div class="spec-list">
          <div v-for="spec in product.specs" :key="spec.label" class="spec-row">
            <span>{{ spec.label }}</span>
            <strong>{{ spec.value }}</strong>
          </div>
        </div>
      </article>

      <article v-if="product.detailImages.length" class="panel">
        <h2>商品詳情圖</h2>
        <div class="detail-images">
          <img
            v-for="(image, index) in product.detailImages"
            :key="`${image}-${index}`"
            :src="image"
            :alt="`${product.name}-detail-${index}`"
          />
        </div>
      </article>

      <article class="panel">
        <div class="section-head">
          <h2>推薦商品</h2>
          <RouterLink to="/products">瀏覽更多</RouterLink>
        </div>
        <div class="related-list">
          <RouterLink v-for="item in relatedProducts" :key="item.id" :to="`/products/${item.id}`" class="related-item">
            <strong>{{ item.name }}</strong>
            <span>{{ item.category }}</span>
            <small>NT$ {{ (item.salePrice ?? item.price).toFixed(2) }}</small>
          </RouterLink>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.product-detail-page {
  display: grid;
  gap: 1rem;
}

.breadcrumb,
.category,
.description,
.long-copy,
.spec-row span,
.related-item span,
.related-item small {
  color: var(--wp-text-muted);
}

.panel {
  background: var(--wp-surface);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1.5rem;
}

.variant-picker {
  display: grid;
  gap: 0.65rem;
}

.variant-group {
  display: grid;
  gap: 0.5rem;
}

.variant-picker > span,
.variant-picker > small {
  color: var(--wp-text-muted);
}

.variant-options {
  display: flex;
  gap: 0.625rem;
  flex-wrap: wrap;
}

.variant-button {
  min-height: 2.4rem;
  padding: 0.55rem 0.9rem;
  border-radius: 999px;
  border: 1px solid var(--wp-border-strong);
  background: #fff;
  color: var(--wp-heading);
  font-weight: 600;
  cursor: pointer;
}

.variant-button.active {
  background: var(--tenant-accent, var(--wp-blue));
  border-color: var(--tenant-accent, var(--wp-blue));
  color: #fff;
}

.detail-layout {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: 0.95fr 1.05fr;
}

.image-stage {
  position: relative;
  min-height: 420px;
  display: grid;
  place-items: center;
  background: linear-gradient(180deg, #f8fbfe 0%, #eef3f7 100%);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
}

.image-stage img {
  width: min(320px, 78%);
  max-height: 360px;
  object-fit: contain;
}

.thumb-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(auto-fit, minmax(72px, 1fr));
}

.thumb-button {
  padding: 0;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  background: #fff;
  overflow: hidden;
  cursor: pointer;
}

.thumb-button.active {
  border-color: var(--tenant-accent, var(--wp-blue));
  box-shadow: 0 0 0 2px rgba(34, 113, 177, 0.12);
}

.thumb-button img {
  width: 100%;
  height: 84px;
  object-fit: cover;
}

.badge {
  position: absolute;
  top: 1rem;
  left: 1rem;
  padding: 0.3rem 0.65rem;
  border-radius: 0.25rem;
  background: var(--wp-green);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
}

.content-panel h1 {
  font-size: clamp(2rem, 4vw, 2.8rem);
  line-height: 1.1;
  margin: 0.25rem 0 0.75rem;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  margin: 1rem 0;
}

.price-row strong {
  font-size: 1.8rem;
}

.price-row span {
  text-decoration: line-through;
  color: var(--wp-text-muted);
}

.meta-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, 1fr);
  margin: 1rem 0;
}

.meta-grid div,
.spec-row {
  padding: 0.85rem 1rem;
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
  border: 1px solid var(--wp-border);
}

.detail-images {
  display: grid;
  gap: 1rem;
}

.detail-images img {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.meta-grid span {
  display: block;
  color: var(--wp-text-muted);
  font-size: 0.8125rem;
}

.meta-grid strong,
.spec-row strong {
  display: block;
  margin-top: 0.2rem;
}

.chip-group {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin: 1rem 0;
}

.chip {
  padding: 0.45rem 0.7rem;
  border-radius: 999px;
  background: var(--tenant-surface, var(--wp-blue-soft));
  color: var(--tenant-accent, var(--wp-blue));
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1rem;
}

.primary,
.secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.7rem;
  padding: 0.75rem 1rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid var(--wp-border-strong);
}

.primary {
  background: var(--tenant-accent, var(--wp-blue));
  color: #fff;
  border-color: var(--tenant-accent, var(--wp-blue));
}

.secondary {
  background: #fff;
  color: var(--wp-heading);
}

.detail-sections {
  display: grid;
  gap: 1rem;
  grid-template-columns: 0.9fr 1.1fr;
}

.spec-list,
.related-list {
  display: grid;
  gap: 0.75rem;
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.related-item {
  display: grid;
  gap: 0.2rem;
  padding: 0.9rem 1rem;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: var(--wp-surface-soft);
}

@media (max-width: 980px) {
  .detail-layout,
  .detail-sections {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .meta-grid {
    grid-template-columns: 1fr;
  }
}
</style>
