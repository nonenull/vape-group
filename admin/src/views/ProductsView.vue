<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAdminStore } from '@/stores/admin'
import { adminAPI, resolveAssetURL } from '@/api/admin'
import type {
  ProductOptionGroupRecord,
  ProductOverrideRecord,
  ProductRecord,
  ProductSkuVariantRecord,
  ProductVariantRecord,
} from '@/data/adminMock'

const store = useAdminStore()

const emptyProduct = (): ProductRecord => ({
  id: 0,
  sku: '',
  baseName: '',
  basePrice: 0,
  baseStockQuantity: 0,
  category: '',
  categoryId: null,
  brand: '',
  brandId: null,
  previewImage: '/src/assets/logo.svg',
  gallery: ['/src/assets/logo.svg'],
  detailImages: ['/src/assets/logo.svg'],
  status: '草稿',
  flavors: [],
  variants: [],
  optionGroups: [],
  skuVariants: [],
  updatedAt: '',
})

const selectedProductId = ref(store.products[0]?.id ?? 0)
const selectedTenantId = ref(store.tenants[0]?.id ?? 0)
const isCreating = ref(false)
const selectedProductIds = ref<number[]>([])
const galleryInput = ref('')
const detailImagesInput = ref('')
const skuVariantInput = ref('')

const selectedProduct = computed(() => {
  return store.products.find((item) => item.id === selectedProductId.value) ?? null
})

const isAllProductsSelected = computed(() => {
  return store.products.length > 0 && selectedProductIds.value.length === store.products.length
})

const selectedOverride = computed(() => {
  if (!selectedProduct.value) {
    return null
  }
  return store.getOverride(selectedProduct.value.id, selectedTenantId.value)
})

const categoryOptions = computed(() => store.getCategoriesByTenant(selectedTenantId.value))
const brandOptions = computed(() => store.getBrandsByTenant(selectedTenantId.value))

const productForm = ref<ProductRecord>(store.products[0] ? { ...store.products[0] } : emptyProduct())
const overrideForm = ref<ProductOverrideRecord>({
  id: 0,
  tenantId: selectedTenantId.value,
  productId: selectedProduct.value?.id ?? 0,
  customName: '',
  customDescription: '',
  customPrice: null,
  customStockQuantity: null,
  seoTitle: '',
  seoDescription: '',
  isVisible: true,
})

function formatSkuVariantInput(variants: ProductSkuVariantRecord[]) {
  return variants
    .map((item) => {
      const selectionText = Object.entries(item.selections)
        .map(([key, value]) => `${key}:${value}`)
        .join('; ')
      const priceText = item.price == null ? '' : `|${item.price}`
      const stockText = item.stock == null ? '' : `|${item.stock}`
      return `${selectionText}|${item.sku}${priceText}${stockText}`
    })
    .join('\n')
}

watch(
  selectedProductId,
  (productId) => {
    if (isCreating.value) {
      return
    }
    const product = store.products.find((item) => item.id === productId)
    if (product) {
      productForm.value = {
        ...product,
        gallery: [...product.gallery],
        detailImages: [...product.detailImages],
        variants: [...(product.variants ?? [])],
        optionGroups: [...(product.optionGroups ?? [])],
        skuVariants: [...(product.skuVariants ?? [])],
      }
      galleryInput.value = product.gallery.join('\n')
      detailImagesInput.value = product.detailImages.join('\n')
      skuVariantInput.value = formatSkuVariantInput(product.skuVariants ?? [])
    }
  },
  { immediate: true },
)

watch(
  [selectedProduct, selectedOverride, selectedTenantId],
  ([product, override, tenantId]) => {
    if (!product) {
      return
    }

    overrideForm.value = {
      id: override?.id ?? 0,
      tenantId,
      productId: product.id,
      customName: override?.customName ?? '',
      customDescription: override?.customDescription ?? '',
      customPrice: override?.customPrice ?? null,
      customStockQuantity: override?.customStockQuantity ?? null,
      seoTitle: override?.seoTitle ?? '',
      seoDescription: override?.seoDescription ?? '',
      isVisible: override?.isVisible ?? true,
    }
  },
  { immediate: true },
)

function selectProduct(productId: number) {
  isCreating.value = false
  selectedProductId.value = productId
}

