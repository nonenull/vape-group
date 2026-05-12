<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { productAPI } from '@/api/products'
import { resolveAssetURL } from '@/api/client'
import { useCartStore } from '@/stores/cart'
import type { Product, ProductSkuVariant } from '@/data/mockProducts'

const props = defineProps<{ product: Product }>()
const router = useRouter()
const cartStore = useCartStore()

const displayPrice = computed(() => props.product.salePrice ?? props.product.price)
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

const mapDetailProduct = (input: any): Product => {
  const gallery =
    input.custom_images?.length
      ? input.custom_images
      : input.gallery?.length
        ? input.gallery
        : input.base_images?.length
          ? input.base_images
          : input.preview_image
            ? [input.preview_image]
            : [props.product.image]

  const detailImages =
    input.custom_detail_images?.length
      ? input.custom_detail_images
      : input.detail_images?.length
        ? input.detail_images
        : input.specifications?.detailImages?.length
          ? input.specifications.detailImages
          : props.product.detailImages

  const optionGroups = Array.isArray(input.option_groups ?? input.optionGroups)
    ? (input.option_groups ?? input.optionGroups)
      .map((item: any) => ({
        name: String(item?.name ?? '').trim(),
        values: Array.isArray(item?.values) ? item.values.map((value: any) => String(value).trim()).filter(Boolean) : [],
      }))
      .filter((item: Product['optionGroups'][number]) => item.name && item.values.length)
    : props.product.optionGroups

  const skuVariants = Array.isArray(input.sku_variants ?? input.skuVariants)
    ? (input.sku_variants ?? input.skuVariants)
      .map((item: any) => ({
        sku: String(item?.sku ?? '').trim(),
        price: item?.price == null ? null : Number(item.price),
        stock: item?.stock == null ? null : Number(item.stock),
        selections: item?.selections && typeof item.selections === 'object' ? item.selections : {},
      }))
      .filter((item: ProductSkuVariant) => item.sku && Object.keys(item.selections).length)
    : props.product.skuVariants

  return {
    ...props.product,
    image: resolveAssetURL(input.preview_image ?? gallery[0] ?? props.product.image),
    gallery: gallery.map(resolveAssetURL),
    detailImages: detailImages.map(resolveAssetURL),
    optionGroups,
    skuVariants,
  }
}

const ensureVariantSourceProduct = async () => {
  if (variantSourceProduct.value.skuVariants.length || variantSourceProduct.value.optionGroups.length) {
    return variantSourceProduct.value
  }

  try {
    const detail = await productAPI.getProductDetail(props.product.id)
    variantSourceProduct.value = mapDetailProduct(detail)
  } catch (error) {
    console.error('載入商品規格失敗:', error)
    variantSourceProduct.value = props.product
  }

  return variantSourceProduct.value
}

const goToDetail = () => {
  router.push(`/products/${props.product.id}`)
}

const closeVariantPicker = () => {
  showVariantPicker.value = false
}

const confirmVariantAddToCart = () => {
  if (!selectedSkuVariant.value) {
    alert('請先選擇完整規格')
    return
  }
  const optionLabel = variantSourceProduct.value.optionGroups
    .map((group) => `${group.name}：${selectedOptions.value[group.name]}`)
    .filter((item) => !item.endsWith('：'))
    .join(' / ')
  cartStore.addItem(variantSourceProduct.value, selectedSkuVariant.value, optionLabel)
  closeVariantPicker()
  router.push('/cart')
}

const addToCart = async () => {
  const sourceProduct = await ensureVariantSourceProduct()
  if (sourceProduct.skuVariants.length) {
    initializeSelections()
    showVariantPicker.value = true
    return
  }
  cartStore.addItem(sourceProduct)
  router.push('/cart')
}
</script>

