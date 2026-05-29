<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Product, ProductSkuVariant } from '~/types/store'
import { fetchProductDetail } from '~/composables/useStoreApi'
import { buildProductPath } from '~/composables/useProductSlug'

const props = withDefaults(defineProps<{
  product: Product
  showDetailButton?: boolean
}>(), {
  showDetailButton: true,
})

const cartStore = useCartStore()

const displayPrice = computed(() => props.product.salePrice ?? props.product.price)
const productHref = computed(() => buildProductPath(props.product))
const showVariantPicker = ref(false)
const variantSourceProduct = ref<Product>(props.product)
const selectedOptions = ref<Record<string, string>>({})
const selectedSkuVariant = computed<ProductSkuVariant | null>(() => {
  if (!variantSourceProduct.value.skuVariants.length) {
    return null
  }

  return variantSourceProduct.value.skuVariants.find((variant) =>
    variantSourceProduct.value.optionGroups.every((group) => selectedOptions.value[group.name] === variant.selections[group.name]),
  ) ?? null
})
const selectedPrice = computed(() => selectedSkuVariant.value?.price ?? displayPrice.value)

const initializeSelections = () => {
  selectedOptions.value = Object.fromEntries(
    variantSourceProduct.value.optionGroups.map((group) => [group.name, group.values[0] ?? '']),
  )
}

const ensureVariantSourceProduct = async () => {
  if (variantSourceProduct.value.skuVariants.length || variantSourceProduct.value.optionGroups.length) {
    return variantSourceProduct.value
  }

  try {
    const detail = await fetchProductDetail(props.product.id)
    if (detail) {
      variantSourceProduct.value = detail
    }
  } catch (error) {
    console.error('載入商品規格失敗:', error)
    variantSourceProduct.value = props.product
  }

  return variantSourceProduct.value
}

const closeVariantPicker = () => {
  showVariantPicker.value = false
}

const confirmVariantAddToCart = async () => {
  if (!selectedSkuVariant.value) {
    window.alert('請先選擇完整規格')
    return
  }

  const optionLabel = variantSourceProduct.value.optionGroups
    .map((group) => `${group.name}：${selectedOptions.value[group.name]}`)
    .filter((item) => !item.endsWith('：'))
    .join(' / ')

  cartStore.addItem(variantSourceProduct.value, selectedSkuVariant.value, optionLabel)
  closeVariantPicker()
}

const addToCart = async () => {
  const sourceProduct = await ensureVariantSourceProduct()
  if (sourceProduct.skuVariants.length) {
    initializeSelections()
    showVariantPicker.value = true
    return
  }

  cartStore.addItem(sourceProduct)
}
</script>

<template>
  <article class="product-card">
    <NuxtLink :to="productHref" class="media-link">
      <div class="media">
        <img :src="product.image" :alt="product.name">
        <span v-if="product.badge" class="badge">{{ product.badge }}</span>
      </div>
    </NuxtLink>

    <div class="body">
      <NuxtLink :to="productHref" class="title-link">
        <h3>{{ product.name }}</h3>
      </NuxtLink>

      <div class="product-footer">
        <div class="price-group">
          <strong class="price">NT$ {{ displayPrice.toFixed(2) }}</strong>
          <span v-if="product.salePrice" class="base-price">NT$ {{ product.price.toFixed(2) }}</span>
        </div>
      </div>

      <div class="card-actions" :class="{ 'single-action': !props.showDetailButton }">
        <button class="btn-primary" type="button" @click.stop="addToCart">加入購物車</button>
        <NuxtLink v-if="props.showDetailButton" :to="productHref" class="btn-secondary">商品詳情</NuxtLink>
      </div>
    </div>
  </article>

  <div v-if="showVariantPicker" class="variant-modal" @click="closeVariantPicker">
    <div class="variant-dialog" @click.stop>
      <div class="variant-dialog-head">
        <div>
          <p class="variant-dialog-label">選擇規格</p>
          <h4>{{ variantSourceProduct.name }}</h4>
        </div>
        <button class="variant-close" type="button" @click="closeVariantPicker">關閉</button>
      </div>

      <div class="variant-group-list">
        <div v-for="group in variantSourceProduct.optionGroups" :key="group.name" class="variant-group">
          <span>{{ group.name }}</span>
          <div class="variant-options">
            <button
              v-for="value in group.values"
              :key="`${group.name}-${value}`"
              type="button"
              class="variant-option-button"
              :class="{ active: selectedOptions[group.name] === value }"
              @click="selectedOptions[group.name] = value"
            >
              {{ value }}
            </button>
          </div>
        </div>
      </div>

      <div class="variant-summary">
        <strong>NT$ {{ selectedPrice.toFixed(2) }}</strong>
        <small>SKU：{{ selectedSkuVariant?.sku ?? variantSourceProduct.sku }}</small>
      </div>

      <div class="variant-actions">
        <button class="btn-secondary" type="button" @click="closeVariantPicker">取消</button>
        <button class="btn-primary" type="button" @click="confirmVariantAddToCart">加入購物車</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.product-card {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  overflow: hidden;
  background: #ffffff;
  box-shadow: var(--wp-shadow);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.product-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--wp-shadow-hover);
}