function toggleProductSelection(productId: number, checked: boolean) {
  if (checked) {
    if (!selectedProductIds.value.includes(productId)) {
      selectedProductIds.value = [...selectedProductIds.value, productId]
    }
    return
  }
  selectedProductIds.value = selectedProductIds.value.filter((id) => id !== productId)
}

function toggleSelectAllProducts(checked: boolean) {
  selectedProductIds.value = checked ? store.products.map((item) => item.id) : []
}

function startCreateProduct() {
  isCreating.value = true
  selectedProductId.value = 0
  productForm.value = emptyProduct()
  galleryInput.value = productForm.value.gallery.join('\n')
  detailImagesInput.value = productForm.value.detailImages.join('\n')
  skuVariantInput.value = ''
}

function syncGalleryFromInput() {
  const items = galleryInput.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
  productForm.value.gallery = items.length ? items : [productForm.value.previewImage]
}

function syncDetailImagesFromInput() {
  const items = detailImagesInput.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
  productForm.value.detailImages = items.length ? items : [productForm.value.previewImage]
}

function syncGalleryInput() {
  galleryInput.value = productForm.value.gallery.join('\n')
}

function syncDetailImagesInput() {
  detailImagesInput.value = productForm.value.detailImages.join('\n')
}

function syncVariantConfigFromInput() {
  const skuVariants = skuVariantInput.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item): ProductSkuVariantRecord | null => {
      const [selectionPart, skuPart, pricePart, stockPart] = item.split('|').map((part) => part.trim())
      if (!selectionPart || !skuPart) {
        return null
      }
      const selections = Object.fromEntries(
        selectionPart
          .split(';')
          .map((entry) => entry.trim())
          .filter(Boolean)
          .map((entry) => {
            const [key, value] = entry.split(':')
            return [key?.trim() ?? '', value?.trim() ?? '']
          })
          .filter(([key, value]) => key && value),
      )
      if (!Object.keys(selections).length) {
        return null
      }
      const price = pricePart ? Number(pricePart) : null
      const stock = stockPart ? Number(stockPart) : null
      return {
        sku: skuPart,
        price: Number.isFinite(price) ? price : null,
        stock: Number.isFinite(stock) ? stock : null,
        selections,
      }
    })
    .filter((item): item is ProductSkuVariantRecord => Boolean(item))

  const optionGroupMap = new Map<string, string[]>()
  for (const variant of skuVariants) {
    for (const [groupName, value] of Object.entries(variant.selections)) {
      const currentValues = optionGroupMap.get(groupName) ?? []
      if (!currentValues.includes(value)) {
        currentValues.push(value)
      }
      optionGroupMap.set(groupName, currentValues)
    }
  }

  const optionGroups: ProductOptionGroupRecord[] = Array.from(optionGroupMap.entries()).map(([name, values]) => ({
    name,
    values,
  }))

  productForm.value.optionGroups = optionGroups
  productForm.value.skuVariants = skuVariants

  const flavorGroup = optionGroups.find((item) => item.name === '口味')
  productForm.value.flavors = flavorGroup?.values ?? []
  productForm.value.variants = skuVariants
    .filter((item) => Boolean(item.selections['口味']))
    .map((item): ProductVariantRecord => ({
      name: item.selections['口味'] ?? '',
      sku: item.sku,
    }))
}

const displayImage = (value: string) => resolveAssetURL(value)

async function uploadFile(file: File) {
  const uploaded = await adminAPI.uploadImage(file)
  return uploaded.url
}

async function appendImages(files: FileList | null, field: 'gallery' | 'detailImages') {
  if (!files?.length) {
    return
  }

  const uploadedImages = await Promise.all(Array.from(files).map(uploadFile))
  const currentImages = field === 'gallery' ? productForm.value.gallery : productForm.value.detailImages
  const mergedImages = [...currentImages, ...uploadedImages.filter(Boolean)]

  if (field === 'gallery') {
    productForm.value.gallery = mergedImages
    if (!productForm.value.previewImage || productForm.value.previewImage === '/src/assets/logo.svg') {
      productForm.value.previewImage = mergedImages[0] ?? productForm.value.previewImage
    }
    syncGalleryInput()
    return
  }

  productForm.value.detailImages = mergedImages
  syncDetailImagesInput()
}