<template>
  <article class="product-card" role="button" tabindex="0" @click="goToDetail" @keydown.enter="goToDetail" @keydown.space.prevent="goToDetail">
    <div class="media">
      <img :src="product.image" :alt="product.name" />
      <span v-if="product.badge" class="badge">{{ product.badge }}</span>
    </div>

    <div class="body">
      <h3>{{ product.name }}</h3>
      <p>{{ product.description }}</p>

      <div class="product-footer">
        <div class="price-group">
          <strong class="price">NT$ {{ displayPrice.toFixed(2) }}</strong>
          <span v-if="product.salePrice" class="base-price">NT$ {{ product.price.toFixed(2) }}</span>
        </div>
      </div>

      <div class="card-actions">
        <button class="btn-primary" type="button" @click.stop="addToCart">加入購物車</button>
        <button class="btn-secondary" type="button" @click.stop="goToDetail">商品詳情</button>
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
  cursor: pointer;
}

.product-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--wp-shadow-hover);
}

.product-card:focus-visible {
  outline: 2px solid var(--tenant-accent, var(--wp-blue));
  outline-offset: 2px;
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
  padding: 0.9rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  flex: 1;
}

h3 {
  color: var(--wp-heading);
  font-size: 1rem;
  line-height: 1.35;
}

p {
  color: var(--wp-text-muted);
  line-height: 1.55;
  font-size: 0.9rem;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-footer {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 1rem;
}

.price-group {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.price {
  color: var(--wp-heading);
  font-size: 1.15rem;
}

.base-price {
  color: var(--wp-text-muted);
  text-decoration: line-through;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: auto;
}

.btn-primary,
.btn-secondary {
  flex: 1;
  min-height: 2.05rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  padding: 0.5rem 0.68rem;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary {
  background: var(--tenant-accent, var(--wp-blue));
  color: white;
}

.btn-secondary {
  background: #fff;
  color: var(--wp-heading);
  border-color: var(--wp-border-strong);
}

.variant-modal {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(15, 23, 42, 0.48);
}

.variant-dialog {
  width: min(100%, 440px);
  max-height: min(80vh, 680px);
  overflow-y: auto;
  background: #fff;
  border-radius: 0.75rem;
  padding: 1rem;
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.22);
}

.variant-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.variant-dialog-label,
.variant-summary small,
.variant-group span {
  color: var(--wp-text-muted);
}

.variant-dialog-head h4 {
  margin-top: 0.2rem;
  font-size: 1rem;
  color: var(--wp-heading);
}

.variant-close {
  border: 0;
  background: transparent;
  color: var(--wp-text-muted);
  cursor: pointer;
}

.variant-group-list {
  display: grid;
  gap: 0.9rem;
  margin-top: 1rem;
}

.variant-group {
  display: grid;
  gap: 0.45rem;
}

.variant-options {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.variant-option-button {
  min-height: 2rem;
  padding: 0.4rem 0.75rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 999px;
  background: #fff;
  color: var(--wp-heading);
  cursor: pointer;
}

.variant-option-button.active {
  border-color: var(--tenant-accent, var(--wp-blue));
  background: color-mix(in srgb, var(--tenant-accent, var(--wp-blue)) 10%, #fff);
  color: var(--tenant-accent, var(--wp-blue));
}

.variant-summary {
  display: grid;
  gap: 0.2rem;
  margin-top: 1rem;
}

.variant-summary strong {
  color: var(--wp-heading);
  font-size: 1.08rem;
}

.variant-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}

@media (max-width: 720px) {
  h3 {
    font-size: 0.9rem;
    line-height: 1.28;
  }

  p {
    font-size: 0.84rem;
    line-height: 1.45;
    -webkit-line-clamp: 2;
  }

  .price {
    font-size: 1rem;
  }

  .btn-primary,
  .btn-secondary {
    min-height: 1.9rem;
    padding: 0.42rem 0.6rem;
    font-size: 0.84rem;
  }

  .btn-secondary {
    display: none;
  }

  .variant-dialog {
    padding: 0.9rem;
    border-radius: 0.65rem;
  }

  .variant-actions {
    flex-direction: column;
  }
}
</style>
