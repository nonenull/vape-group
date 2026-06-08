<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { Delete, Edit, Plus, Search, View } from '@element-plus/icons-vue'
import type { CheckboxValueType } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAdminStore } from '@/stores/admin'
import { adminAPI, resolveAssetURL } from '@/api/admin'
import type {
  BulkGeneratedOverrideNameRecord,
  ProductOptionGroupRecord,
  ProductOverrideRecord,
  ProductRecord,
  ProductSkuVariantRecord,
  ProductVariantRecord,
} from '@/data/adminMock'

const store = useAdminStore()

const slugUnsafeChars = /[^a-z0-9]+/g

function slugifyProductName(value: string) {
  const source = value.trim().toLowerCase()
  if (!source) {
    return 'product'
  }
  return source
    .replace(slugUnsafeChars, '-')
    .replace(/^-+|-+$/g, '') || 'product'
}

function cloneProduct(product: ProductRecord): ProductRecord {
  return {
    ...product,
    categoryIds: [...(product.categoryIds ?? (product.categoryId != null ? [product.categoryId] : []))],
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
  categoryIds: [],
  brand: '',
  brandId: null,
  previewImage: '/src/assets/logo.svg',
  gallery: ['/src/assets/logo.svg'],
  detailImages: ['/src/assets/logo.svg'],
  status: '草稿',
  description: '',
  longDescription: '',
  specificationHtml: '',
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
const bulkNameGenerationInstruction = ref('')
const isGeneratingBulkCustomNames = ref(false)
const bulkCustomNameResultMessage = ref('')
const bulkCustomNameResultTenants = ref<string[]>([])
const selectedProductIds = ref<number[]>([])
const selectedOverrideTenantIds = ref<number[]>([])
const categoryFilterDraftId = ref(0)
const brandFilterDraftId = ref(0)
const selectedCategoryFilterId = ref(0)
const selectedBrandFilterId = ref(0)
const searchDraftKeyword = ref('')
const appliedSearchKeyword = ref('')
const bulkStatus = ref<ProductRecord['status']>('上架中')
const galleryInput = ref('')
const skuVariantInput = ref('')
const categoryTreeRef = ref<any>(null)

function sortCategories(items: typeof store.categories) {
  return [...items].sort((a, b) => {
    if (a.sortOrder !== b.sortOrder) {
      return a.sortOrder - b.sortOrder
    }
    return a.id - b.id
  })
}

const selectedProduct = computed(() => {
  return store.products.find((item) => item.id === selectedProductId.value) ?? null
})

const overrideTenants = computed(() => store.tenants)
const isAllOverrideTenantsSelected = computed(() =>
  overrideTenants.value.length > 0 && selectedOverrideTenantIds.value.length === overrideTenants.value.length,
)

const selectedCategoryFilter = computed(() => {
  return categoryOptions.value.find((item) => item.id === selectedCategoryFilterId.value) ?? null
})

const selectedBrandFilter = computed(() => {
  return brandOptions.value.find((item) => item.id === selectedBrandFilterId.value) ?? null
})

const visibleSelectedProductIds = computed(() => {
  const visibleIds = new Set(filteredProducts.value.map((item) => item.id))
  return selectedProductIds.value.filter((id) => visibleIds.has(id))
})

const isAllProductsSelected = computed(() => {
  return filteredProducts.value.length > 0 && visibleSelectedProductIds.value.length === filteredProducts.value.length
})

const selectedTenant = computed(() => {
  return store.tenants.find((item) => item.id === selectedTenantId.value) ?? null
})

const categoryOptions = computed(() => store.getCategories())
const brandOptions = computed(() => store.getBrands())
const categoriesByParent = computed(() => {
  const byParent = new Map<number | null, typeof categoryOptions.value>()
  for (const category of categoryOptions.value) {
    const key = category.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(category)
    byParent.set(key, bucket)
  }

  for (const [key, bucket] of byParent.entries()) {
    byParent.set(key, sortCategories(bucket))
  }

  return byParent
})
const categoryTreeOptions = computed(() => {
  const result: Array<{ id: number; label: string }> = []

  const walk = (parentId: number | null, ancestorHasNext: boolean[]) => {
    const items = categoriesByParent.value.get(parentId) ?? []
    items.forEach((item, index) => {
      const isLast = index === items.length - 1
      const guides = ancestorHasNext
        .map((hasNext) => (hasNext ? '│  ' : '   '))
        .join('')
      const branch = ancestorHasNext.length ? `${isLast ? '└─ ' : '├─ '}` : '● '

      result.push({
        id: item.id,
        label: `${guides}${branch}${item.name}`,
      })

      walk(item.id, [...ancestorHasNext, !isLast])
    })
  }

  walk(null, [])
  return result
})
const categoryNameMap = computed(() => new Map(categoryOptions.value.map((item) => [item.id, item.name])))
const categoryTreeList = computed(() => {
  const result: Array<{
    id: number
    name: string
    depth: number
    isLast: boolean
    ancestorHasNext: boolean[]
  }> = []

  const walk = (parentId: number | null, depth: number, ancestorHasNext: boolean[]) => {
    const items = categoriesByParent.value.get(parentId) ?? []
    items.forEach((item, index) => {
      const isLast = index === items.length - 1
      result.push({
        id: item.id,
        name: item.name,
        depth,
        isLast,
        ancestorHasNext,
      })
      walk(item.id, depth + 1, [...ancestorHasNext, !isLast])
    })
  }

  walk(null, 0, [])
  return result
})
const categoryTreeData = computed(() => {
  type CategoryTreeNode = { id: number; label: string; children: CategoryTreeNode[] }
  const buildNodes = (parentId: number | null): CategoryTreeNode[] => {
    const items = categoriesByParent.value.get(parentId) ?? []
    return items.map((item) => ({
      id: item.id,
      label: item.name,
      children: buildNodes(item.id),
    }))
  }

  return buildNodes(null)
})
const selectedCategoryTagList = computed(() => {
  const selectedIds = productForm.value.categoryIds?.length
    ? productForm.value.categoryIds
    : productForm.value.categoryId != null
      ? [productForm.value.categoryId]
      : []

  return selectedIds.map((id) => ({
    id,
    name: categoryNameMap.value.get(id) ?? `#${id}`,
    isPrimary: productForm.value.categoryId === id,
  }))
})
const categoryDescendantIdsMap = computed(() => {
  const result = new Map<number, Set<number>>()

  const collectDescendants = (categoryId: number): Set<number> => {
    const cached = result.get(categoryId)
    if (cached) {
      return cached
    }

    const ids = new Set<number>([categoryId])
    const children = categoriesByParent.value.get(categoryId) ?? []
    for (const child of children) {
      for (const descendantId of collectDescendants(child.id)) {
        ids.add(descendantId)
      }
    }

    result.set(categoryId, ids)
    return ids
  }

  for (const category of categoryOptions.value) {
    collectDescendants(category.id)
  }

  return result
})
const categoryProductCountMap = computed(() => {
  const normalizedKeyword = appliedSearchKeyword.value.trim().toLowerCase()
  const result = new Map<number, number>()

  for (const category of categoryOptions.value) {
    const descendantIds = categoryDescendantIdsMap.value.get(category.id) ?? new Set([category.id])
    const count = store.products.filter((product) => {
      const productCategoryIds = product.categoryIds?.length
        ? product.categoryIds
        : product.categoryId != null
          ? [product.categoryId]
          : []
      const brandMatches =
        selectedBrandFilterId.value === 0 ||
        product.brandId === selectedBrandFilterId.value ||
        (!!selectedBrandFilter.value && product.brand === selectedBrandFilter.value.name)

      const searchMatches =
        !normalizedKeyword ||
        [
          product.baseName,
          product.sku,
          product.slug,
          product.category,
          product.brand,
          product.description,
        ].some((value) => String(value ?? '').toLowerCase().includes(normalizedKeyword))

      return brandMatches && searchMatches && productCategoryIds.some((id) => descendantIds.has(id))
    }).length

    result.set(category.id, count)
  }

  return result
})
const filteredProducts = computed(() => {
  return store.products.filter((product) => {
    const selectedCategoryIds =
      selectedCategoryFilterId.value === 0
        ? null
        : categoryDescendantIdsMap.value.get(selectedCategoryFilterId.value) ?? new Set([selectedCategoryFilterId.value])
    const productCategoryIds = product.categoryIds?.length
      ? product.categoryIds
      : product.categoryId != null
        ? [product.categoryId]
        : []
    const normalizedKeyword = appliedSearchKeyword.value.trim().toLowerCase()
    const searchMatches =
      !normalizedKeyword ||
      [
        product.baseName,
        product.sku,
        product.slug,
        product.category,
        product.brand,
        product.description,
      ].some((value) => String(value ?? '').toLowerCase().includes(normalizedKeyword))

    const categoryMatches =
      selectedCategoryFilterId.value === 0 ||
      productCategoryIds.some((id) => selectedCategoryIds?.has(id)) ||
      (!!selectedCategoryFilter.value && product.category === selectedCategoryFilter.value.name)

    const brandMatches =
      selectedBrandFilterId.value === 0 ||
      product.brandId === selectedBrandFilterId.value ||
      (!!selectedBrandFilter.value && product.brand === selectedBrandFilter.value.name)

    return categoryMatches && brandMatches && searchMatches
  })
})

const productForm = ref<ProductRecord>(store.products[0] ? cloneProduct(store.products[0]) : emptyProduct())
const overrideForms = ref<Record<number, ProductOverrideRecord>>({})

function hydrateProductForm(product: ProductRecord) {
  const cloned = cloneProduct(product)
  productForm.value = cloned
  galleryInput.value = cloned.gallery.join('\n')
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
  [selectedProduct, () => store.tenants, () => store.overrides],
  ([product]) => {
    if (!product) {
      overrideForms.value = {}
      return
    }

    const nextForms: Record<number, ProductOverrideRecord> = {}
    for (const tenant of store.tenants) {
      const override = store.getOverride(product.id, tenant.id)
      nextForms[tenant.id] = {
        id: override?.id ?? 0,
        tenantId: tenant.id,
        productId: product.id,
        customName: override?.customName ?? '',
        customDescription: override?.customDescription ?? '',
        customPrice: override?.customPrice ?? null,
        customStockQuantity: override?.customStockQuantity ?? null,
        seoTitle: override?.seoTitle ?? '',
        seoDescription: override?.seoDescription ?? '',
        isVisible: override?.isVisible ?? true,
      }
    }
    overrideForms.value = nextForms
  },
  { immediate: true, deep: true },
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
  selectedProductIds.value = checked ? filteredProducts.value.map((item) => item.id) : []
}

function handleProductSelectionChange(rows: ProductRecord[]) {
  selectedProductIds.value = rows.map((row) => row.id)
}

function handleCurrentProductChange(row: ProductRecord | undefined) {
  if (!row) {
    return
  }
  selectProduct(row.id)
}

function toggleOverrideTenantSelection(tenantId: number, checked: boolean) {
  if (checked) {
    if (!selectedOverrideTenantIds.value.includes(tenantId)) {
      selectedOverrideTenantIds.value = [...selectedOverrideTenantIds.value, tenantId]
    }
    return
  }
  selectedOverrideTenantIds.value = selectedOverrideTenantIds.value.filter((id) => id !== tenantId)
}

function toggleSelectAllOverrideTenants(checked: boolean) {
  selectedOverrideTenantIds.value = checked ? overrideTenants.value.map((tenant) => tenant.id) : []
}

function applyFilters() {
  selectedCategoryFilterId.value = categoryFilterDraftId.value
  selectedBrandFilterId.value = brandFilterDraftId.value
}

function applySearch() {
  appliedSearchKeyword.value = searchDraftKeyword.value.trim()
}

function submitToolbarFilters() {
  applyFilters()
  applySearch()
}

function resetFilters() {
  categoryFilterDraftId.value = 0
  brandFilterDraftId.value = 0
  selectedCategoryFilterId.value = 0
  selectedBrandFilterId.value = 0
  searchDraftKeyword.value = ''
  appliedSearchKeyword.value = ''
}

function clearSearch() {
  searchDraftKeyword.value = ''
  appliedSearchKeyword.value = ''
}

function startCreateProduct() {
  isCreating.value = true
  isProductModalOpen.value = true
  productForm.value = emptyProduct()
  galleryInput.value = productForm.value.gallery.join('\n')
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

function openStorefrontProduct(product: ProductRecord) {
  const slug = product.slug?.trim() || slugifyProductName(product.baseName)
  const path = `/products/${product.id}-${slug}`
  const base = window.location.origin.replace(/\/fuck\/?$/, '')
  window.open(`${base}${path}`, '_blank', 'noopener,noreferrer')
}

function syncGalleryFromInput() {
  const items = galleryInput.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
  productForm.value.gallery = items.length ? items : [productForm.value.previewImage]
}

function syncGalleryInput() {
  galleryInput.value = productForm.value.gallery.join('\n')
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

async function appendImages(files: FileList | null) {
  if (!files?.length) {
    return
  }

  const uploadedImages = await Promise.all(Array.from(files).map(uploadFile))
  const currentImages = productForm.value.gallery
  const mergedImages = [...currentImages, ...uploadedImages.filter(Boolean)]

  productForm.value.gallery = mergedImages
  if (!productForm.value.previewImage || productForm.value.previewImage === '/src/assets/logo.svg') {
    productForm.value.previewImage = mergedImages[0] ?? productForm.value.previewImage
  }
  syncGalleryInput()
}

async function uploadGalleryImages(event: Event) {
  const target = event.target as HTMLInputElement
  try {
    await appendImages(target.files)
  } catch (error) {
    console.error('上傳商品圖失敗:', error)
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

function setPreviewImage(image: string) {
  const currentIndex = productForm.value.gallery.findIndex((item) => item === image)
  productForm.value.previewImage = image
  if (currentIndex > 0) {
    const [selectedImage] = productForm.value.gallery.splice(currentIndex, 1)
    if (selectedImage) {
      productForm.value.gallery.unshift(selectedImage)
    }
    syncGalleryInput()
  }
}

function syncCategorySelection() {
  const category = categoryOptions.value.find((item) => item.id === productForm.value.categoryId)
  productForm.value.category = category?.name ?? ''
  const categoryId = productForm.value.categoryId
  if (categoryId == null) {
    productForm.value.categoryIds = []
    return
  }
  const categoryIds = productForm.value.categoryIds ?? []
  if (!categoryIds.includes(categoryId)) {
    productForm.value.categoryIds = [categoryId, ...categoryIds]
  }
}

function syncAdditionalCategorySelection() {
  const categoryIds = [...new Set((productForm.value.categoryIds ?? []).filter((id) => Number.isFinite(id) && id > 0))]
  productForm.value.categoryIds = categoryIds

  if (!categoryIds.length) {
    productForm.value.categoryId = null
    productForm.value.category = ''
    return
  }

  if (productForm.value.categoryId == null || !categoryIds.includes(productForm.value.categoryId)) {
    productForm.value.categoryId = categoryIds[0]
  }
  syncCategorySelection()
}

function isCategoryChecked(categoryId: number) {
  return (productForm.value.categoryIds ?? []).includes(categoryId)
}

function toggleProductCategory(categoryId: number, checked: boolean) {
  const currentIds = new Set(productForm.value.categoryIds ?? [])

  if (checked) {
    currentIds.add(categoryId)
  } else {
    currentIds.delete(categoryId)
  }

  productForm.value.categoryIds = [...currentIds]
  syncAdditionalCategorySelection()
}

function removeProductCategory(categoryId: number) {
  toggleProductCategory(categoryId, false)
}

async function syncCategoryTreeSelection() {
  await nextTick()
  if (!categoryTreeRef.value) {
    return
  }
  categoryTreeRef.value.setCheckedKeys(productForm.value.categoryIds ?? [])
}

function handleCategoryTreeCheck() {
  if (!categoryTreeRef.value) {
    return
  }
  productForm.value.categoryIds = categoryTreeRef.value.getCheckedKeys(false)
  syncAdditionalCategorySelection()
}

function formatProductCategorySummary(product: ProductRecord) {
  const categoryIds = product.categoryIds?.length
    ? product.categoryIds
    : product.categoryId != null
      ? [product.categoryId]
      : []

  if (!categoryIds.length) {
    return '未分類'
  }

  return categoryIds
    .map((id) => categoryNameMap.value.get(id) ?? `#${id}`)
    .join(' / ')
}

function getProductCategoryDisplay(product: ProductRecord) {
  const categoryIds = product.categoryIds?.length
    ? product.categoryIds
    : product.categoryId != null
      ? [product.categoryId]
      : []

  if (!categoryIds.length) {
    return {
      primary: '未分類',
      extras: [] as string[],
    }
  }

  const names = categoryIds
    .map((id) => categoryNameMap.value.get(id) ?? `#${id}`)
    .filter(Boolean)

  return {
    primary: names[0] ?? '未分類',
    extras: names.slice(1),
  }
}

function syncBrandSelection() {
  const brand = brandOptions.value.find((item) => item.id === productForm.value.brandId)
  productForm.value.brand = brand?.name ?? ''
}

function getFallbackProductId() {
  return filteredProducts.value[0]?.id ?? store.products[0]?.id ?? 0
}

async function saveProduct() {
  try {
    syncGalleryFromInput()
    syncVariantConfigFromInput()
    syncAdditionalCategorySelection()

    if (isCreating.value || productForm.value.id === 0) {
      const created = await store.createProduct({
        sku: productForm.value.sku,
        baseName: productForm.value.baseName,
        basePrice: productForm.value.basePrice,
        baseStockQuantity: productForm.value.baseStockQuantity,
        category: productForm.value.category,
        categoryId: productForm.value.categoryId,
        categoryIds: productForm.value.categoryIds,
        brand: productForm.value.brand,
        brandId: productForm.value.brandId,
        previewImage: productForm.value.previewImage,
        gallery: productForm.value.gallery,
        detailImages: productForm.value.detailImages,
        status: productForm.value.status,
        description: productForm.value.description,
        longDescription: productForm.value.longDescription,
        specificationHtml: productForm.value.specificationHtml,
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
      categoryIds: [...(productForm.value.categoryIds ?? [])],
      gallery: [...productForm.value.gallery],
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
  try {
    await ElMessageBox.confirm(`確定要刪除商品「${product.baseName}」嗎？`, '刪除商品', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    await store.deleteProduct(productId)
    isCreating.value = false
    isProductModalOpen.value = false
    const fallback = store.products.find((item) => item.id === getFallbackProductId())
    selectedProductId.value = fallback?.id ?? 0
    if (fallback) {
      hydrateProductForm(fallback)
    } else {
      productForm.value = emptyProduct()
      galleryInput.value = productForm.value.gallery.join('\n')
      skuVariantInput.value = ''
    }
    ElMessage.success('商品已成功刪除')
  } catch (error) {
    console.error('刪除商品失敗:', error)
    ElMessage.error('刪除商品失敗: ' + (error as Error).message)
  }
}

async function removeSelectedProducts() {
  if (!selectedProductIds.value.length) {
    ElMessage.warning('請先選擇至少一個商品')
    return
  }

  try {
    await ElMessageBox.confirm(
      `確定要批量刪除這 ${selectedProductIds.value.length} 個商品嗎？此操作無法復原。`,
      '批量刪除商品',
      {
        type: 'warning',
        confirmButtonText: '刪除',
        cancelButtonText: '取消',
      },
    )
  } catch {
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

    const fallback = store.products.find((item) => item.id === getFallbackProductId())
    selectedProductId.value = fallback?.id ?? 0
    if (fallback) {
      hydrateProductForm(fallback)
    } else {
      productForm.value = emptyProduct()
      galleryInput.value = productForm.value.gallery.join('\n')
      skuVariantInput.value = ''
    }

    ElMessage.success('已批量刪除選中的商品')
  } catch (error) {
    console.error('批量刪除商品失敗:', error)
    ElMessage.error('批量刪除商品失敗: ' + (error as Error).message)
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

function getOverrideForm(tenantId: number) {
  return overrideForms.value[tenantId]
}

async function saveOverride(tenantId: number) {
  if (!selectedProduct.value) {
    alert('請先選擇商品')
    return
  }
  const tenant = store.tenants.find((item) => item.id === tenantId)
  if (!tenant) {
    alert('請先選擇有效租戶')
    return
  }
  const form = getOverrideForm(tenantId)
  if (!form) {
    alert('找不到租戶覆寫表單')
    return
  }

  try {
    await store.upsertOverride({
      ...form,
      tenantId: tenant.id,
      productId: selectedProduct.value.id,
    })
    const refreshed = store.getOverride(selectedProduct.value.id, tenant.id)
    if (refreshed) {
      overrideForms.value = {
        ...overrideForms.value,
        [tenant.id]: {
          ...refreshed,
          customImages: refreshed.customImages ?? [],
          customDetailImages: refreshed.customDetailImages ?? [],
        },
      }
    }
    alert(`「${tenant.name}」租戶覆寫已成功儲存`)
  } catch (error) {
    console.error('保存租戶覆寫失敗:', error)
    alert('保存租戶覆寫失敗: ' + (error as Error).message)
  }
}

async function bulkGenerateCustomNames() {
  if (!selectedTenant.value) {
    alert('請先選擇有效租戶')
    return
  }
  if (!selectedProductIds.value.length) {
    alert('請先選擇至少一個商品')
    return
  }

  isGeneratingBulkCustomNames.value = true
  try {
    const results = await adminAPI.bulkGenerateCustomNames(
      selectedProductIds.value,
      selectedTenant.value.id,
      bulkNameGenerationInstruction.value.trim(),
    )

    for (const item of results as BulkGeneratedOverrideNameRecord[]) {
      const existing = store.getOverride(item.productId, item.tenantId)
      await store.upsertOverride({
        id: existing?.id ?? 0,
        tenantId: item.tenantId,
        productId: item.productId,
        customName: item.customName,
        customDescription: existing?.customDescription ?? '',
        customPrice: existing?.customPrice ?? null,
        customStockQuantity: existing?.customStockQuantity ?? null,
        customImages: existing?.customImages ?? [],
        customDetailImages: existing?.customDetailImages ?? [],
        seoTitle: existing?.seoTitle ?? '',
        seoDescription: existing?.seoDescription ?? '',
        isVisible: existing?.isVisible ?? true,
      })
    }

    if (selectedProduct.value && selectedTenant.value) {
      const refreshed = store.getOverride(selectedProduct.value.id, selectedTenant.value.id)
      if (refreshed) {
        overrideForms.value = {
          ...overrideForms.value,
          [selectedTenant.value.id]: {
            ...refreshed,
            customImages: refreshed.customImages ?? [],
            customDetailImages: refreshed.customDetailImages ?? [],
          },
        }
      }
    }

    bulkCustomNameResultMessage.value = `已為 ${results.length} 個商品生成自訂商品名稱`
    alert(`已為 ${results.length} 個商品生成自訂商品名稱`)
  } catch (error) {
    bulkCustomNameResultMessage.value = ''
    console.error('批量生成自訂商品名稱失敗:', error)
    alert('批量生成自訂商品名稱失敗: ' + (error as Error).message)
  } finally {
    isGeneratingBulkCustomNames.value = false
  }
}

async function bulkGenerateOverrideNamesForCurrentProduct() {
  if (!selectedProduct.value) {
    alert('請先選擇商品')
    return
  }
  if (!selectedOverrideTenantIds.value.length) {
    alert('請先勾選至少一個租戶')
    return
  }

  isGeneratingBulkCustomNames.value = true
  bulkCustomNameResultMessage.value = ''
  bulkCustomNameResultTenants.value = []

  try {
    let successCount = 0

    const targetTenants = overrideTenants.value.filter((tenant) => selectedOverrideTenantIds.value.includes(tenant.id))

    for (const tenant of targetTenants) {
      const results = await adminAPI.bulkGenerateCustomNames(
        [selectedProduct.value.id],
        tenant.id,
        bulkNameGenerationInstruction.value.trim(),
      )

      for (const item of results as BulkGeneratedOverrideNameRecord[]) {
        const existing = store.getOverride(item.productId, item.tenantId)
        await store.upsertOverride({
          id: existing?.id ?? 0,
          tenantId: item.tenantId,
          productId: item.productId,
          customName: item.customName,
          customDescription: existing?.customDescription ?? '',
          customPrice: existing?.customPrice ?? null,
          customStockQuantity: existing?.customStockQuantity ?? null,
          customImages: existing?.customImages ?? [],
          customDetailImages: existing?.customDetailImages ?? [],
          seoTitle: existing?.seoTitle ?? '',
          seoDescription: existing?.seoDescription ?? '',
          isVisible: existing?.isVisible ?? true,
        })
        successCount += 1
      }
    }

    for (const tenant of targetTenants) {
      const refreshed = store.getOverride(selectedProduct.value.id, tenant.id)
      if (refreshed) {
        overrideForms.value = {
          ...overrideForms.value,
          [tenant.id]: {
            ...refreshed,
            customImages: refreshed.customImages ?? [],
            customDetailImages: refreshed.customDetailImages ?? [],
          },
        }
      }
    }

    bulkCustomNameResultMessage.value = `已為勾選的 ${successCount} 個租戶生成自訂商品名稱`
    bulkCustomNameResultTenants.value = targetTenants.map((tenant) => tenant.name)
  } catch (error) {
    console.error('批量生成目前商品租戶名稱失敗:', error)
    bulkCustomNameResultMessage.value = ''
    bulkCustomNameResultTenants.value = []
    alert('批量生成目前商品租戶名稱失敗: ' + (error as Error).message)
  } finally {
    isGeneratingBulkCustomNames.value = false
  }
}

onMounted(async () => {
  if (!store.products.length) {
    await store.fetchProducts()
  }
  if (!store.tenants.length) {
    await store.fetchTenants()
  }
  await Promise.all([
    store.fetchCategories(),
    store.fetchBrands(),
  ])

  const fallback = store.products.find((item) => item.id === getFallbackProductId())
  if (fallback && !selectedProduct.value) {
    selectedProductId.value = fallback.id
    hydrateProductForm(fallback)
  }
  const selectedTenant = store.tenants[0]
  if (selectedTenant) {
    selectedTenantId.value = selectedTenant.id
  }
  if (selectedProductId.value) {
    await Promise.all(store.tenants.map((tenant) => store.fetchProductOverride(selectedProductId.value, tenant.id)))
  }
})

watch(
  [selectedProductId, () => store.tenants.length],
  async ([productId]) => {
    if (productId && store.tenants.length && !isCreating.value) {
      await Promise.all(store.tenants.map((tenant) => store.fetchProductOverride(productId, tenant.id)))
    }
  },
)

watch(
  filteredProducts,
  (products) => {
    const visibleIds = new Set(products.map((item) => item.id))
    selectedProductIds.value = selectedProductIds.value.filter((id) => visibleIds.has(id))

    if (isCreating.value || !products.length) {
      if (!products.length && !isCreating.value) {
        selectedProductId.value = 0
      }
      return
    }

    if (!visibleIds.has(selectedProductId.value)) {
      const fallback = products[0]
      selectedProductId.value = fallback?.id ?? 0
      if (fallback) {
        hydrateProductForm(fallback)
      }
    }
  },
  { immediate: true },
)

watch(
  () => [productForm.value.categoryId, ...(productForm.value.categoryIds ?? [])],
  () => {
    syncCategoryTreeSelection()
  },
)
</script>

<template>
  <section class="products-page">
    <div class="element-page-heading">
      <div>
        <p class="label">Product Center</p>
        <h2>商品 CRUD、覆寫與預覽圖</h2>
        <p class="subcopy">現在可直接新增、修改、刪除全域商品，維護主圖與圖庫預覽，並繼續為不同租戶設定商品覆寫。</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="startCreateProduct">新增商品</el-button>
    </div>

    <el-row :gutter="16" class="metrics-grid">
      <el-col :md="8" :sm="24">
        <el-card shadow="never">
          <p class="metric-label">全域商品</p>
          <strong class="metric-value">{{ store.products.length }}</strong>
        </el-card>
      </el-col>
      <el-col :md="8" :sm="24">
        <el-card shadow="never">
          <p class="metric-label">篩選結果</p>
          <strong class="metric-value">{{ filteredProducts.length }}</strong>
        </el-card>
      </el-col>
      <el-col :md="8" :sm="24">
        <el-card shadow="never">
          <p class="metric-label">可見覆寫</p>
          <strong class="metric-value">{{ store.visibleOverrides.length }}</strong>
        </el-card>
      </el-col>
    </el-row>

    <div class="workspace-grid">
      <el-card class="table-card" shadow="never">
        <template #header>
          <div class="card-header">
            <div>
              <span class="title">商品清單</span>
              <small>{{ filteredProducts.length }} / {{ store.products.length }} 件商品</small>
            </div>
            <div class="toolbar-actions">
              <el-select v-model="categoryFilterDraftId" class="toolbar-field toolbar-field-wide">
                <el-option :value="0" :label="`全部分类 (${filteredProducts.length})`" />
                <el-option
                  v-for="category in categoryTreeOptions"
                  :key="category.id"
                  :value="category.id"
                  :label="`${category.label} (${categoryProductCountMap.get(category.id) ?? 0})`"
                />
              </el-select>
              <el-select v-model="brandFilterDraftId" class="toolbar-field">
                <el-option :value="0" label="全部品牌" />
                <el-option v-for="brand in brandOptions" :key="brand.id" :value="brand.id" :label="brand.name" />
              </el-select>
              <el-input
                v-model="searchDraftKeyword"
                placeholder="搜索商品名称、SKU、Slug"
                class="toolbar-field toolbar-search-field"
                @keyup.enter="submitToolbarFilters"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
              <el-button @click="submitToolbarFilters">搜索 / 篩選</el-button>
              <el-button @click="resetFilters">清空</el-button>
            </div>
          </div>
        </template>

        <div class="table-toolbar">
          <div class="bulk-toolbar">
            <div class="bulk-meta-group">
              <span>已選 <strong class="bulk-count-badge">{{ selectedProductIds.length }}</strong></span>
            </div>
            <span class="filter-summary" v-if="selectedCategoryFilterId || selectedBrandFilterId || appliedSearchKeyword">
              分类: {{ selectedCategoryFilter?.name || '全部' }} · 品牌: {{ selectedBrandFilter?.name || '全部' }}
              <template v-if="appliedSearchKeyword"> · 搜索: {{ appliedSearchKeyword }}</template>
            </span>
            <el-select v-model="bulkStatus" class="bulk-status-field">
              <el-option value="上架中" label="上架中" />
              <el-option value="草稿" label="草稿" />
              <el-option value="缺貨" label="缺貨" />
            </el-select>
            <el-button @click="applyBulkStatusChange">批量修改狀態</el-button>
            <el-button type="danger" :icon="Delete" @click="removeSelectedProducts">批量刪除</el-button>
          </div>
        </div>
        <div class="product-table-shell">
          <el-table
            :data="filteredProducts"
            stripe
            highlight-current-row
            row-key="id"
            class="product-admin-table"
            @selection-change="handleProductSelectionChange"
            @current-change="handleCurrentProductChange"
          >
            <el-table-column type="selection" width="50" fixed="left" />
            <el-table-column label="商品" min-width="220">
              <template #default="{ row }">
                <div class="product-cell">
                  <img :src="displayImage(row.previewImage)" :alt="row.baseName" class="product-thumb" />
                  <div class="product-copy">
                    <strong>{{ row.baseName }}</strong>
                    <small>{{ row.description || '尚未填寫商品摘要' }}</small>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="sku" label="SKU" min-width="120" />
            <el-table-column label="Slug" min-width="130">
              <template #default="{ row }">
                <span class="slug-text">{{ row.slug || '—' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="分類" min-width="150">
              <template #default="{ row }">
                <div class="category-cell">
                  <strong>{{ getProductCategoryDisplay(row).primary }}</strong>
                  <div v-if="getProductCategoryDisplay(row).extras.length" class="category-chip-list">
                    <span
                      v-for="extraCategory in getProductCategoryDisplay(row).extras"
                      :key="`${row.id}-${extraCategory}`"
                      class="category-chip"
                    >
                      {{ extraCategory }}
                    </span>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="品牌" min-width="110">
              <template #default="{ row }">
                {{ row.brand || '未綁定品牌' }}
              </template>
            </el-table-column>
            <el-table-column label="價格" min-width="90">
              <template #default="{ row }">
                NT$ {{ row.basePrice }}
              </template>
            </el-table-column>
            <el-table-column prop="baseStockQuantity" label="庫存" width="90" />
            <el-table-column label="狀態" width="110">
              <template #default="{ row }">
                <el-tag size="small" effect="light">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="更新時間" min-width="120">
              <template #default="{ row }">
                {{ row.updatedAt || '尚未更新' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="240" fixed="right">
              <template #default="{ row }">
              <div class="row-actions">
                <el-button size="small" type="success" @click.stop="openStorefrontProduct(row)">前台查看</el-button>
                <el-button size="small" :icon="View" @click.stop="openPreviewForProduct(row.id)">預覽</el-button>
                <el-button size="small" @click.stop="openOverrideForProduct(row.id)">覆寫</el-button>
                <el-button size="small" type="primary" :icon="Edit" @click.stop="openEditProduct(row.id)">編輯</el-button>
                  <el-button size="small" type="danger" :icon="Delete" @click.stop="removeProductById(row.id)">刪除</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-card>
    </div>

    <el-dialog
      v-model="isPreviewModalOpen"
      title="預覽商品內容"
      width="960px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <template #header>
        <div class="element-dialog-heading">
          <div>
            <h3>預覽商品內容</h3>
            <small>{{ selectedProduct?.baseName }}</small>
          </div>
        </div>
      </template>

      <div v-if="selectedProduct" class="element-dialog-body">
          <div class="preview-grid">
            <div class="main-preview">
              <el-image
                :src="displayImage(selectedProduct.previewImage)"
                :preview-src-list="selectedProduct.gallery?.length ? selectedProduct.gallery.map(displayImage) : [displayImage(selectedProduct.previewImage)]"
                :initial-index="0"
                fit="cover"
              />
            </div>
            <div class="gallery-preview" v-if="selectedProduct.gallery?.length">
              <el-image
                v-for="(image, index) in selectedProduct.gallery"
                :key="`preview-gallery-${image}-${index}`"
                :src="displayImage(image)"
                :preview-src-list="selectedProduct.gallery.map(displayImage)"
                :initial-index="index"
                fit="cover"
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
                <dd>{{ formatProductCategorySummary(selectedProduct) }}</dd>
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
              <h4>產品說明</h4>
              <p class="overview-description">{{ selectedProduct.longDescription }}</p>
            </div>

            <div class="preview-copy-block" v-if="selectedProduct.specificationHtml">
              <h4>產品規格</h4>
              <p class="overview-description">{{ selectedProduct.specificationHtml }}</p>
            </div>

          </div>
      </div>
      <template #footer>
        <el-button @click="closePreviewModal">關閉</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="isOverrideModalOpen"
      title="租戶商品覆寫"
      width="1220px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <template #header>
        <div class="element-dialog-heading">
          <div>
            <h3>租戶商品覆寫</h3>
            <small>同一商品下，直接查看並編輯所有租戶的覆寫內容。可勾選租戶後批量生成名稱。</small>
          </div>
        </div>
      </template>

      <div v-if="selectedProduct" class="element-dialog-body">
        <div class="override-toolbar">
          <section class="override-toolbar-panel override-toolbar-summary">
            <p class="override-toolbar-eyebrow">Batch Scope</p>
            <div class="override-toolbar-title-row">
              <strong>批量命名工具</strong>
              <span class="override-selection-badge">已選 {{ selectedOverrideTenantIds.length }}</span>
            </div>
            <el-checkbox
              :model-value="isAllOverrideTenantsSelected"
              @change="(value: CheckboxValueType) => toggleSelectAllOverrideTenants(Boolean(value))"
            >
              全選租戶
            </el-checkbox>
            <p class="override-toolbar-hint">
              勾選要處理的租戶後，再批量生成自訂名稱。生成後仍可逐行微調再儲存。
            </p>
          </section>

          <section class="override-toolbar-panel override-toolbar-generator">
            <p class="override-toolbar-eyebrow">AI Prompt</p>
            <strong class="override-toolbar-panel-title">補充要求</strong>
            <el-input
              v-model="bulkNameGenerationInstruction"
              class="override-bulk-input"
              type="textarea"
              :rows="3"
              resize="none"
              placeholder="例如：偏高端、簡潔、強調冰葡萄口味"
            />
            <div class="override-toolbar-actions">
              <el-button
                type="primary"
                size="default"
                :disabled="isGeneratingBulkCustomNames"
                @click="bulkGenerateOverrideNamesForCurrentProduct"
              >
                {{ isGeneratingBulkCustomNames ? '生成中...' : '批量生成自訂商品名稱' }}
              </el-button>
            </div>
          </section>

          <div v-if="bulkCustomNameResultMessage" class="override-toolbar-result">
            <div class="override-toolbar-result-head">
              <strong>{{ bulkCustomNameResultMessage }}</strong>
              <span>本次更新租戶</span>
            </div>
            <div v-if="bulkCustomNameResultTenants.length" class="override-result-tags">
              <span
                v-for="tenantName in bulkCustomNameResultTenants"
                :key="tenantName"
                class="override-result-tag"
              >
                {{ tenantName }}
              </span>
            </div>
          </div>
        </div>

        <el-table :data="store.tenants" stripe style="width: 100%">
          <el-table-column width="56">
            <template #header>
              <input
                :checked="isAllOverrideTenantsSelected"
                type="checkbox"
                @change="toggleSelectAllOverrideTenants(($event.target as HTMLInputElement).checked)"
              />
            </template>
            <template #default="{ row }">
              <input
                :checked="selectedOverrideTenantIds.includes(row.id)"
                type="checkbox"
                @change="toggleOverrideTenantSelection(row.id, ($event.target as HTMLInputElement).checked)"
              />
            </template>
          </el-table-column>
          <el-table-column label="租戶" min-width="120">
            <template #default="{ row }">
              <strong>{{ row.name }}</strong>
            </template>
          </el-table-column>
          <el-table-column label="域名" min-width="140">
            <template #default="{ row }">
              <span class="override-domain-cell">{{ row.domain || '未設定主域名' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="原商品名稱" min-width="160">
            <template #default>
              <strong class="override-original-name-cell">{{ selectedProduct.baseName }}</strong>
            </template>
          </el-table-column>
          <el-table-column label="顯示" width="100">
            <template #default="{ row }">
              <el-switch
                v-if="getOverrideForm(row.id)"
                v-model="getOverrideForm(row.id)!.isVisible"
                inline-prompt
                active-text="顯示"
                inactive-text="隱藏"
              />
            </template>
          </el-table-column>
          <el-table-column label="自訂商品名稱" min-width="260">
            <template #default="{ row }">
              <el-input
                v-if="getOverrideForm(row.id)"
                v-model="getOverrideForm(row.id)!.customName"
                type="textarea"
                :autosize="{ minRows: 2, maxRows: 4 }"
                resize="none"
              />
            </template>
          </el-table-column>
          <el-table-column label="租戶售價" width="120">
            <template #default="{ row }">
              <el-input-number v-if="getOverrideForm(row.id)" v-model="getOverrideForm(row.id)!.customPrice" size="small" :min="0" :step="0.01" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="租戶庫存" width="120">
            <template #default="{ row }">
              <el-input-number v-if="getOverrideForm(row.id)" v-model="getOverrideForm(row.id)!.customStockQuantity" size="small" :min="0" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="SEO 標題" min-width="260">
            <template #default="{ row }">
              <el-input
                v-if="getOverrideForm(row.id)"
                v-model="getOverrideForm(row.id)!.seoTitle"
                type="textarea"
                :autosize="{ minRows: 2, maxRows: 4 }"
                resize="none"
              />
            </template>
          </el-table-column>
          <el-table-column label="SEO 描述" min-width="320">
            <template #default="{ row }">
              <el-input
                v-if="getOverrideForm(row.id)"
                v-model="getOverrideForm(row.id)!.seoDescription"
                type="textarea"
                :autosize="{ minRows: 3, maxRows: 6 }"
                resize="none"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" @click="saveOverride(row.id)">儲存</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <template #footer>
        <el-button @click="closeOverrideModal">關閉</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="isProductModalOpen"
      :title="isCreating ? '新增商品' : '修改商品'"
      width="1120px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <template #header>
        <div class="element-dialog-heading">
          <div>
            <h3>{{ isCreating ? '新增商品' : '修改商品' }}</h3>
            <small v-if="!isCreating">最後更新 {{ productForm.updatedAt || '尚未更新' }}</small>
          </div>
        </div>
      </template>

      <div class="element-dialog-body">
          <div class="edit-gallery-panel">
            <div class="edit-gallery-head">
              <strong>圖片預覽</strong>
              <span>點擊圖片可放大，點按「設為主圖」可調整商品主圖。</span>
            </div>
          <div class="preview-grid">
            <div class="gallery-preview">
              <div
                v-for="(image, index) in productForm.gallery"
                :key="`gallery-${image}-${index}`"
                class="image-tile"
                :class="{ active: image === productForm.previewImage }"
              >
                <el-image
                  class="gallery-lightbox-image"
                  :src="displayImage(image)"
                  :preview-src-list="productForm.gallery.map(displayImage)"
                  :initial-index="index"
                  fit="cover"
                  />
                <div class="image-actions">
                  <el-button
                    size="small"
                    :disabled="image === productForm.previewImage"
                    @click="setPreviewImage(image)"
                  >
                    {{ image === productForm.previewImage ? '目前主圖' : '設為主圖' }}
                  </el-button>
                  <el-button size="small" type="danger" plain @click="removeGalleryImage(index)">移除</el-button>
                </div>
              </div>
            </div>
          </div>
          </div>
          <div class="element-form-grid">
            <el-form label-width="100px">
              <div class="element-form-grid two-columns">
                <el-form-item label="SKU">
                  <el-input v-model="productForm.sku" />
                </el-form-item>
                <el-form-item label="Slug">
                  <el-input :model-value="productForm.slug || '儲存後自動產生'" readonly />
                </el-form-item>
                <el-form-item label="主分類">
                  <el-select v-model="productForm.categoryId" style="width: 100%" @change="syncCategorySelection">
                    <el-option :value="null" label="請選擇分類" />
                    <el-option v-for="category in categoryTreeOptions" :key="category.id" :value="category.id" :label="category.label" />
                  </el-select>
                </el-form-item>
            <div class="full category-panel">
              <div class="category-panel-head">
                <span>所属分类</span>
                <small class="field-hint">主分类也必须属于这里。勾选多个分类后，前台会同时归档到这些分类下。</small>
              </div>
              <div class="category-panel-grid">
                <div class="category-tree-picker">
                  <el-tree
                    ref="categoryTreeRef"
                    :data="categoryTreeData"
                    node-key="id"
                    show-checkbox
                    check-on-click-node
                    default-expand-all
                    :expand-on-click-node="false"
                    @check="handleCategoryTreeCheck"
                  >
                    <template #default="{ data }">
                      <div class="category-tree-node">
                        <span>{{ data.label }}</span>
                        <el-tag v-if="productForm.categoryId === data.id" size="small" effect="light" type="primary">主分类</el-tag>
                      </div>
                    </template>
                  </el-tree>
                </div>
                <div class="category-selection-summary">
                  <strong>已选分类</strong>
                  <div v-if="selectedCategoryTagList.length" class="category-tag-list">
                    <el-button
                      v-for="item in selectedCategoryTagList"
                      :key="`selected-category-${item.id}`"
                      class="category-tag"
                      size="small"
                      :type="item.isPrimary ? 'primary' : 'info'"
                      plain
                      @click="removeProductCategory(item.id)"
                    >
                      {{ item.name }}
                      {{ item.isPrimary ? ' · 主分类' : ' · 移除' }}
                    </el-button>
                  </div>
                  <p v-else class="empty-selection-text">暂未选择任何分类。</p>
                </div>
              </div>
            </div>
                <el-form-item label="品牌">
                  <el-select v-model="productForm.brandId" style="width: 100%" @change="syncBrandSelection">
                    <el-option :value="null" label="請選擇品牌" />
                    <el-option v-for="brand in brandOptions" :key="brand.id" :value="brand.id" :label="brand.name" />
                  </el-select>
                </el-form-item>
                <el-form-item class="full" label="基礎商品名稱">
                  <el-input v-model="productForm.baseName" />
                </el-form-item>
                <el-form-item class="full" label="商品簡介">
                  <el-input v-model="productForm.description" type="textarea" :rows="3" placeholder="顯示在商品卡片與詳情頁標題下方的摘要，支援 HTML" />
                </el-form-item>
                <el-form-item class="full" label="產品說明">
                  <el-input v-model="productForm.longDescription" type="textarea" :rows="6" placeholder="顯示在前台商品詳情區塊，支援 HTML" />
                </el-form-item>
                <el-form-item class="full" label="產品規格">
                  <el-input v-model="productForm.specificationHtml" type="textarea" :rows="6" placeholder="顯示在前台商品詳情頁常見問題上方，支援 HTML" />
                </el-form-item>
                <el-form-item label="基礎價格">
                  <el-input-number v-model="productForm.basePrice" :min="0" :step="0.01" style="width: 100%" />
                </el-form-item>
                <el-form-item label="基礎庫存">
                  <el-input-number v-model="productForm.baseStockQuantity" :min="0" style="width: 100%" />
                </el-form-item>
                <el-form-item class="full" label="主圖 URL">
                  <el-input v-model="productForm.previewImage" />
                </el-form-item>
                <el-form-item class="full" label="圖庫 URL（每行一張）">
                  <el-input v-model="galleryInput" type="textarea" :rows="4" @change="syncGalleryFromInput" />
                  <input type="file" multiple accept="image/*" @change="uploadGalleryImages" />
                </el-form-item>
                <el-form-item class="full" label="組合 SKU">
                  <el-input
                    v-model="skuVariantInput"
                    type="textarea"
                    :rows="5"
                    placeholder="例如&#10;口味:冰葡萄; 盒裝:2入|POD-GRAPE-2|299|12&#10;口味:冰葡萄; 盒裝:4入|POD-GRAPE-4|499|8&#10;&#10;規格群組會依這裡的組合自動推導"
                    @change="syncVariantConfigFromInput"
                  />
              <div v-if="productForm.skuVariants?.length" class="sku-variant-preview">
                <div class="sku-variant-preview-head">
                  <strong>目前組合 SKU 配置</strong>
                  <span>{{ productForm.skuVariants.length }} 組</span>
                </div>
                <div class="sku-variant-table">
                  <div class="sku-variant-row sku-variant-row-head">
                    <span>規格組合</span>
                    <span>SKU</span>
                    <span>價格</span>
                    <span>庫存</span>
                  </div>
                  <div
                    v-for="(variant, index) in productForm.skuVariants"
                    :key="`${variant.sku}-${index}`"
                    class="sku-variant-row"
                  >
                    <span>{{ Object.entries(variant.selections).map(([key, value]) => `${key}: ${value}`).join(' / ') }}</span>
                    <span>{{ variant.sku }}</span>
                    <span>{{ variant.price ?? '—' }}</span>
                    <span>{{ variant.stock ?? '—' }}</span>
                  </div>
                </div>
              </div>
                </el-form-item>
                <el-form-item class="full" label="狀態">
                  <el-select v-model="productForm.status" style="width: 100%">
                    <el-option value="上架中" label="上架中" />
                    <el-option value="草稿" label="草稿" />
                    <el-option value="缺貨" label="缺貨" />
                  </el-select>
                </el-form-item>
              </div>
            </el-form>
          </div>

      </div>

      <template #footer>
        <div class="element-dialog-footer">
          <el-button v-if="!isCreating" type="danger" @click="removeProduct">刪除商品</el-button>
          <div class="action-group">
            <el-button @click="closeProductModal">取消</el-button>
            <el-button type="primary" @click="saveProduct">{{ isCreating ? '建立商品' : '儲存全域商品' }}</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.products-page {
  display: grid;
  gap: 1rem;
  min-width: 0;
  max-width: 100%;
  overflow-x: hidden;
}

.element-page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.2rem 0 0.1rem;
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
  margin-bottom: 0.1rem;
}

.metrics-grid :deep(.el-card) {
  border-radius: 16px;
  border-color: #e5e7eb;
  background: linear-gradient(180deg, #ffffff 0%, #fbfcfe 100%);
}

.metric-label {
  color: var(--wp-text-muted);
  margin: 0;
  font-size: 0.82rem;
}

.metric-value {
  display: block;
  margin-top: 0.3rem;
  font-size: 1.45rem;
  line-height: 1.1;
}

.workspace-grid {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

.table-card {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  border-radius: 18px;
  border-color: #e5e7eb;
  overflow: hidden;
}

.workspace-grid > * {
  min-width: 0;
  max-width: 100%;
}

.table-card :deep(.el-card__header) {
  padding: 1rem 1.1rem 0.85rem;
  border-bottom: 1px solid #edf0f3;
  background: linear-gradient(180deg, #fcfdff 0%, #f8fafc 100%);
}

.table-card :deep(.el-card__body) {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  padding: 1rem 1.1rem 1.1rem;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  min-width: 0;
}

.card-header .title {
  display: block;
  font-size: 1rem;
  font-weight: 700;
  color: #111827;
}

.card-header small {
  display: block;
  margin-top: 0.22rem;
  color: var(--wp-text-muted);
  font-size: 0.78rem;
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.6rem;
  min-width: 0;
  max-width: 100%;
}

.toolbar-field {
  width: 160px;
  min-width: 0;
}

.toolbar-field-wide {
  width: 210px;
}

.toolbar-search-field {
  width: 220px;
  min-width: 0;
}

.table-toolbar {
  display: grid;
  gap: 0.9rem;
  min-width: 0;
  max-width: 100%;
}

.bulk-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.7rem;
  align-items: center;
  min-width: 0;
  max-width: 100%;
}

.bulk-meta-group {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.2rem;
  padding: 0 0.85rem;
  border: 1px solid #dbe2ea;
  border-radius: 999px;
  background: #f8fbff;
  color: #4b5563;
  font-size: 0.82rem;
}

.bulk-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.75rem;
  min-height: 1.75rem;
  padding: 0 0.4rem;
  border-radius: 999px;
  background: #2271b1;
  color: #fff;
  font-size: 0.78rem;
  line-height: 1;
}

.filter-summary {
  display: flex;
  align-items: center;
  min-height: 2.2rem;
  padding: 0 0.85rem;
  border: 1px dashed #d6dde5;
  border-radius: 999px;
  background: #fafcff;
  flex: 1 1 260px;
  color: var(--wp-text-muted);
  font-size: 0.8rem;
}

.bulk-status-field {
  width: 140px;
  min-width: 0;
}

.product-table-shell {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
}

.product-admin-table {
  width: 100%;
  min-width: 1180px;
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

.category-cell {
  display: grid;
  gap: 0.3rem;
  min-width: 0;
}

.category-cell strong {
  font-size: 0.88rem;
  line-height: 1.25;
  overflow-wrap: anywhere;
}

.category-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.28rem;
}

.category-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.35rem;
  padding: 0.1rem 0.42rem;
  border-radius: 999px;
  background: #f3f5f6;
  color: var(--wp-text-muted);
  font-size: 0.72rem;
  line-height: 1.1;
}

.slug-cell {
  max-width: 220px;
  color: var(--wp-text-muted);
  font-size: 0.82rem;
}

.slug-text {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-actions {
  display: flex;
  gap: 0.45rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.table-card :deep(.el-table) {
  --el-table-border-color: #edf0f3;
  --el-table-header-bg-color: #f8fafc;
  --el-table-row-hover-bg-color: #f8fbff;
  border-radius: 12px;
  overflow: hidden;
}

.table-card :deep(.el-table th.el-table__cell) {
  background: #f8fafc;
  color: #6b7280;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.table-card :deep(.el-table td.el-table__cell) {
  padding-top: 14px;
  padding-bottom: 14px;
}

.table-card :deep(.el-table .cell) {
  min-width: 0;
}

.table-card :deep(.el-input__wrapper),
.table-card :deep(.el-select__wrapper),
.table-card :deep(.el-button) {
  border-radius: 10px;
}

.table-card :deep(.el-table .cell) {
  min-width: 0;
}

.element-dialog-heading h3 {
  margin: 0;
}

.element-dialog-heading small {
  display: block;
  margin-top: 0.2rem;
  color: var(--wp-text-muted);
  font-size: 0.78rem;
}

.element-dialog-body {
  display: grid;
  gap: 1rem;
}

.element-form-grid {
  display: grid;
  gap: 1rem;
}

.element-form-grid.two-columns {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.element-dialog-footer {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  width: 100%;
}

.preview-copy-grid {
  display: grid;
  gap: 1rem;
}

.preview-copy-block {
  display: grid;
  gap: 0.45rem;
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

.edit-gallery-panel {
  display: grid;
  gap: 0.75rem;
  padding: 0.85rem 0.95rem;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  background: linear-gradient(180deg, #fcfdff 0%, #f8fafc 100%);
}

.edit-gallery-head {
  display: grid;
  gap: 0.22rem;
}

.edit-gallery-head strong {
  font-size: 0.92rem;
  color: #111827;
}

.edit-gallery-head span {
  color: var(--wp-text-muted);
  font-size: 0.78rem;
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

.main-preview :deep(.el-image) {
  display: block;
  width: 100%;
  height: 180px;
}

.gallery-preview {
  display: grid;
  gap: 0.5rem;
  grid-template-columns: repeat(auto-fit, minmax(64px, 1fr));
}

.gallery-preview :deep(.el-image),
.detail-preview img {
  width: 100%;
  height: 100%;
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

.image-tile.active img {
  border-color: var(--wp-blue);
  box-shadow: 0 0 0 2px rgba(34, 113, 177, 0.14);
}

.gallery-lightbox-image :deep(.el-image__inner),
.gallery-preview :deep(.el-image__inner) {
  display: block;
  width: 100%;
  min-height: 64px;
  object-fit: cover;
}

.image-actions {
  display: flex;
  gap: 0.4rem;
}

.field-hint {
  color: var(--wp-text-muted);
  font-size: 0.78rem;
  line-height: 1.4;
}

.category-panel {
  display: grid;
  gap: 0.55rem;
  padding: 0.75rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.6rem;
  background: var(--wp-surface-soft);
}

.category-panel-head {
  display: grid;
  gap: 0.2rem;
}

.category-panel-head span {
  font-weight: 600;
  font-size: 0.84rem;
}

.category-panel-grid {
  display: grid;
  gap: 0.6rem;
  grid-template-columns: minmax(0, 1.8fr) minmax(220px, 0.9fr);
}

.category-tree-picker,
.category-selection-summary {
  display: grid;
  gap: 0.4rem;
  padding: 0.6rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.55rem;
  background: #fff;
}

.category-tree-picker {
  height: 232px;
  overflow-y: auto;
}

.category-tree-picker :deep(.el-tree) {
  background: transparent;
}

.category-tree-picker :deep(.el-tree-node__content) {
  min-height: 34px;
  border-radius: 8px;
}

.category-tree-picker :deep(.el-tree-node__content:hover) {
  background: #f8fbff;
}

.category-tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  width: 100%;
  min-width: 0;
}

.category-selection-summary {
  align-content: start;
  min-height: 232px;
}

.category-check-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 1.8rem;
  padding: 0.22rem 0.3rem;
  border-radius: 0.45rem;
}

.category-check-row:hover {
  background: #f8fbff;
}

.category-check-row.primary {
  background: #eef6ff;
}

.category-check-row input {
  width: auto;
  min-height: auto;
  margin: 0;
}

.picker-guides {
  flex: 0 0 auto;
}

.category-check-name {
  min-width: 0;
  font-size: 0.84rem;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.primary-badge {
  margin-left: auto;
  padding: 0.08rem 0.35rem;
  border-radius: 999px;
  background: rgba(34, 113, 177, 0.12);
  color: var(--wp-blue);
  font-size: 0.7rem;
  font-weight: 700;
}

.category-selection-summary strong {
  font-size: 0.84rem;
}

.category-tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-content: start;
}

.category-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  min-height: 1.6rem;
  padding: 0.18rem 0.45rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.8rem;
  background: #fff;
  color: var(--wp-text);
  font-size: 0.76rem;
  line-height: 1.2;
}

.category-tag.primary {
  border-color: rgba(34, 113, 177, 0.3);
  background: #eef6ff;
  color: var(--wp-blue);
}

.category-tag span {
  color: var(--wp-text-muted);
  font-size: 0.68rem;
}

.empty-selection-text {
  color: var(--wp-text-muted);
  font-size: 0.78rem;
}

.sku-variant-preview {
  display: grid;
  gap: 0.625rem;
  margin-top: 0.5rem;
  padding: 0.85rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
}

.sku-variant-preview-head {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
}

.sku-variant-preview-head strong {
  font-size: 0.9rem;
}

.sku-variant-preview-head span {
  color: var(--wp-text-muted);
  font-size: 0.8rem;
}

.sku-variant-table {
  display: grid;
  gap: 0.35rem;
}

.sku-variant-row {
  display: grid;
  grid-template-columns: minmax(0, 2.2fr) minmax(120px, 1.2fr) 88px 88px;
  gap: 0.6rem;
  align-items: start;
  padding: 0.5rem 0.6rem;
  border-radius: 0.4rem;
  background: #fff;
  font-size: 0.84rem;
}

.sku-variant-row span {
  min-width: 0;
  overflow-wrap: anywhere;
}

.sku-variant-row-head {
  background: transparent;
  padding: 0;
  color: var(--wp-text-muted);
  font-size: 0.76rem;
  font-weight: 600;
}

.action-group {
  display: flex;
  gap: 0.625rem;
}

.override-selection-meta {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  color: var(--wp-text-muted);
  font-size: 0.78rem;
}

.override-selection-meta strong {
  color: var(--wp-text);
  font-size: 0.8rem;
}

.override-select-all {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.override-select-all input,
.override-bulk-input {
  flex: 1 1 260px;
  min-width: 220px;
  min-height: 2.2rem;
}

.override-table-wrap {
  overflow-x: auto;
}

.override-table {
  width: 100%;
  min-width: 1080px;
  border-collapse: separate;
  border-spacing: 0;
}

.override-table th,
.override-table td {
  padding: 0.55rem 0.6rem;
  border-bottom: 1px solid var(--wp-border);
  vertical-align: top;
  background: #fff;
}

.override-checkbox-col {
  width: 40px;
  text-align: center;
}

.override-table th {
  position: sticky;
  top: 0;
  z-index: 1;
  text-align: left;
  color: var(--wp-text-muted);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  background: #f8fafc;
  white-space: nowrap;
}

.override-table tr:last-child td {
  border-bottom: none;
}

.override-tenant-cell strong {
  display: block;
  min-width: 104px;
  color: var(--wp-text);
  font-size: 0.84rem;
  line-height: 1.25;
}

.override-domain-cell {
  min-width: 130px;
  max-width: 150px;
  color: var(--wp-text-muted);
  font-size: 0.72rem;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.override-original-name-cell {
  min-width: 150px;
  max-width: 190px;
}

.override-original-name-cell strong {
  display: block;
  color: var(--wp-text-muted);
  font-size: 0.78rem;
  line-height: 1.4;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.override-toolbar {
  display: grid;
  grid-template-columns: minmax(250px, 0.95fr) minmax(360px, 1.25fr);
  gap: 1rem;
  padding: 1rem;
  border: 1px solid #dbe5ef;
  border-radius: 18px;
  background:
    radial-gradient(circle at top right, rgba(34, 113, 177, 0.08), transparent 28%),
    linear-gradient(180deg, #fcfdff 0%, #f7fafc 100%);
}

.override-toolbar-copy {
  display: grid;
  gap: 0.6rem;
  align-content: start;
  padding: 0.9rem 0.95rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.82);
}

.override-toolbar-hint {
  margin: 0;
  color: var(--wp-text-muted);
  font-size: 0.78rem;
  line-height: 1.55;
}

.override-toolbar-actions {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.65rem;
  align-items: center;
  padding: 0.9rem 0.95rem;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.82);
}

.override-toolbar-result {
  grid-column: 1 / -1;
  display: grid;
  gap: 0.55rem;
  padding: 0.8rem 0.95rem;
  border: 1px solid rgba(10, 92, 54, 0.16);
  border-radius: 16px;
  background: linear-gradient(180deg, #f3fbf5 0%, #eef9f2 100%);
  color: #0a5c36;
  font-size: 0.78rem;
  line-height: 1.45;
}

.override-toolbar-result-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.override-toolbar-result-head strong {
  font-size: 0.84rem;
  font-weight: 700;
}

.override-toolbar-result-head span {
  color: rgba(10, 92, 54, 0.72);
  font-size: 0.72rem;
  font-weight: 600;
}

.override-result-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.override-result-tag {
  display: inline-flex;
  align-items: center;
  min-height: 1.65rem;
  padding: 0.1rem 0.55rem;
  border-radius: 999px;
  background: rgba(10, 92, 54, 0.1);
  color: #0a5c36;
  font-size: 0.74rem;
  font-weight: 600;
}

.element-dialog-body :deep(.el-table) {
  --el-table-border-color: #edf0f3;
  --el-table-header-bg-color: #f8fafc;
  --el-table-row-hover-bg-color: #f8fbff;
  border-radius: 12px;
  overflow: hidden;
}

.element-dialog-body :deep(.el-table th.el-table__cell) {
  background: #f8fafc;
  color: #6b7280;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.element-dialog-body :deep(.el-table td.el-table__cell) {
  padding-top: 12px;
  padding-bottom: 12px;
}

.element-dialog-body :deep(.el-dialog__body),
.element-dialog-body :deep(.el-input__wrapper),
.element-dialog-body :deep(.el-textarea__inner),
.element-dialog-body :deep(.el-select__wrapper),
.element-dialog-body :deep(.el-input-number),
.element-dialog-body :deep(.el-button) {
  border-radius: 10px;
}

:deep(.el-dialog) {
  border-radius: 18px;
  overflow: hidden;
}

:deep(.el-dialog__header) {
  margin-right: 0;
  padding: 1rem 1.1rem 0.9rem;
  border-bottom: 1px solid #edf0f3;
  background: linear-gradient(180deg, #fcfdff 0%, #f8fafc 100%);
}

:deep(.el-dialog__body) {
  padding: 1rem 1.1rem 1rem;
}

:deep(.el-dialog__footer) {
  padding: 0.9rem 1.1rem 1rem;
  border-top: 1px solid #edf0f3;
  background: #fff;
}

@media (max-width: 1100px) {
  .overview-details {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 780px) {
  .element-page-heading,
  .card-header,
  .toolbar-actions,
  .element-dialog-footer {
    flex-direction: column;
  }

  .toolbar-field,
  .toolbar-field-wide,
  .toolbar-search-field,
  .bulk-status-field {
    width: 100%;
  }

  .bulk-toolbar,
  .filter-summary {
    width: 100%;
  }

  .metrics-grid,
  .element-form-grid.two-columns {
    grid-template-columns: 1fr;
  }

  .overview-details {
    grid-template-columns: 1fr;
  }

  .action-group {
    width: 100%;
    flex-direction: column;
  }

  .category-panel-grid {
    grid-template-columns: 1fr;
  }

  .override-toolbar,
  .override-toolbar-actions {
    grid-template-columns: 1fr;
  }

  .override-toolbar-result-head {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