async function uploadGalleryImages(event: Event) {
  const target = event.target as HTMLInputElement
  try {
    await appendImages(target.files, 'gallery')
  } catch (error) {
    console.error('上傳商品圖失敗:', error)
    alert((error as Error).message)
  }
  target.value = ''
}

async function uploadDetailImages(event: Event) {
  const target = event.target as HTMLInputElement
  try {
    await appendImages(target.files, 'detailImages')
  } catch (error) {
    console.error('上傳詳情圖失敗:', error)
    alert((error as Error).message)
  }
  target.value = ''
}

function removeGalleryImage(index: number) {
  const removedImage = productForm.value.gallery[index]
  productForm.value.gallery.splice(index, 1)
  if (productForm.value.previewImage === removedImage) {
    productForm.value.previewImage = productForm.value.gallery[0] ?? ''
  }
  syncGalleryInput()
}

function removeDetailImage(index: number) {
  productForm.value.detailImages.splice(index, 1)
  syncDetailImagesInput()
}

function setPreviewImage(image: string) {
  productForm.value.previewImage = image
}

function syncCategorySelection() {
  const category = categoryOptions.value.find((item) => item.id === productForm.value.categoryId)
  productForm.value.category = category?.name ?? ''
}

function syncBrandSelection() {
  const brand = brandOptions.value.find((item) => item.id === productForm.value.brandId)
  productForm.value.brand = brand?.name ?? ''
}

async function saveProduct() {
  try {
    syncGalleryFromInput()
    syncDetailImagesFromInput()
    syncVariantConfigFromInput()

    if (isCreating.value || productForm.value.id === 0) {
      const created = await store.createProduct({
        sku: productForm.value.sku,
        baseName: productForm.value.baseName,
        basePrice: productForm.value.basePrice,
        baseStockQuantity: productForm.value.baseStockQuantity,
        category: productForm.value.category,
        categoryId: productForm.value.categoryId,
        brand: productForm.value.brand,
        brandId: productForm.value.brandId,
        previewImage: productForm.value.previewImage,
        gallery: productForm.value.gallery,
        detailImages: productForm.value.detailImages,
        status: productForm.value.status,
        description: productForm.value.description,
        longDescription: productForm.value.longDescription,
        badge: productForm.value.badge,
        rating: productForm.value.rating,
        reviews: productForm.value.reviews,
        flavors: productForm.value.flavors,
        variants: productForm.value.variants,
        optionGroups: productForm.value.optionGroups,
        skuVariants: productForm.value.skuVariants,
        isActive: productForm.value.isActive,
      })
      isCreating.value = false
      selectedProductId.value = created.id
      productForm.value = {
        ...created,
        gallery: [...created.gallery],
        detailImages: [...created.detailImages],
        variants: [...(created.variants ?? [])],
        optionGroups: [...(created.optionGroups ?? [])],
        skuVariants: [...(created.skuVariants ?? [])],
      }
      galleryInput.value = created.gallery.join('\n')
      detailImagesInput.value = created.detailImages.join('\n')
      skuVariantInput.value = formatSkuVariantInput(created.skuVariants ?? [])
      alert('商品已成功建立')
      return
    }

    await store.updateProduct({
      ...productForm.value,
      gallery: [...productForm.value.gallery],
      detailImages: [...productForm.value.detailImages],
      variants: [...(productForm.value.variants ?? [])],
      optionGroups: [...(productForm.value.optionGroups ?? [])],
      skuVariants: [...(productForm.value.skuVariants ?? [])],
    })
    alert('商品已成功儲存')
  } catch (error) {
    console.error('保存商品失敗:', error)
    alert('保存商品失敗: ' + (error as Error).message)
  }
}

async function removeProduct() {
  if (!selectedProduct.value) {
    return
  }
  if (!confirm('確定要刪除此商品嗎？')) {
    return
  }
  try {
    await store.deleteProduct(selectedProduct.value.id)
    isCreating.value = false
    const fallback = store.products[0]
    selectedProductId.value = fallback?.id ?? 0
    productForm.value = fallback
      ? {
        ...fallback,
        gallery: [...fallback.gallery],
        detailImages: [...fallback.detailImages],
        variants: [...(fallback.variants ?? [])],
        optionGroups: [...(fallback.optionGroups ?? [])],
        skuVariants: [...(fallback.skuVariants ?? [])],
      }
      : emptyProduct()
    galleryInput.value = productForm.value.gallery.join('\n')
    detailImagesInput.value = productForm.value.detailImages.join('\n')
    skuVariantInput.value = formatSkuVariantInput(productForm.value.skuVariants ?? [])
    alert('商品已成功刪除')
  } catch (error) {
    console.error('刪除商品失敗:', error)
    alert('刪除商品失敗: ' + (error as Error).message)
  }
}

