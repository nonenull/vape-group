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

function cloneProduct(product: ProductRecord): ProductRecord {
  return {
    ...product,
    gallery: [...product.gallery],
    detailImages: [...product.detailImages],
    flavors: [...(product.flavors ?? [])],
    variants: (product.variants ?? []).map((variant) => ({ ...variant })),
    optionGroups: (product.optionGroups ?? []).map((group) => ({
      ...group,
      values: [...group.values],
    })),
    skuVariants: (product.skuVariants ?? []).map((variant) => ({
      ...variant,
      selections: { ...variant.selections },
    })),
  }
}

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
  description: '',
  longDescription: '',
  badge: '',
  rating: 0,
  reviews: 0,
  isActive: true,
  flavors: [],
  variants: [],
  optionGroups: [],
  skuVariants: [],
  updatedAt: '',
})

const selectedProductId = ref(store.products[0]?.id ?? 0)
const selectedTenantId = ref(store.tenants[0]?.id ?? 0)
const isCreating = ref(false)
const isProductModalOpen = ref(false)
const isPreviewModalOpen = ref(false)
const isOverrideModalOpen = ref(false)
const selectedProductIds = ref<number[]>([])
const bulkStatus = ref<ProductRecord['status']>('上架中')
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
const selectedTenant = computed(() => {
  return store.tenants.find((item) => item.id === selectedTenantId.value) ?? null
})

const categoryOptions = computed(() => store.getCategories())
const brandOptions = computed(() => store.getBrands())

const productForm = ref<ProductRecord>(store.products[0] ? cloneProduct(store.products[0]) : emptyProduct())
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

function hydrateProductForm(product: ProductRecord) {
  const cloned = cloneProduct(product)
  productForm.value = cloned
  galleryInput.value = cloned.gallery.join('\n')
  detailImagesInput.value = cloned.detailImages.join('\n')
  skuVariantInput.value = formatSkuVariantInput(cloned.skuVariants ?? [])
}

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
      hydrateProductForm(product)
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
  isProductModalOpen.value = true
  productForm.value = emptyProduct()
  galleryInput.value = productForm.value.gallery.join('\n')
  detailImagesInput.value = productForm.value.detailImages.join('\n')
  skuVariantInput.value = ''
}

function openEditProduct(productId: number) {
  const product = store.products.find((item) => item.id === productId)
  if (!product) {
    return
  }
  isCreating.value = false
  selectedProductId.value = productId
  hydrateProductForm(product)
  isProductModalOpen.value = true
}

function closeProductModal() {
  isProductModalOpen.value = false

  if (selectedProduct.value) {
    hydrateProductForm(selectedProduct.value)
  } else {
    productForm.value = emptyProduct()
    galleryInput.value = productForm.value.gallery.join('\n')
    detailImagesInput.value = productForm.value.detailImages.join('\n')
    skuVariantInput.value = ''
  }

  isCreating.value = false
}

function openPreviewModal() {
  if (!selectedProduct.value) {
    return
  }
  isPreviewModalOpen.value = true
}

function closePreviewModal() {
  isPreviewModalOpen.value = false
}

function openOverrideModal() {
  if (!selectedProduct.value) {
    return
  }
  isOverrideModalOpen.value = true
}

function closeOverrideModal() {
  isOverrideModalOpen.value = false
}

function openPreviewForProduct(productId: number) {
  selectProduct(productId)
  openPreviewModal()
}

function openOverrideForProduct(productId: number) {
  selectProduct(productId)
  openOverrideModal()
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
      isProductModalOpen.value = false
      selectedProductId.value = created.id
      hydrateProductForm(created)
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
    isProductModalOpen.value = false
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
  await removeProductById(selectedProduct.value.id)
}

async function removeProductById(productId: number) {
  const product = store.products.find((item) => item.id === productId)
  if (!product) {
    return
  }
  if (!confirm('確定要刪除此商品嗎？')) {
    return
  }
  try {
    await store.deleteProduct(productId)
    isCreating.value = false
    isProductModalOpen.value = false
    const fallback = store.products[0]
    selectedProductId.value = fallback?.id ?? 0
    if (fallback) {
      hydrateProductForm(fallback)
    } else {
      productForm.value = emptyProduct()
      galleryInput.value = productForm.value.gallery.join('\n')
      detailImagesInput.value = productForm.value.detailImages.join('\n')
      skuVariantInput.value = ''
    }
    alert('商品已成功刪除')
  } catch (error) {
    console.error('刪除商品失敗:', error)
    alert('刪除商品失敗: ' + (error as Error).message)
  }
}

