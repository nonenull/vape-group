import type {
  BrandRecord,
  CategoryRecord,
  DomainRecord,
  DomainDnsRecord,
  OrderRecord,
  PlatformConfigRecord,
  ProductOptionGroupRecord,
  ProductOverrideRecord,
  BulkGeneratedOverrideNameRecord,
  ProductRecord,
  ProductSkuVariantRecord,
  ProductVariantRecord,
  TenantRecord,
} from '@/data/adminMock'

const getApiBaseURL = () => {
  const envBaseURL = import.meta.env.VITE_API_URL?.trim()
  if (envBaseURL) {
    return envBaseURL.replace(/\/$/, '')
  }

  const { protocol, host, hostname, port } = window.location

  if (port === '8880') {
    return `${protocol}//${host}`
  }

  // In local Vite dev, keep the current hostname and send API traffic to the backend port.
  if (port === '5173' || port === '5174') {
    return `http://${hostname}:8088`
  }

  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'http://localhost:8088'
  }

  return `${protocol}//${host}`
}

export const resolveAssetURL = (value: string) => {
  if (!value) {
    return value
  }
  if (/^(https?:)?\/\//.test(value) || value.startsWith('data:')) {
    return value
  }
  if (value.startsWith('/')) {
    return `${getApiBaseURL()}${value}`
  }
  return value
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${getApiBaseURL()}${path}`, {
    ...init,
    headers: {
      ...(init?.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
      ...(init?.headers ?? {}),
    },
  })

  if (!response.ok) {
    let message = `Request failed: ${response.status}`
    try {
      const payload = await response.json()
      if (payload?.error) {
        message = payload.error
      }
    } catch {
      // Ignore JSON parsing failures and keep the status message.
    }
    throw new Error(message)
  }

  return response.json() as Promise<T>
}

const mapTenant = (input: any): TenantRecord => ({
  id: Number(input.id),
  domain: input.domain ?? '',
  boundDomains: input.bound_domains ?? input.boundDomains ?? [],
  name: input.name ?? '',
  isActive: Boolean(input.is_active ?? input.isActive),
  theme: input.theme ?? '',
  homeTemplate: input.home_template ?? input.homeTemplate ?? '',
  homeModuleOrder: input.home_module_order ?? input.homeModuleOrder ?? [],
  primaryBrandId: input.primary_brand_id ?? input.primaryBrandId ?? null,
  previewImage: input.preview_image ?? input.previewImage ?? '',
  logoImage: input.logo_image ?? input.logoImage ?? '',
  accentColor: input.accent_color ?? input.accentColor ?? '',
  accentStrongColor: input.accent_strong_color ?? input.accentStrongColor ?? '',
  surfaceColor: input.surface_color ?? input.surfaceColor ?? '',
  pageBgColor: input.page_bg_color ?? input.pageBgColor ?? '',
  cardBgColor: input.card_bg_color ?? input.cardBgColor ?? '',
  textColor: input.text_color ?? input.textColor ?? '',
  mutedTextColor: input.muted_text_color ?? input.mutedTextColor ?? '',
  borderColor: input.border_color ?? input.borderColor ?? '',
  heroBgColor: input.hero_bg_color ?? input.heroBgColor ?? '',
  tagBgColor: input.tag_bg_color ?? input.tagBgColor ?? '',
  heroTitle: input.hero_title ?? input.heroTitle ?? '',
  tagline: input.tagline ?? '',
  announcement: input.announcement ?? '',
  supportText: input.support_text ?? input.supportText ?? '',
  seoTitle: input.seo_title ?? input.seoTitle ?? '',
  seoDescription: input.seo_description ?? input.seoDescription ?? '',
})

const tenantPayload = (payload: Omit<TenantRecord, 'id'>) => ({
  domain: payload.domain,
  bound_domains: payload.boundDomains,
  name: payload.name,
  is_active: payload.isActive,
  theme: payload.theme,
  home_template: payload.homeTemplate,
  home_module_order: payload.homeModuleOrder ?? [],
  primary_brand_id: payload.primaryBrandId ?? null,
  preview_image: payload.previewImage,
  logo_image: payload.logoImage,
  accent_color: payload.accentColor,
  accent_strong_color: payload.accentStrongColor,
  surface_color: payload.surfaceColor,
  page_bg_color: payload.pageBgColor,
  card_bg_color: payload.cardBgColor,
  text_color: payload.textColor,
  muted_text_color: payload.mutedTextColor,
  border_color: payload.borderColor,
  hero_bg_color: payload.heroBgColor,
  tag_bg_color: payload.tagBgColor,
  hero_title: payload.heroTitle,
  tagline: payload.tagline,
  announcement: payload.announcement,
  support_text: payload.supportText,
  seo_title: payload.seoTitle,
  seo_description: payload.seoDescription,
})

const mapProduct = (input: any): ProductRecord => ({
  id: Number(input.id),
  sku: input.sku ?? '',
  slug: input.slug ?? '',
  baseName: input.base_name ?? input.baseName ?? '',
  basePrice: Number(input.base_price ?? input.basePrice ?? 0),
  baseStockQuantity: Number(input.base_stock_quantity ?? input.baseStockQuantity ?? 0),
  category: input.category ?? '',
  categoryId: input.category_id ?? input.categoryId ?? null,
  categoryIds: Array.isArray(input.category_ids ?? input.categoryIds ?? input.specifications?.categoryIds)
    ? (input.category_ids ?? input.categoryIds ?? input.specifications?.categoryIds)
      .map((item: any) => Number(item))
      .filter((item: number) => Number.isFinite(item) && item > 0)
    : (input.category_id ?? input.categoryId ?? null) != null
      ? [Number(input.category_id ?? input.categoryId)]
      : [],
  brand: input.brand ?? '',
  brandId: input.brand_id ?? input.brandId ?? null,
  previewImage: input.preview_image ?? input.previewImage ?? '',
  gallery: input.gallery ?? [],
  detailImages: input.detail_images ?? input.detailImages ?? [],
  status: input.status ?? '草稿',
  description: input.description ?? '',
  longDescription: input.long_description ?? input.longDescription ?? '',
  specificationHtml:
    input.specification_html ??
    input.specificationHtml ??
    input.specifications?.specificationHtml ??
    '',
  badge: input.badge ?? '',
  rating: Number(input.rating ?? 0),
  reviews: Number(input.reviews ?? 0),
  flavors: input.flavors ?? [],
  variants: Array.isArray(input.variants)
    ? input.variants
      .map((item: any): ProductVariantRecord => ({
        name: String(item?.name ?? '').trim(),
        sku: String(item?.sku ?? '').trim(),
      }))
      .filter((item: ProductVariantRecord) => item.name && item.sku)
    : [],
  optionGroups: Array.isArray(input.option_groups ?? input.optionGroups)
    ? (input.option_groups ?? input.optionGroups)
      .map((item: any): ProductOptionGroupRecord => ({
        name: String(item?.name ?? '').trim(),
        values: Array.isArray(item?.values) ? item.values.map((value: any) => String(value).trim()).filter(Boolean) : [],
      }))
      .filter((item: ProductOptionGroupRecord) => item.name && item.values.length)
    : [],
  skuVariants: Array.isArray(input.sku_variants ?? input.skuVariants)
    ? (input.sku_variants ?? input.skuVariants)
      .map((item: any): ProductSkuVariantRecord => ({
        sku: String(item?.sku ?? '').trim(),
        price: item?.price ?? null,
        stock: item?.stock ?? null,
        selections: item?.selections && typeof item.selections === 'object' ? item.selections : {},
      }))
      .filter((item: ProductSkuVariantRecord) => item.sku && Object.keys(item.selections).length)
    : [],
  isActive: Boolean(input.is_active ?? input.isActive),
  updatedAt: input.updated_at ?? input.updatedAt ?? '',
})

const productPayload = (payload: Omit<ProductRecord, 'id' | 'updatedAt'>) => ({
  sku: payload.sku,
  base_name: payload.baseName,
  base_price: payload.basePrice,
  base_stock_quantity: payload.baseStockQuantity,
  category: payload.category,
  category_id: payload.categoryId ?? null,
  category_ids: payload.categoryIds ?? (payload.categoryId != null ? [payload.categoryId] : []),
  brand: payload.brand ?? '',
  brand_id: payload.brandId ?? null,
  preview_image: payload.previewImage,
  gallery: payload.gallery,
  detail_images: payload.detailImages,
  status: payload.status,
  description: payload.description,
  long_description: payload.longDescription,
  specification_html: payload.specificationHtml,
  badge: payload.badge,
  rating: payload.rating,
  reviews: payload.reviews,
  flavors: payload.flavors,
  variants: payload.variants ?? [],
  option_groups: payload.optionGroups ?? [],
  sku_variants: payload.skuVariants ?? [],
  is_active: payload.isActive ?? true,
})

const mapOverride = (input: any, tenantId = 0, productId = 0): ProductOverrideRecord => ({
  id: Number(input.id ?? 0),
  tenantId: Number(input.tenant_id ?? input.tenantId ?? tenantId),
  productId: Number(input.product_id ?? input.productId ?? productId),
  customName: input.custom_name ?? input.customName ?? '',
  customDescription: input.custom_description ?? input.customDescription ?? '',
  customPrice: input.custom_price ?? input.customPrice ?? null,
  customStockQuantity: input.custom_stock_quantity ?? input.customStockQuantity ?? null,
  customImages: input.custom_images ?? input.customImages ?? [],
  customDetailImages: input.custom_detail_images ?? input.customDetailImages ?? [],
  seoTitle: input.seo_title ?? input.seoTitle ?? '',
  seoDescription: input.seo_description ?? input.seoDescription ?? '',
  isVisible: Boolean(input.is_visible ?? input.isVisible ?? true),
})

const mapCategory = (input: any): CategoryRecord => ({
  id: Number(input.id),
  name: input.name ?? '',
  parentId: input.parent_id ?? input.parentId ?? null,
  sortOrder: Number(input.sort_order ?? input.sortOrder ?? 0),
})

const mapBrand = (input: any): BrandRecord => ({
  id: Number(input.id),
  name: input.name ?? '',
  logoUrl: input.logo_url ?? input.logoUrl ?? '',
  description: input.description ?? '',
})

const mapDomain = (input: any): DomainRecord => ({
  id: Number(input.id ?? 0),
  domainName: input.domain_name ?? input.domainName ?? '',
  registrar: input.registrar ?? 'manual',
  expireDate: input.expire_date ?? input.expireDate ?? null,
  dnsRecords: Array.isArray(input.dns_records)
    ? input.dns_records.map((item: any): DomainDnsRecord => ({
      type: item.type ?? '',
      name: item.name ?? '',
      data: item.data ?? '',
      ttl: Number(item.ttl ?? 600),
    }))
    : [],
  isBlocked: Boolean(input.is_blocked ?? input.isBlocked ?? false),
  lastCheckIp: input.last_check_ip ?? input.lastCheckIp ?? null,
  lastCheckedAt: input.last_checked_at ?? input.lastCheckedAt ?? null,
})

const mapPlatformConfig = (input: any): PlatformConfigRecord => ({
  id: Number(input.id ?? 0),
  lineContactUrl: input.line_contact_url ?? input.lineContactUrl ?? '',
  faqHtml: input.faq_html ?? input.faqHtml ?? '',
  shippingFee: Number(input.shipping_fee ?? input.shippingFee ?? 90),
  freeShippingThreshold: Number(input.free_shipping_threshold ?? input.freeShippingThreshold ?? 1200),
  featuredCategoryIds: Array.isArray(input.featured_category_ids ?? input.featuredCategoryIds)
    ? (input.featured_category_ids ?? input.featuredCategoryIds)
      .map((item: any) => Number(item))
      .filter((item: number) => Number.isFinite(item) && item > 0)
    : [],
  featuredBrandIds: Array.isArray(input.featured_brand_ids ?? input.featuredBrandIds)
    ? (input.featured_brand_ids ?? input.featuredBrandIds)
      .map((item: any) => Number(item))
      .filter((item: number) => Number.isFinite(item) && item > 0)
    : [],
})

const mapOrder = (input: any): OrderRecord => ({
  id: Number(input.id ?? 0),
  tenantId: Number(input.tenant_id ?? input.tenantId ?? 0),
  orderNo: input.order_no ?? input.orderNo ?? `#${input.id ?? 0}`,
  customerName: input.line_id ?? input.customerName ?? '未提供',
  totalAmount: Number(input.total_amount ?? input.totalAmount ?? 0),
  status: input.status ?? '已下单',
  paymentMethod: input.payment_method ?? input.paymentMethod ?? '',
  createdAt: input.created_at ?? input.createdAt ?? '',
  items: Array.isArray(input.items)
    ? input.items.map((item: any) => ({
      name: item.name ?? `商品 #${item.product_id ?? item.productId ?? ''}`,
      sku: item.variant_sku ?? item.sku ?? '',
      variantLabel: item.variant_name ?? item.variantLabel ?? '',
      quantity: Number(item.quantity ?? 0),
      price: Number(item.price ?? 0),
    }))
    : [],
})