async function applyBulkProductUpdate(payload: { status?: ProductRecord['status'], isActive?: boolean }, successMessage: string) {
  if (!selectedProductIds.value.length) {
    alert('請先選擇至少一個商品')
    return
  }

  try {
    await store.bulkUpdateProducts(selectedProductIds.value, payload)
    alert(successMessage)
  } catch (error) {
    console.error('批量更新商品失敗:', error)
    alert('批量更新商品失敗: ' + (error as Error).message)
  }
}

async function saveOverride() {
  try {
    await store.upsertOverride({ ...overrideForm.value })
    alert('租戶覆寫已成功儲存')
  } catch (error) {
    console.error('保存租戶覆寫失敗:', error)
    alert('保存租戶覆寫失敗: ' + (error as Error).message)
  }
}

onMounted(async () => {
  if (!store.products.length) {
    await store.fetchProducts()
  }
  if (!store.tenants.length) {
    await store.fetchTenants()
  }
  const fallback = store.products[0]
  if (fallback && !selectedProduct.value) {
    selectedProductId.value = fallback.id
    productForm.value = {
      ...fallback,
      gallery: [...fallback.gallery],
      detailImages: [...fallback.detailImages],
      variants: [...(fallback.variants ?? [])],
      optionGroups: [...(fallback.optionGroups ?? [])],
      skuVariants: [...(fallback.skuVariants ?? [])],
    }
    galleryInput.value = fallback.gallery.join('\n')
    detailImagesInput.value = fallback.detailImages.join('\n')
    skuVariantInput.value = formatSkuVariantInput(fallback.skuVariants ?? [])
  }
  const selectedTenant = store.tenants[0]
  if (selectedTenant) {
    selectedTenantId.value = selectedTenant.id
    await Promise.all([
      store.fetchCategories(selectedTenant.id),
      store.fetchBrands(selectedTenant.id),
    ])
    if (selectedProductId.value) {
      await store.fetchProductOverride(selectedProductId.value, selectedTenant.id)
    }
  }
})

watch(
  [selectedProductId, selectedTenantId],
  async ([productId, tenantId]) => {
    if (productId && tenantId && !isCreating.value) {
      await Promise.all([
        store.fetchProductOverride(productId, tenantId),
        store.fetchCategories(tenantId),
        store.fetchBrands(tenantId),
      ])
    }
  },
)
</script>