async function removeSelectedProducts() {
  if (!selectedProductIds.value.length) {
    alert('請先選擇至少一個商品')
    return
  }

  if (!confirm(`確定要批量刪除這 ${selectedProductIds.value.length} 個商品嗎？此操作無法復原。`)) {
    return
  }

  try {
    const productIds = [...selectedProductIds.value]
    for (const productId of productIds) {
      await store.deleteProduct(productId)
    }

    selectedProductIds.value = []
    isCreating.value = false
    isProductModalOpen.value = false

    const fallback = store.products[0]
    selectedProductId.value = fallback?.id ?? 0
    if (fallback) {
      hydrateProductForm(fallback)
    } else {
      productForm.value = emptyProduct()
      galleryInput.value = productForm.value.gallery.join('\n')
      detailImagesInput.value = productForm.value.detailImages.join('\n')
      skuVariantInput.value = ''
    }

    alert('已批量刪除選中的商品')
  } catch (error) {
    console.error('批量刪除商品失敗:', error)
    alert('批量刪除商品失敗: ' + (error as Error).message)
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

function getBulkStatusPayload(status: ProductRecord['status']) {
  if (status === '上架中') {
    return {
      status,
      isActive: true,
    }
  }

  return {
    status,
    isActive: false,
  }
}

async function applyBulkStatusChange() {
  const status = bulkStatus.value
  await applyBulkProductUpdate(
    getBulkStatusPayload(status),
    `已批量將商品狀態修改為「${status}」`,
  )
}

async function saveOverride() {
  if (!selectedProduct.value) {
    alert('請先選擇商品')
    return
  }
  if (!selectedTenant.value) {
    alert('請先選擇有效租戶')
    return
  }

  try {
    await store.upsertOverride({
      ...overrideForm.value,
      tenantId: selectedTenant.value.id,
      productId: selectedProduct.value.id,
    })
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
    hydrateProductForm(fallback)
  }
  const selectedTenant = store.tenants[0]
  if (selectedTenant) {
    selectedTenantId.value = selectedTenant.id
    await Promise.all([
      store.fetchCategories(),
      store.fetchBrands(),
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
        store.fetchCategories(),
        store.fetchBrands(),
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
          <div class="table-toolbar-head">
            <span>商品清單</span>
          </div>
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
            <div class="bulk-status-group">
              <select v-model="bulkStatus" class="bulk-status-select">
                <option value="上架中">上架中</option>
                <option value="草稿">草稿</option>
                <option value="缺貨">缺貨</option>
              </select>
              <button class="secondary" type="button" @click="applyBulkStatusChange">批量修改狀態</button>
            </div>
            <button class="danger" type="button" @click="removeSelectedProducts">批量刪除</button>
          </div>
        </div>
        <div class="table-scroller">
          <table class="product-table">
            <thead>
              <tr>
                <th class="checkbox-col">
                  <input
                    :checked="isAllProductsSelected"
                    type="checkbox"
                    @change="toggleSelectAllProducts(($event.target as HTMLInputElement).checked)"
                  />
                </th>
                <th>商品</th>
                <th>SKU</th>
                <th>Slug</th>
                <th>分類 / 品牌</th>
                <th>價格</th>
                <th>庫存</th>
                <th>狀態</th>
                <th>更新時間</th>
                <th class="action-col">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="product in store.products"
                :key="product.id"
                :class="{
                  selected: product.id === selectedProductId && !isCreating,
                  checked: selectedProductIds.includes(product.id),
                }"
                @click="selectProduct(product.id)"
              >
                <td class="checkbox-col" @click.stop>
                  <input
                    :checked="selectedProductIds.includes(product.id)"
                    type="checkbox"
                    @change="toggleProductSelection(product.id, ($event.target as HTMLInputElement).checked)"
                  />
                </td>
                <td>
                  <div class="product-cell">
                    <img :src="displayImage(product.previewImage)" :alt="product.baseName" class="product-thumb" />
                    <div class="product-copy">
                      <strong>{{ product.baseName }}</strong>
                      <small>{{ product.description || '尚未填寫商品摘要' }}</small>
                    </div>
                  </div>
                </td>
                <td>{{ product.sku }}</td>
                <td class="slug-cell">{{ product.slug || '—' }}</td>
                <td>
                  <div class="meta-stack">
                    <span>{{ product.category || '未分類' }}</span>
                    <small>{{ product.brand || '未綁定品牌' }}</small>
                  </div>
                </td>
                <td>NT$ {{ product.basePrice }}</td>
              <td>{{ product.baseStockQuantity }}</td>
              <td><span class="status-pill">{{ product.status }}</span></td>
              <td>{{ product.updatedAt || '尚未更新' }}</td>
              <td class="action-col" @click.stop>
                  <div class="row-actions">
                    <button class="secondary small-button" type="button" @click="openPreviewForProduct(product.id)">預覽</button>
                    <button class="secondary small-button" type="button" @click="openOverrideForProduct(product.id)">覆寫</button>
                    <button class="secondary small-button" type="button" @click="openEditProduct(product.id)">編輯</button>
                    <button class="danger small-button" type="button" @click="removeProductById(product.id)">刪除</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </div>

    <div
      v-if="isPreviewModalOpen && selectedProduct"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="closePreviewModal"
    >
      <article class="modal-card preview-modal-card">
        <div class="card-heading modal-heading">
          <div>
            <h3>預覽商品內容</h3>
            <small>{{ selectedProduct.baseName }}</small>
          </div>
          <button class="secondary" type="button" @click="closePreviewModal">關閉</button>
        </div>

        <div class="modal-body">
          <div class="preview-grid">
            <div class="main-preview">
              <img :src="displayImage(selectedProduct.previewImage)" :alt="selectedProduct.baseName || 'product preview'" />
            </div>
            <div class="gallery-preview" v-if="selectedProduct.gallery?.length">
              <img
                v-for="(image, index) in selectedProduct.gallery"
                :key="`preview-gallery-${image}-${index}`"
                :src="displayImage(image)"
                :alt="`${selectedProduct.baseName || 'product'}-gallery-${index}`"
              />
            </div>
          </div>

          <div class="preview-copy-grid">
            <dl class="overview-details">
              <div>
                <dt>SKU</dt>
                <dd>{{ selectedProduct.sku }}</dd>
              </div>
              <div>
                <dt>Slug</dt>
                <dd>{{ selectedProduct.slug || '—' }}</dd>
              </div>
              <div>
                <dt>分類</dt>
                <dd>{{ selectedProduct.category || '未分類' }}</dd>
              </div>
              <div>
                <dt>品牌</dt>
                <dd>{{ selectedProduct.brand || '未綁定品牌' }}</dd>
              </div>
              <div>
                <dt>價格</dt>
                <dd>NT$ {{ selectedProduct.basePrice }}</dd>
              </div>
              <div>
                <dt>庫存</dt>
                <dd>{{ selectedProduct.baseStockQuantity }}</dd>
              </div>
              <div>
                <dt>狀態</dt>
                <dd>{{ selectedProduct.status }}</dd>
              </div>
            </dl>

            <div class="preview-copy-block">
              <h4>商品摘要</h4>
              <p class="overview-description">{{ selectedProduct.description || '尚未填寫商品簡介。' }}</p>
            </div>

            <div class="preview-copy-block" v-if="selectedProduct.longDescription">
              <h4>商品說明</h4>
              <p class="overview-description">{{ selectedProduct.longDescription }}</p>
            </div>

            <div class="preview-grid" v-if="selectedProduct.detailImages?.length">
              <h4>詳情圖預覽</h4>
              <div class="detail-preview">
                <img
                  v-for="(image, index) in selectedProduct.detailImages"
                  :key="`preview-detail-${image}-${index}`"
                  :src="displayImage(image)"
                  :alt="`${selectedProduct.baseName || 'product'}-detail-${index}`"
                />
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>

    <div
      v-if="isOverrideModalOpen && selectedProduct"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="closeOverrideModal"
    >
      <article class="modal-card override-modal-card">
        <div class="card-heading modal-heading split">
          <h3>租戶商品覆寫</h3>
          <select v-model.number="selectedTenantId">
            <option :value="0">請選擇租戶</option>
            <option v-for="tenant in store.tenants" :key="tenant.id" :value="tenant.id">
              {{ tenant.name }}
            </option>
          </select>
        </div>

        <div class="modal-body">
          <div class="form-grid">
            <label class="full">
              <span>自訂商品名稱</span>
              <input v-model="overrideForm.customName" />
            </label>
            <label class="full">
              <span>自訂商品簡介</span>
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
        </div>

        <div class="actions modal-actions">
          <button class="secondary" type="button" @click="closeOverrideModal">取消</button>
          <button class="primary" type="button" :disabled="!selectedTenant || !selectedProduct" @click="saveOverride">
            儲存租戶覆寫
          </button>
        </div>
      </article>
    </div>

    <div
      v-if="isProductModalOpen"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="closeProductModal"
    >
      <article class="modal-card">
        <div class="card-heading modal-heading">
          <div>
            <h3>{{ isCreating ? '新增商品' : '修改商品' }}</h3>
            <small v-if="!isCreating">最後更新 {{ productForm.updatedAt || '尚未更新' }}</small>
          </div>
          <button class="secondary" type="button" @click="closeProductModal">關閉</button>
        </div>

        <div class="modal-body">
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
              <span>Slug</span>
              <input :value="productForm.slug || '儲存後自動產生'" readonly />
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
            <label class="full">
              <span>商品簡介</span>
              <textarea
                v-model="productForm.description"
                rows="3"
                placeholder="顯示在商品卡片與詳情頁標題下方的摘要"
              ></textarea>
            </label>
            <label class="full">
              <span>商品說明</span>
              <textarea
                v-model="productForm.longDescription"
                rows="6"
                placeholder="顯示在前台商品詳情區塊，支援多行內容"
              ></textarea>
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
        </div>

        <div class="actions modal-actions">
          <button v-if="!isCreating" class="danger" type="button" @click="removeProduct">刪除商品</button>
          <div class="action-group">
            <button class="secondary" type="button" @click="closeProductModal">取消</button>
            <button class="primary" type="button" @click="saveProduct">{{ isCreating ? '建立商品' : '儲存全域商品' }}</button>
          </div>
        </div>
      </article>
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
.editor-card,
.modal-card {
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
  display: block;
}

.table-card {
  overflow: hidden;
}

.table-toolbar,
.editor-card,
.modal-card {
  padding: 0.85rem 1rem;
}

.table-toolbar {
  display: grid;
  gap: 0.75rem;
}

.table-scroller {
  overflow-x: auto;
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

.bulk-status-group {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.bulk-status-select {
  width: auto;
  min-width: 132px;
}

.product-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 860px;
}

.product-table th,
.product-table td {
  padding: 0.8rem 0.85rem;
  border-top: 1px solid var(--wp-border);
  text-align: left;
  vertical-align: middle;
}

.product-table th {
  color: var(--wp-text-muted);
  font-size: 0.78rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: #f8fafc;
}

.product-table tbody tr {
  cursor: pointer;
  transition: background-color 0.18s ease;
}

.product-table tbody tr:hover {
  background: #f8fbff;
}

.product-table tbody tr.selected {
  background: #f0f6fc;
}

.product-table tbody tr.checked {
  box-shadow: inset 3px 0 0 var(--wp-blue);
}

.checkbox-col {
  width: 48px;
}

.action-col {
  width: 280px;
}

.checkbox-col input {
  width: auto;
  min-height: auto;
}

.product-cell {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr);
  gap: 0.75rem;
  align-items: center;
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
  margin-bottom: 0.15rem;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.product-copy small {
  color: var(--wp-text-muted);
  font-size: 0.82rem;
  line-height: 1.4;
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
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

.meta-stack {
  display: grid;
  gap: 0.15rem;
}

.meta-stack small {
  color: var(--wp-text-muted);
}

.slug-cell {
  max-width: 220px;
  color: var(--wp-text-muted);
  font-size: 0.82rem;
  word-break: break-word;
}

.row-actions {
  display: flex;
  gap: 0.45rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.table-toolbar-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
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

.preview-copy-grid {
  display: grid;
  gap: 1rem;
}

.preview-copy-block {
  display: grid;
  gap: 0.45rem;
}

.preview-copy-block h4 {
  margin: 0;
}

.overview-title small,
.overview-description {
  color: var(--wp-text-muted);
}

.overview-details {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.overview-details div {
  padding: 0.75rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.625rem;
  background: var(--wp-surface-soft);
}

.overview-details dt {
  margin: 0 0 0.25rem;
  color: var(--wp-text-muted);
  font-size: 0.8rem;
}

.overview-details dd {
  margin: 0;
  font-weight: 600;
}

.empty-state {
  padding: 1rem;
  border: 1px dashed var(--wp-border-strong);
  border-radius: 0.625rem;
  color: var(--wp-text-muted);
  background: var(--wp-surface-soft);
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

.action-group {
  display: flex;
  gap: 0.625rem;
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

.small-button {
  min-height: 1.95rem;
  padding: 0.35rem 0.65rem;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  overflow-y: auto;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
}

.modal-card {
  width: min(1080px, 100%);
  height: min(calc(100vh - 3rem), 100%);
  max-height: calc(100vh - 3rem);
  overflow: hidden;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  gap: 0;
}

.preview-modal-card {
  width: min(960px, 100%);
}

.override-modal-card {
  width: min(760px, 100%);
  height: auto;
  max-height: min(calc(100vh - 3rem), 820px);
}

.modal-heading {
  margin-bottom: 0;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid var(--wp-border);
}

.modal-body {
  min-height: 0;
  overflow-y: auto;
  scrollbar-gutter: stable;
  overscroll-behavior: contain;
  padding-top: 0.85rem;
}

.modal-actions {
  align-items: center;
  padding-top: 0.85rem;
  border-top: 1px solid var(--wp-border);
}

@media (max-width: 1100px) {
  .overview-details {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 780px) {
  .page-heading {
    flex-direction: column;
  }

  .metrics-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }

  .overview-details {
    grid-template-columns: 1fr;
  }

  .actions,
  .modal-actions {
    flex-direction: column;
  }

  .action-group {
    width: 100%;
    flex-direction: column;
  }

  .modal-overlay {
    padding: 0.75rem;
  }

  .modal-card {
    height: min(calc(100vh - 1.5rem), 100%);
    max-height: calc(100vh - 1.5rem);
  }
}
</style>