export const adminAPI = {
  async uploadImage(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    return request<{ url: string }>('/api/admin/uploads/images', {
      method: 'POST',
      body: formData,
    })
  },
  async getTenants() {
    return (await request<any[]>('/api/admin/tenants')).map(mapTenant)
  },
  async getPlatformConfig() {
    return mapPlatformConfig(await request<any>('/api/admin/platform-config'))
  },
  async updatePlatformConfig(payload: Omit<PlatformConfigRecord, 'id'>) {
    return mapPlatformConfig(await request<any>('/api/admin/platform-config', {
      method: 'PUT',
      body: JSON.stringify({
        line_contact_url: payload.lineContactUrl,
        faq_html: payload.faqHtml ?? '',
        shipping_fee: payload.shippingFee,
        free_shipping_threshold: payload.freeShippingThreshold,
        featured_category_ids: payload.featuredCategoryIds,
        featured_brand_ids: payload.featuredBrandIds,
      }),
    }))
  },
  async createTenant(payload: Omit<TenantRecord, 'id'>) {
    return mapTenant(await request<any>('/api/admin/tenants', {
      method: 'POST',
      body: JSON.stringify(tenantPayload(payload)),
    }))
  },
  async updateTenant(id: number, payload: Omit<TenantRecord, 'id'>) {
    return mapTenant(await request<any>(`/api/admin/tenants/${id}`, {
      method: 'PUT',
      body: JSON.stringify(tenantPayload(payload)),
    }))
  },
  deleteTenant(id: number) {
    return request<{ success: boolean }>(`/api/admin/tenants/${id}`, {
      method: 'DELETE',
    })
  },
  async getProducts() {
    return (await request<any[]>('/api/admin/products')).map(mapProduct)
  },
  async createProduct(payload: Omit<ProductRecord, 'id' | 'updatedAt'>) {
    return mapProduct(await request<any>('/api/admin/products', {
      method: 'POST',
      body: JSON.stringify(productPayload(payload)),
    }))
  },
  async updateProduct(id: number, payload: Omit<ProductRecord, 'id' | 'updatedAt'>) {
    return mapProduct(await request<any>(`/api/admin/products/${id}`, {
      method: 'PUT',
      body: JSON.stringify(productPayload(payload)),
    }))
  },
  async bulkUpdateProducts(productIds: number[], payload: { status?: ProductRecord['status'], isActive?: boolean }) {
    return (await request<any[]>('/api/admin/products/bulk-update', {
      method: 'PUT',
      body: JSON.stringify({
        product_ids: productIds,
        status: payload.status,
        is_active: payload.isActive,
      }),
    })).map(mapProduct)
  },
  deleteProduct(id: number) {
    return request<{ success: boolean }>(`/api/admin/products/${id}`, {
      method: 'DELETE',
    })
  },
  async getProductOverride(productId: number, tenantId: number) {
    return mapOverride(await request<any>(`/api/admin/products/${productId}/overrides/${tenantId}`), tenantId, productId)
  },
  async updateProductOverride(productId: number, tenantId: number, payload: Omit<ProductOverrideRecord, 'id' | 'tenantId' | 'productId'> & { customImages?: string[], customDetailImages?: string[] }) {
    return mapOverride(await request<any>(`/api/admin/products/${productId}/overrides/${tenantId}`, {
      method: 'PUT',
      body: JSON.stringify({
        custom_name: payload.customName,
        custom_description: payload.customDescription,
        custom_price: payload.customPrice,
        custom_stock_quantity: payload.customStockQuantity,
        custom_images: payload.customImages ?? [],
        custom_detail_images: payload.customDetailImages ?? [],
        seo_title: payload.seoTitle,
        seo_description: payload.seoDescription,
        is_visible: payload.isVisible,
      }),
    }))
  },
  async bulkGenerateCustomNames(productIds: number[], tenantId: number, instruction: string) {
    return (await request<any[]>('/api/admin/products/bulk-generate-custom-names', {
      method: 'POST',
      body: JSON.stringify({
        product_ids: productIds,
        tenant_id: tenantId,
        instruction,
      }),
    })).map((item) => ({
      productId: Number(item.product_id ?? 0),
      tenantId: Number(item.tenant_id ?? 0),
      customName: item.custom_name ?? '',
    } satisfies BulkGeneratedOverrideNameRecord))
  },
  async getCategories() {
    return (await request<any[]>('/api/admin/categories')).map(mapCategory)
  },
  async createCategory(payload: Omit<CategoryRecord, 'id'>) {
    return mapCategory(await request<any>('/api/admin/categories', {
      method: 'POST',
      body: JSON.stringify({
        name: payload.name,
        parent_id: payload.parentId ?? null,
        sort_order: payload.sortOrder,
      }),
    }))
  },
  async updateCategory(categoryId: number, payload: Omit<CategoryRecord, 'id'>) {
    return mapCategory(await request<any>(`/api/admin/categories/${categoryId}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: payload.name,
        parent_id: payload.parentId ?? null,
        sort_order: payload.sortOrder,
      }),
    }))
  },
  deleteCategory(categoryId: number) {
    return request<{ success: boolean }>(`/api/admin/categories/${categoryId}`, {
      method: 'DELETE',
    })
  },
  async getBrands() {
    return (await request<any[]>('/api/admin/brands')).map(mapBrand)
  },
  async createBrand(payload: Omit<BrandRecord, 'id'>) {
    return mapBrand(await request<any>('/api/admin/brands', {
      method: 'POST',
      body: JSON.stringify({
        name: payload.name,
        logo_url: payload.logoUrl ?? '',
        description: payload.description ?? '',
      }),
    }))
  },
  async updateBrand(brandId: number, payload: Omit<BrandRecord, 'id'>) {
    return mapBrand(await request<any>(`/api/admin/brands/${brandId}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: payload.name,
        logo_url: payload.logoUrl ?? '',
        description: payload.description ?? '',
      }),
    }))
  },
  deleteBrand(brandId: number) {
    return request<{ success: boolean }>(`/api/admin/brands/${brandId}`, {
      method: 'DELETE',
    })
  },
  async getDomains(search = '') {
    return (await request<any[]>(`/api/admin/domains${search ? `?search=${encodeURIComponent(search)}` : ''}`)).map(mapDomain)
  },
  async createDomain(payload: Omit<DomainRecord, 'id'>) {
    return mapDomain(await request<any>('/api/admin/domains', {
      method: 'POST',
      body: JSON.stringify({
        domain_name: payload.domainName,
        registrar: payload.registrar,
        expire_date: payload.expireDate || null,
      }),
    }))
  },
  async updateDomain(domainId: number, payload: Omit<DomainRecord, 'id'>) {
    return mapDomain(await request<any>(`/api/admin/domains/${domainId}`, {
      method: 'PUT',
      body: JSON.stringify({
        domain_name: payload.domainName,
        registrar: payload.registrar,
        expire_date: payload.expireDate || null,
      }),
    }))
  },
  deleteDomain(domainId: number) {
    return request<{ success: boolean }>(`/api/admin/domains/${domainId}`, {
      method: 'DELETE',
    })
  },
  syncDomains(ip?: string) {
    return request<{ status: string; created: number; updated: number; total: number }>('/api/admin/domains/sync', {
      method: 'POST',
      body: JSON.stringify({ ip: ip || null }),
    })
  },
  checkDomainDns() {
    return request<{ checked: number; blocked: number; total: number }>('/api/admin/domains/check-dns', {
      method: 'POST',
      body: JSON.stringify({}),
    })
  },
  getDomainDnsRecords(domainId: number) {
    return request<any[]>(`/api/admin/domains/${domainId}/dns`).then((records) =>
      records.map((item): DomainDnsRecord => ({
        type: item.type ?? '',
        name: item.name ?? '',
        data: item.data ?? '',
        ttl: Number(item.ttl ?? 600),
      })),
    )
  },
  updateDomainDnsRecords(domainId: number, records: DomainDnsRecord[]) {
    return request<{ success: boolean }>(`/api/admin/domains/${domainId}/dns`, {
      method: 'PUT',
      body: JSON.stringify(records),
    })
  },
  async getOrders() {
    return (await request<any[]>('/api/admin/orders')).map(mapOrder)
  },
}