<template>
  <section class="products-page">
    <div class="page-heading">
      <div>
        <p class="label">Product Center</p>
        <h2>商品 CRUD、覆寫與預覽圖</h2>
        <p class="subcopy">現在可直接新增、修改、刪除全域商品，維護主圖與圖庫預覽，並繼續為不同租戶設定商品覆寫。</p>
      </div>
      <button class="primary" type="button" @click="startCreateProduct">新增商品</button>
    </div>

    <div class="metrics-grid">
      <article class="metric-card">
        <p>全域商品</p>
        <strong>{{ store.products.length }}</strong>
      </article>
      <article class="metric-card">
        <p>可見覆寫</p>
        <strong>{{ store.visibleOverrides.length }}</strong>
      </article>
      <article class="metric-card">
        <p>啟用租戶</p>
        <strong>{{ store.activeTenants.length }}</strong>
      </article>
    </div>

    <div class="workspace-grid">
      <article class="table-card">
        <div class="table-toolbar">
          <span>商品清單</span>
          <div class="bulk-toolbar">
            <label class="bulk-select-all">
              <input
                :checked="isAllProductsSelected"
                type="checkbox"
                @change="toggleSelectAllProducts(($event.target as HTMLInputElement).checked)"
              />
              <span>全選</span>
            </label>
            <span class="bulk-count">已選 {{ selectedProductIds.length }}</span>
            <button class="secondary" type="button" @click="applyBulkProductUpdate({ status: '上架中', isActive: true }, '已批量上架商品')">批量上架</button>
            <button class="secondary" type="button" @click="applyBulkProductUpdate({ status: '草稿', isActive: false }, '已批量設為草稿')">批量草稿</button>
            <button class="secondary" type="button" @click="applyBulkProductUpdate({ status: '缺貨', isActive: false }, '已批量設為缺貨')">批量缺貨</button>
            <button class="danger" type="button" @click="applyBulkProductUpdate({ isActive: false }, '已批量下架商品')">批量下架</button>
          </div>
        </div>
        <div class="product-list">
          <div
            v-for="product in store.products"
            :key="product.id"
            class="product-row"
            :class="{
              selected: product.id === selectedProductId && !isCreating,
              checked: selectedProductIds.includes(product.id),
            }"
          >
            <label class="product-select">
              <input
                :checked="selectedProductIds.includes(product.id)"
                type="checkbox"
                @change="toggleProductSelection(product.id, ($event.target as HTMLInputElement).checked)"
              />
            </label>
            <button type="button" class="product-main" @click="selectProduct(product.id)">
              <img :src="displayImage(product.previewImage)" :alt="product.baseName" class="product-thumb" />
              <div class="product-copy">
                <strong>{{ product.baseName }}</strong>
                <p>{{ product.sku }} · {{ product.category }}</p>
                <small>NT$ {{ product.basePrice }} · 庫存 {{ product.baseStockQuantity }}</small>
              </div>
              <span class="status-pill">{{ product.status }}</span>
            </button>
          </div>
        </div>
      </article>

      <div class="editor-stack">
        <article class="editor-card">
          <div class="card-heading">
            <h3>{{ isCreating ? '新增商品' : '全域商品資料' }}</h3>
            <small v-if="!isCreating">最後更新 {{ productForm.updatedAt }}</small>
          </div>

          <div class="preview-grid">
            <div class="main-preview">
              <img :src="displayImage(productForm.previewImage)" :alt="productForm.baseName || 'product preview'" />
            </div>
            <div class="gallery-preview">
              <div
                v-for="(image, index) in productForm.gallery"
                :key="`gallery-${image}-${index}`"
                class="image-tile"
              >
                <img
                  :src="displayImage(image)"
                  :alt="`${productForm.baseName || 'product'}-gallery-${index}`"
                />
                <div class="image-actions">
                  <button class="secondary" type="button" @click="setPreviewImage(image)">設為主圖</button>
                  <button class="ghost-danger" type="button" @click="removeGalleryImage(index)">移除</button>
                </div>
              </div>
            </div>
          </div>
          <div class="preview-grid" v-if="productForm.detailImages && productForm.detailImages.length">
            <h4>詳情圖預覽</h4>
            <div class="detail-preview">
              <div
                v-for="(image, index) in productForm.detailImages"
                :key="`detail-${image}-${index}`"
                class="image-tile detail-tile"
              >
                <img
                  :src="displayImage(image)"
                  :alt="`${productForm.baseName || 'product'}-detail-${index}`"
                />
                <div class="image-actions">
                  <button class="ghost-danger" type="button" @click="removeDetailImage(index)">移除</button>
                </div>
              </div>
            </div>
          </div>

          <div class="form-grid">
            <label>
              <span>SKU</span>
              <input v-model="productForm.sku" />
            </label>
            <label>
              <span>分類</span>
              <select v-model="productForm.categoryId" @change="syncCategorySelection">
                <option :value="null">請選擇分類</option>
                <option v-for="category in categoryOptions" :key="category.id" :value="category.id">
                  {{ category.name }}
                </option>
              </select>
            </label>
            <label>
              <span>品牌</span>
              <select v-model="productForm.brandId" @change="syncBrandSelection">
                <option :value="null">請選擇品牌</option>
                <option v-for="brand in brandOptions" :key="brand.id" :value="brand.id">
                  {{ brand.name }}
                </option>
              </select>
            </label>
            <label class="full">
              <span>基礎商品名稱</span>
              <input v-model="productForm.baseName" />
            </label>
            <label>
              <span>基礎價格</span>
              <input v-model.number="productForm.basePrice" type="number" min="0" step="0.01" />
            </label>
            <label>
              <span>基礎庫存</span>
              <input v-model.number="productForm.baseStockQuantity" type="number" min="0" />
            </label>
            <label class="full">
              <span>主圖 URL</span>
              <input v-model="productForm.previewImage" />
            </label>
            <label class="full">
              <span>圖庫 URL（每行一張）</span>
              <textarea v-model="galleryInput" rows="4" @change="syncGalleryFromInput"></textarea>
              <input type="file" multiple accept="image/*" @change="uploadGalleryImages" />
            </label>
            <label class="full">
              <span>詳情圖 URL（每行一張）</span>
              <textarea v-model="detailImagesInput" rows="4" @change="syncDetailImagesFromInput"></textarea>
              <input type="file" multiple accept="image/*" @change="uploadDetailImages" />
            </label>
            <label class="full">
              <span>組合 SKU（每行：群組:值; 群組:值 | SKU | 價格 | 庫存）</span>
              <textarea
                v-model="skuVariantInput"
                rows="5"
                placeholder="例如&#10;口味:冰葡萄; 盒裝:2入|POD-GRAPE-2|299|12&#10;口味:冰葡萄; 盒裝:4入|POD-GRAPE-4|499|8&#10;&#10;規格群組會依這裡的組合自動推導"
                @change="syncVariantConfigFromInput"
              ></textarea>
            </label>
            <label class="full">
              <span>狀態</span>
              <select v-model="productForm.status">
                <option value="上架中">上架中</option>
                <option value="草稿">草稿</option>
                <option value="缺貨">缺貨</option>
              </select>
            </label>
          </div>

          <div class="actions">
            <button v-if="!isCreating" class="danger" type="button" @click="removeProduct">刪除商品</button>
            <button class="primary" type="button" @click="saveProduct">{{ isCreating ? '建立商品' : '儲存全域商品' }}</button>
          </div>
        </article>

        <article class="editor-card">
          <div class="card-heading split">
            <h3>租戶商品覆寫</h3>
            <select v-model.number="selectedTenantId">
              <option v-for="tenant in store.tenants" :key="tenant.id" :value="tenant.id">
                {{ tenant.name }}
              </option>
            </select>
          </div>
          <div class="form-grid">
            <label class="full">
              <span>自訂商品名稱</span>
              <input v-model="overrideForm.customName" />
            </label>
            <label class="full">
              <span>自訂商品描述</span>
              <textarea v-model="overrideForm.customDescription" rows="4"></textarea>
            </label>
            <label>
              <span>租戶售價</span>
              <input v-model.number="overrideForm.customPrice" type="number" min="0" step="0.01" />
            </label>
            <label>
              <span>租戶庫存</span>
              <input v-model.number="overrideForm.customStockQuantity" type="number" min="0" />
            </label>
            <label class="full">
              <span>SEO 標題</span>
              <input v-model="overrideForm.seoTitle" />
            </label>
            <label class="full">
              <span>SEO 描述</span>
              <textarea v-model="overrideForm.seoDescription" rows="3"></textarea>
            </label>
          </div>
          <label class="toggle">
            <input v-model="overrideForm.isVisible" type="checkbox" />
            <span>此租戶站點顯示這個商品</span>
          </label>
          <div class="actions">
            <button class="primary" type="button" @click="saveOverride">儲存租戶覆寫</button>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.products-page {
  display: grid;
  gap: 0.875rem;
}