.media-link,
.title-link {
  color: inherit;
  text-decoration: none;
}

.media {
  position: relative;
  min-height: 190px;
  background: linear-gradient(180deg, #f8fbfe 0%, #eef3f7 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--wp-border);
}

.media img {
  max-height: 150px;
  width: min(150px, 68%);
  object-fit: contain;
  opacity: 0.92;
}

.badge {
  position: absolute;
  top: 0.875rem;
  left: 0.875rem;
  padding: 0.28rem 0.6rem;
  background: var(--wp-green);
  color: #ffffff;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.body {
  display: grid;
  gap: 0.85rem;
  padding: 1rem;
  height: 100%;
  min-width: 0;
}

.body h3 {
  color: var(--wp-heading);
  font-size: 1rem;
  min-width: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.body p {
  color: var(--wp-text-muted);
  line-height: 1.6;
  min-width: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.product-footer,
.card-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
}

.card-actions {
  margin-top: auto;
}

.card-actions.single-action .btn-primary {
  width: 100%;
}

.card-actions .btn-primary {
  min-height: 38px;
  padding: 0.58rem 0.8rem;
  font-size: 0.84rem;
}

.price-group {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.price {
  color: var(--wp-heading);
  font-size: 1.05rem;
}

.base-price {
  color: var(--wp-text-muted);
  text-decoration: line-through;
}

.variant-modal {
  position: fixed;
  inset: 0;
  background: rgba(16, 21, 23, 0.42);
  display: grid;
  place-items: center;
  padding: 1rem;
  z-index: 40;
}

.variant-dialog {
  width: min(100%, 520px);
  border-radius: 1rem;
  border: 1px solid var(--wp-border);
  background: #ffffff;
  box-shadow: var(--wp-shadow-hover);
  padding: 1.25rem;
  display: grid;
  gap: 1rem;
}

.variant-dialog-head,
.variant-actions,
.variant-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.variant-dialog-head > div {
  min-width: 0;
}

.variant-dialog-label,
.variant-summary small {
  color: var(--wp-text-muted);
}

.variant-group-list,
.variant-options {
  display: grid;
  gap: 0.75rem;
}

.variant-group {
  display: grid;
  gap: 0.5rem;
}

.variant-options {
  grid-template-columns: repeat(auto-fit, minmax(82px, 1fr));
}

.variant-option-button,
.variant-close {
  border: 1px solid var(--wp-border);
  background: #ffffff;
  border-radius: 999px;
  padding: 0.55rem 0.75rem;
  cursor: pointer;
}

.variant-option-button {
  min-height: 34px;
  font-size: 0.82rem;
  line-height: 1.2;
}

.variant-close {
  flex: 0 0 auto;
  white-space: nowrap;
}

.variant-option-button.active {
  border-color: var(--tenant-accent, var(--wp-blue));
  background: var(--wp-blue-soft);
  color: var(--tenant-accent, var(--wp-blue));
}
</style>
