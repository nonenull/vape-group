import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { adminAPI } from '@/api/admin'
import {
  type BrandRecord,
  type CategoryRecord,
  orderRecords,
  type OrderRecord,
  type ProductOverrideRecord,
  type ProductRecord,
  type TenantRecord,
} from '@/data/adminMock'

export const useAdminStore = defineStore('admin', () => {
  const tenants = ref<TenantRecord[]>([])
  const products = ref<ProductRecord[]>([])
  const categories = ref<CategoryRecord[]>([])
  const brands = ref<BrandRecord[]>([])
  const overrides = ref<ProductOverrideRecord[]>([])
  const orders = ref<OrderRecord[]>(structuredClone(orderRecords))
  const loading = ref(false)

  const activeTenants = computed(() => tenants.value.filter((tenant) => tenant.isActive))
  const visibleOverrides = computed(() => overrides.value.filter((item) => item.isVisible))

  async function fetchTenants() {
    loading.value = true
    try {
      tenants.value = await adminAPI.getTenants()
    } finally {
      loading.value = false
    }
  }

  async function fetchProducts() {
    loading.value = true
    try {
      products.value = await adminAPI.getProducts()
    } finally {
      loading.value = false
    }
  }

  async function fetchProductOverride(productId: number, tenantId: number) {
    const override = await adminAPI.getProductOverride(productId, tenantId)

    const index = overrides.value.findIndex(
      (item) => item.productId === productId && item.tenantId === tenantId,
    )
    if (index >= 0) {
      overrides.value[index] = override
    } else {
      overrides.value.push(override)
    }
    return override
  }

  async function fetchCategories(tenantId: number) {
    const result = await adminAPI.getCategories(tenantId)
    categories.value = [
      ...categories.value.filter((item) => item.tenantId !== tenantId),
      ...result,
    ]
    return result
  }

  async function fetchBrands(tenantId: number) {
    const result = await adminAPI.getBrands(tenantId)
    brands.value = [
      ...brands.value.filter((item) => item.tenantId !== tenantId),
      ...result,
    ]
    return result
  }

  async function createTenant(payload: Omit<TenantRecord, 'id'>) {
    const created = await adminAPI.createTenant(payload)
    tenants.value.unshift(created)
    return created
  }

  async function updateTenant(payload: TenantRecord) {
    const updated = await adminAPI.updateTenant(payload.id, {
      domain: payload.domain,
      boundDomains: payload.boundDomains,
      name: payload.name,
      isActive: payload.isActive,
      theme: payload.theme,
      previewImage: payload.previewImage,
      logoImage: payload.logoImage,
      accentColor: payload.accentColor,
      surfaceColor: payload.surfaceColor,
      heroTitle: payload.heroTitle,
      tagline: payload.tagline,
      announcement: payload.announcement,
      supportText: payload.supportText,
      seoTitle: payload.seoTitle,
      seoDescription: payload.seoDescription,
    })
    const index = tenants.value.findIndex((item) => item.id === payload.id)
    if (index >= 0) {
      tenants.value[index] = updated
    }
  }

  async function deleteTenant(tenantId: number) {
    await adminAPI.deleteTenant(tenantId)
    tenants.value = tenants.value.filter((item) => item.id !== tenantId)
    overrides.value = overrides.value.filter((item) => item.tenantId !== tenantId)
  }

  async function createProduct(payload: Omit<ProductRecord, 'id' | 'updatedAt'>) {
    const created = await adminAPI.createProduct(payload)
    products.value.unshift(created)
    return created
  }

  async function updateProduct(payload: ProductRecord) {
    const updated = await adminAPI.updateProduct(payload.id, {
      sku: payload.sku,
      baseName: payload.baseName,
      basePrice: payload.basePrice,
      baseStockQuantity: payload.baseStockQuantity,
      category: payload.category,
      categoryId: payload.categoryId,
      brand: payload.brand,
      brandId: payload.brandId,
      previewImage: payload.previewImage,
      gallery: payload.gallery,
      detailImages: payload.detailImages,
      status: payload.status,
      description: payload.description,
      longDescription: payload.longDescription,
      badge: payload.badge,
      rating: payload.rating,
      reviews: payload.reviews,
      flavors: payload.flavors,
      variants: payload.variants,
      optionGroups: payload.optionGroups,
      skuVariants: payload.skuVariants,
      isActive: payload.isActive,
    })
    const index = products.value.findIndex((item) => item.id === payload.id)
    if (index >= 0) {
      products.value[index] = updated
    }
  }

  async function createCategory(tenantId: number, payload: Omit<CategoryRecord, 'id' | 'tenantId'>) {
    const created = await adminAPI.createCategory(tenantId, payload)
    categories.value.unshift(created)
    return created
  }

  async function updateCategory(payload: CategoryRecord) {
    const updated = await adminAPI.updateCategory(payload.tenantId, payload.id, {
      name: payload.name,
      parentId: payload.parentId,
      sortOrder: payload.sortOrder,
    })
    const index = categories.value.findIndex((item) => item.id === payload.id)
    if (index >= 0) {
      categories.value[index] = updated
    }
    return updated
  }

  async function deleteCategory(tenantId: number, categoryId: number) {
    await adminAPI.deleteCategory(tenantId, categoryId)
    categories.value = categories.value.filter((item) => item.id !== categoryId)
  }

  async function createBrand(tenantId: number, payload: Omit<BrandRecord, 'id' | 'tenantId'>) {
    const created = await adminAPI.createBrand(tenantId, payload)
    brands.value.unshift(created)
    return created
  }

  async function updateBrand(payload: BrandRecord) {
    const updated = await adminAPI.updateBrand(payload.tenantId, payload.id, {
      name: payload.name,
      logoUrl: payload.logoUrl,
      description: payload.description,
    })
    const index = brands.value.findIndex((item) => item.id === payload.id)
    if (index >= 0) {
      brands.value[index] = updated
    }
    return updated
  }

  async function deleteBrand(tenantId: number, brandId: number) {
    await adminAPI.deleteBrand(tenantId, brandId)
    brands.value = brands.value.filter((item) => item.id !== brandId)
  }

  async function deleteProduct(productId: number) {
    await adminAPI.deleteProduct(productId)
    products.value = products.value.filter((item) => item.id !== productId)
    overrides.value = overrides.value.filter((item) => item.productId !== productId)
  }

  async function bulkUpdateProducts(productIds: number[], payload: { status?: ProductRecord['status'], isActive?: boolean }) {
    const updatedProducts = await adminAPI.bulkUpdateProducts(productIds, payload)
    const updatedMap = new Map(updatedProducts.map((item) => [item.id, item]))
    products.value = products.value.map((item) => updatedMap.get(item.id) ?? item)
    return updatedProducts
  }

  async function upsertOverride(payload: ProductOverrideRecord) {
    const updated = await adminAPI.updateProductOverride(payload.productId, payload.tenantId, {
      customName: payload.customName,
      customDescription: payload.customDescription,
      customPrice: payload.customPrice,
      customStockQuantity: payload.customStockQuantity,
      customImages: payload.customImages ?? [],
      customDetailImages: payload.customDetailImages ?? [],
      seoTitle: payload.seoTitle,
      seoDescription: payload.seoDescription,
      isVisible: payload.isVisible,
    })
    const merged: ProductOverrideRecord = {
      id: updated.id,
      tenantId: updated.tenantId,
      productId: updated.productId,
      customName: updated.customName,
      customDescription: updated.customDescription,
      customPrice: updated.customPrice,
      customStockQuantity: updated.customStockQuantity,
      customImages: updated.customImages ?? [],
      customDetailImages: updated.customDetailImages ?? [],
      seoTitle: updated.seoTitle,
      seoDescription: updated.seoDescription,
      isVisible: updated.isVisible,
    }
    const index = overrides.value.findIndex(
      (item) => item.productId === merged.productId && item.tenantId === merged.tenantId,
    )
    if (index >= 0) {
      overrides.value[index] = merged
    } else {
      overrides.value.push(merged)
    }
  }

  function updateOrderStatus(orderId: number, status: OrderRecord['status']) {
    const order = orders.value.find((item) => item.id === orderId)
    if (order) {
      order.status = status
    }
  }

  function getTenantName(tenantId: number) {
    return tenants.value.find((item) => item.id === tenantId)?.name ?? '未指定租戶'
  }

  function getOverride(productId: number, tenantId: number) {
    return overrides.value.find((item) => item.productId === productId && item.tenantId === tenantId)
  }

  function getCategoriesByTenant(tenantId: number) {
    return categories.value
      .filter((item) => item.tenantId === tenantId)
      .sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id)
  }

  function getBrandsByTenant(tenantId: number) {
    return brands.value.filter((item) => item.tenantId === tenantId)
  }

  async function bootstrap() {
    await Promise.all([fetchTenants(), fetchProducts()])
  }

  return {
    tenants,
    products,
    categories,
    brands,
    overrides,
    orders,
    loading,
    activeTenants,
    visibleOverrides,
    fetchTenants,
    fetchProducts,
    fetchProductOverride,
    fetchCategories,
    fetchBrands,
    createTenant,
    updateTenant,
    deleteTenant,
    createProduct,
    updateProduct,
    deleteProduct,
    bulkUpdateProducts,
    createCategory,
    updateCategory,
    deleteCategory,
    createBrand,
    updateBrand,
    deleteBrand,
    upsertOverride,
    updateOrderStatus,
    getTenantName,
    getOverride,
    getCategoriesByTenant,
    getBrandsByTenant,
    bootstrap,
  }
})