.page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.label {
  color: var(--wp-blue);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: 0.75rem;
}

.page-heading h2 {
  margin: 0.35rem 0 0.45rem;
}

.subcopy {
  color: var(--wp-text-muted);
  max-width: 74ch;
}

.metrics-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(3, 1fr);
}

.metric-card,
.table-card,
.editor-card {
  background: #fff;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
}

.metric-card {
  padding: 0.8rem 1rem;
}

.metric-card p {
  color: var(--wp-text-muted);
}

.metric-card strong {
  display: block;
  margin-top: 0.25rem;
  font-size: 1.3rem;
}

.workspace-grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: minmax(420px, 0.95fr) minmax(0, 1.05fr);
}

.table-card {
  overflow: hidden;
}

.table-toolbar,
.editor-card {
  padding: 0.85rem 1rem;
}

.table-toolbar {
  display: grid;
  gap: 0.75rem;
}

.bulk-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.bulk-select-all {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.bulk-select-all input {
  width: auto;
  min-height: auto;
}

.bulk-count {
  color: var(--wp-text-muted);
  font-size: 0.875rem;
  margin-right: 0.25rem;
}

.product-list {
  display: grid;
  gap: 0.5rem;
}

.product-row {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.625rem;
  align-items: center;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.375rem;
  background: var(--wp-surface-soft);
}

.product-row.selected {
  border-color: var(--wp-blue);
  background: #f0f6fc;
}

.product-row.checked {
  box-shadow: inset 0 0 0 1px rgba(34, 113, 177, 0.18);
}

.product-select {
  display: inline-flex;
  align-items: center;
}

.product-select input {
  width: auto;
  min-height: auto;
}

.product-main {
  display: grid;
  grid-template-columns: 56px 1fr auto;
  gap: 0.75rem;
  align-items: center;
  border: 0;
  background: transparent;
  padding: 0;
  text-align: left;
}

.product-thumb {
  width: 56px;
  height: 56px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.product-copy strong {
  display: block;
  font-size: 0.95rem;
  line-height: 1.3;
}

.product-copy p,
.product-copy small {
  color: var(--wp-text-muted);
  font-size: 0.82rem;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  min-height: 1.7rem;
  padding: 0 0.55rem;
  border-radius: 999px;
  background: #eef3f7;
  color: var(--wp-text);
  font-weight: 700;
  font-size: 0.75rem;
}

.editor-stack {
  display: grid;
  gap: 0.875rem;
}

.card-heading {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 0.75rem;
}

.card-heading h3 {
  margin: 0;
}

.card-heading small {
  color: var(--wp-text-muted);
  font-size: 0.8rem;
}

.card-heading.split select {
  min-width: 220px;
}

.preview-grid {
  display: grid;
  gap: 0.625rem;
  margin-bottom: 0.75rem;
}

.main-preview {
  border: 1px solid var(--wp-border);
  border-radius: 0.75rem;
  overflow: hidden;
  background: var(--wp-surface-soft);
}

.main-preview img {
  width: 100%;
  height: 180px;
  object-fit: cover;
}

.gallery-preview {
  display: grid;
  gap: 0.5rem;
  grid-template-columns: repeat(auto-fit, minmax(64px, 1fr));
}

.gallery-preview img,
.detail-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.detail-preview {
  display: grid;
  gap: 0.5rem;
}

.image-tile {
  display: grid;
  gap: 0.4rem;
}

.image-tile img {
  min-height: 64px;
}

.detail-tile img {
  min-height: 140px;
}

.image-actions {
  display: flex;
  gap: 0.4rem;
}

.form-grid {
  display: grid;
  gap: 0.7rem;
  grid-template-columns: repeat(2, 1fr);
}

label {
  display: grid;
  gap: 0.3rem;
}

label span {
  font-weight: 600;
  font-size: 0.84rem;
}

.full {
  grid-column: 1 / -1;
}

input,
select,
textarea {
  width: 100%;
  min-height: 2.2rem;
  padding: 0.52rem 0.65rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 0.375rem;
  background: #fff;
  font-size: 0.92rem;
}

textarea {
  min-height: auto;
  resize: vertical;
}

.toggle {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-top: 0.75rem;
}

.toggle input {
  width: auto;
  min-height: auto;
}

.actions {
  display: flex;
  justify-content: space-between;
  gap: 0.625rem;
  margin-top: 0.75rem;
}

.primary,
.danger,
.secondary,
.ghost-danger {
  min-height: 2.15rem;
  padding: 0.5rem 0.8rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid transparent;
  font-size: 0.88rem;
}

.primary {
  background: var(--wp-blue);
  color: #fff;
  border-color: var(--wp-blue);
}

.danger {
  background: #fff;
  color: var(--wp-red);
  border-color: rgba(214, 54, 56, 0.3);
}

.secondary {
  background: #fff;
  color: var(--wp-blue);
  border-color: rgba(34, 113, 177, 0.24);
}

.ghost-danger {
  background: #fff7f7;
  color: var(--wp-red);
  border-color: rgba(214, 54, 56, 0.18);
}

@media (max-width: 1100px) {
  .page-heading,
  .workspace-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 780px) {
  .metrics-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
