import axios from 'axios'
import { useRequestHeaders, useRuntimeConfig } from '#app'
import type {
  Brand,
  Category,
  ConvenienceStoreLocation,
  HomeBannerConfig,
  HomeSectionConfig,
  PlatformConfig,
  Product,
  ProductListApiItem,
  ProductListResponse,
  StoreOrder,
  TenantInfo,
} from '~/types/store'

interface TenantResponse {
  id: number
  domain: string
  name: string
  is_active: boolean
  theme: string
  home_template?: string
  home_module_order?: string[]
  primary_brand_id?: number | null
  preview_image: string
  logo_image: string
  accent_color?: string
  accent_strong_color?: string
  surface_color?: string
  page_bg_color?: string
  card_bg_color?: string
  text_color?: string
  muted_text_color?: string
  border_color?: string
  hero_bg_color?: string
  tag_bg_color?: string
  hero_title?: string
  tagline?: string
  announcement?: string
  support_text?: string
  seo_title: string
  seo_description: string
  home_banner?: {
    enabled?: boolean
    title?: string
    subtitle?: string
    image?: string
    link?: string
    button_text?: string
  }
  home_sections?: Array<{
    id?: string
    type?: string
    enabled?: boolean
    title?: string
    limit?: number
  }>
}

interface PlatformConfigResponse {
  id: number
  line_contact_url?: string
  faq_html?: string
  shipping_fee?: number
  free_shipping_threshold?: number
  featured_category_ids?: number[]
  featured_brand_ids?: number[]
}

interface CreateOrderItemPayload {
  product_id: number
  variant_name?: string
  variant_sku?: string
  quantity: number
}

interface CreateOrderPayload {
  items: CreateOrderItemPayload[]
  line_id: string
  phone: string
  convenience_store: string
  shipping_address: string
  payment_method: string
}

interface OrderResponse {
  id: number
  tenant_id: number
  user_id: number
  total_amount: number
  status: string
  line_id: string
  phone: string
  convenience_store: string
  shipping_address: string
  payment_method: string
  items: Array<{
    id: number
    product_id: number
    name?: string
    variant_name: string
    variant_sku: string
    quantity: number
    price: number
  }>
  created_at: string
  updated_at: string
}

interface ECPayCvsMapConfigResponse {
  action: string
  method: string
  fields: Record<string, string>
}

interface ConvenienceStoreLocationResponse {
  store_id: string
  store_name: string
  store_address: string
  store_phone: string
  city: string
  district: string
}

interface CategoryResponse {
  id: number
  name: string
  parent_id?: number | null
  sort_order: number
}

interface BrandResponse {
  id: number
  name: string
  logo_url?: string
  description?: string
}

const FALLBACK_IMAGE = '/logo.svg'

function getApiBaseURL() {
  const config = useRuntimeConfig()

  if (import.meta.server) {
    if (config.serverApiBase) {
      return config.serverApiBase.replace(/\/$/, '')
    }
    if (config.public.apiBase) {
      return config.public.apiBase.replace(/\/$/, '')
    }
    return 'http://backend:8088'
  }

  if (config.public.apiBase) {
    return config.public.apiBase.replace(/\/$/, '')
  }

  const { protocol, host, hostname, port } = window.location

  if (port === '8880') {
    return `${protocol}//${host}`
  }

  if (port === '3000') {
    return `http://${hostname}:8088`
  }

  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'http://localhost:8088'
  }

  return `${protocol}//${host}`
}

function getPublicAssetBaseURL() {
  const config = useRuntimeConfig()
  const assetBase = typeof config.public.assetBase === 'string' ? config.public.assetBase : ''
  if (assetBase) {
    return assetBase.replace(/\/$/, '')
  }

  if (import.meta.client) {
    return window.location.origin
  }

  const headers = useRequestHeaders(['x-forwarded-proto', 'host'])
  const protocol = headers['x-forwarded-proto'] || 'http'
  const host = headers.host || 'localhost:3000'
  return `${protocol}://${host}`
}

function resolveAssetURL(value: string, assetBaseURL: string) {
  if (!value) {
    return value
  }
  if (/^(https?:)?\/\//.test(value) || value.startsWith('data:')) {
    return value
  }
  if (value.startsWith('/')) {
    return `${assetBaseURL}${value}`
  }
  return value
}

function createApiClient() {
  const apiBaseURL = getApiBaseURL()
  const headers: Record<string, string> = {}
  if (import.meta.server) {
    const requestHeaders = useRequestHeaders(['host', 'x-tenant-domain'])
    const tenantDomain = (requestHeaders['x-tenant-domain'] || requestHeaders.host || '').split(':')[0]?.trim().toLowerCase() || ''
    if (tenantDomain) {
      headers['X-Tenant-Domain'] = tenantDomain
    }
  } else {
    const tenantDomain = window.location.hostname
    if (tenantDomain) {
      headers['X-Tenant-Domain'] = tenantDomain
    }
  }

  return axios.create({
    baseURL: apiBaseURL,
    timeout: 10000,
    headers,
  })
}

function normalizeProduct(input: ProductListApiItem, assetBaseURL: string): Product {
  const categoryIds = Array.isArray(input.category_ids ?? input.specifications?.categoryIds)
    ? (input.category_ids ?? input.specifications?.categoryIds)
      .map((item: any) => Number(item))
      .filter((item: number) => Number.isFinite(item) && item > 0)
    : input.category_id != null
      ? [Number(input.category_id)]
      : []

  const gallery =
    input.custom_images?.length
      ? input.custom_images
      : input.gallery?.length
        ? input.gallery
        : input.base_images?.length
          ? input.base_images
          : input.preview_image
            ? [input.preview_image]
            : [FALLBACK_IMAGE]

  const detailImages =
    input.custom_detail_images?.length
      ? input.custom_detail_images
      : input.detail_images?.length
        ? input.detail_images
        : Array.isArray(input.specifications?.detailImages)
          ? input.specifications?.detailImages as string[]
          : []

  const optionGroups = Array.isArray(input.option_groups)
    ? input.option_groups
      .map((item) => ({
        name: String(item?.name ?? '').trim(),
        values: Array.isArray(item?.values) ? item.values.map((value) => String(value).trim()).filter(Boolean) : [],
      }))
      .filter((item) => item.name && item.values.length)
    : []

  const skuVariants = Array.isArray(input.sku_variants)
    ? input.sku_variants
      .map((item) => ({
        sku: String(item?.sku ?? '').trim(),
        price: item?.price == null ? null : Number(item.price),
        stock: item?.stock == null ? null : Number(item.stock),
        selections: item?.selections && typeof item.selections === 'object' ? item.selections : {},
      }))
      .filter((item) => item.sku && Object.keys(item.selections).length)
    : []

  const description =
    input.custom_description ??
    input.description ??
    String(input.specifications?.description ?? '暫無商品描述')

  const longDescription =
    input.long_description ??
    String(input.specifications?.longDescription ?? '')

  const specificationHtml =
    input.specification_html ??
    String(input.specifications?.specificationHtml ?? '')

  return {
    id: Number(input.id),
    name: input.custom_name ?? input.base_name ?? '未命名商品',
    sku: input.sku ?? 'N/A',
    slug: input.slug?.trim() || undefined,
    price: Number(input.custom_price ?? input.base_price ?? 0),
    salePrice: undefined,
    category: input.category ?? String(input.specifications?.category ?? '未分類'),
    categoryId: input.category_id ?? null,
    categoryIds,
    brand: input.brand ?? String(input.specifications?.brand ?? ''),
    rating: Number(input.rating ?? input.specifications?.rating ?? 4.5),
    reviews: Number(input.reviews ?? input.specifications?.reviews ?? 0),
    stock: Number(input.custom_stock_quantity ?? input.base_stock_quantity ?? 0),
    image: resolveAssetURL(input.preview_image ?? gallery[0] ?? FALLBACK_IMAGE, assetBaseURL),
    gallery: gallery.map((value) => resolveAssetURL(value, assetBaseURL)),
    detailImages: detailImages.map((value) => resolveAssetURL(value, assetBaseURL)),
    badge: input.badge ?? String(input.specifications?.badge ?? ''),
    description,
    longDescription,
    specificationHtml,
    flavors: Array.isArray(input.flavors) ? input.flavors : [],
    variants: Array.isArray(input.variants) ? input.variants : [],
    optionGroups,
    skuVariants,
    seoTitle: input.seo_title?.trim() || undefined,
    seoDescription: input.seo_description?.trim() || undefined,
    isVisible: input.is_visible !== false,
    specs: [
      { label: 'SKU', value: input.sku ?? 'N/A' },
      { label: '分類', value: input.category ?? String(input.specifications?.category ?? '未分類') },
      { label: '品牌', value: input.brand ?? String(input.specifications?.brand ?? '') },
      { label: '庫存', value: String(input.custom_stock_quantity ?? input.base_stock_quantity ?? 0) },
    ],
  }
}

function normalizeCategory(input: CategoryResponse): Category {
  return {
    id: Number(input.id),
    name: String(input.name ?? '').trim(),
    parentId: input.parent_id ?? null,
    sortOrder: Number(input.sort_order ?? 0),
  }
}

function normalizeBrand(input: BrandResponse, assetBaseURL: string): Brand {
  return {
    id: Number(input.id),
    name: String(input.name ?? '').trim(),
    logoUrl: input.logo_url ? resolveAssetURL(input.logo_url, assetBaseURL) : '',
    description: input.description?.trim() ?? '',
  }
}

function normalizeTenant(input: TenantResponse): TenantInfo {
  const assetBaseURL = getPublicAssetBaseURL()
  const homeBanner: HomeBannerConfig = {
    enabled: Boolean(input.home_banner?.enabled ?? false),
    title: input.home_banner?.title?.trim() ?? '',
    subtitle: input.home_banner?.subtitle?.trim() ?? '',
    image: resolveAssetURL(input.home_banner?.image?.trim() ?? '', assetBaseURL),
    link: input.home_banner?.link?.trim() ?? '',
    buttonText: input.home_banner?.button_text?.trim() ?? '',
  }

  const homeSections: HomeSectionConfig[] = Array.isArray(input.home_sections)
    ? input.home_sections
      .map((item, index) => ({
        id: String(item?.id ?? `${item?.type ?? 'section'}-${index + 1}`).trim(),
        type: String(item?.type ?? '').trim(),
        enabled: item?.enabled !== false,
        title: String(item?.title ?? '').trim(),
        limit: Number(item?.limit ?? 0),
      }))
      .filter((item) => item.type)
    : []

  return {
    id: input.id,
    domain: input.domain,
    name: input.name,
    isActive: input.is_active,
    theme: input.theme,
    homeTemplate: input.home_template,
    homeModuleOrder: Array.isArray(input.home_module_order) ? input.home_module_order : [],
    primaryBrandId: input.primary_brand_id ?? null,
    previewImage: resolveAssetURL(input.preview_image, assetBaseURL),
    logoImage: resolveAssetURL(input.logo_image, assetBaseURL),
    accentColor: input.accent_color,
    accentStrongColor: input.accent_strong_color,
    surfaceColor: input.surface_color,
    pageBgColor: input.page_bg_color,
    cardBgColor: input.card_bg_color,
    textColor: input.text_color,
    mutedTextColor: input.muted_text_color,
    borderColor: input.border_color,
    heroBgColor: input.hero_bg_color,
    tagBgColor: input.tag_bg_color,
    heroTitle: input.hero_title,
    tagline: input.tagline,
    announcement: input.announcement,
    supportText: input.support_text,
    seoTitle: input.seo_title,
    seoDescription: input.seo_description,
    homeBanner,
    homeSections,
  }
}

function normalizePlatformConfig(input: PlatformConfigResponse): PlatformConfig {
  return {
    id: Number(input.id ?? 0),
    lineContactUrl: input.line_contact_url ?? '',
    faqHtml: input.faq_html ?? '',
    shippingFee: Number(input.shipping_fee ?? 90),
    freeShippingThreshold: Number(input.free_shipping_threshold ?? 1200),
    featuredCategoryIds: Array.isArray(input.featured_category_ids)
      ? input.featured_category_ids
        .map((item) => Number(item))
        .filter((item) => Number.isFinite(item) && item > 0)
      : [],
    featuredBrandIds: Array.isArray(input.featured_brand_ids)
      ? input.featured_brand_ids
        .map((item) => Number(item))
        .filter((item) => Number.isFinite(item) && item > 0)
      : [],
  }
}

export async function fetchProducts(
  page = 1,
  limit = 20,
  filters?: {
    keyword?: string
    category?: string | number
    brand?: string
    sort?: string
  },
) {
  const assetBaseURL = getPublicAssetBaseURL()
  const client = createApiClient()
  const response = await client.get<ProductListResponse>('/api/products', {
    params: {
      page,
      limit,
      ...(filters?.keyword ? { keyword: filters.keyword } : {}),
      ...(filters?.category ? { category: filters.category } : {}),
      ...(filters?.brand ? { brand: filters.brand } : {}),
      ...(filters?.sort && filters.sort !== 'default' ? { sort: filters.sort } : {}),
    },
  })
  const products = response.data.data
    .map((item) => normalizeProduct(item, assetBaseURL))
    .filter((item) => item.isVisible)

  return {
    products,
    total: Number(response.data.total ?? products.length),
    page: Number(response.data.page ?? page),
    limit: Number(response.data.limit ?? limit),
  }
}

export async function fetchProductDetail(id: number) {
  const assetBaseURL = getPublicAssetBaseURL()
  const client = createApiClient()
  const response = await client.get<ProductListApiItem>(`/api/products/${id}`)
  const product = normalizeProduct(response.data, assetBaseURL)
  return product.isVisible ? product : null
}

export async function fetchCategories() {
  const client = createApiClient()
  const response = await client.get<CategoryResponse[]>('/api/categories')
  return response.data.map(normalizeCategory)
}

export async function fetchBrands() {
  const assetBaseURL = getPublicAssetBaseURL()
  const client = createApiClient()
  const response = await client.get<BrandResponse[]>('/api/brands')
  return response.data.map((item) => normalizeBrand(item, assetBaseURL))
}

export async function fetchTenantBundle() {
  const client = createApiClient()
  const [tenantResponse, platformConfigResponse] = await Promise.all([
    client.get<TenantResponse>('/api/tenant/current'),
    client.get<PlatformConfigResponse>('/api/platform-config'),
  ])

  return {
    tenant: normalizeTenant(tenantResponse.data),
    platformConfig: normalizePlatformConfig(platformConfigResponse.data),
  }
}

export async function createOrder(payload: CreateOrderPayload) {
  const client = createApiClient()
  const response = await client.post<OrderResponse>('/api/orders', payload)
  return response.data
}

export async function getECPayCvsMapConfig(input: {
  returnUrl: string
  flow: 'cart' | 'checkout'
}) {
  const client = createApiClient()
  const response = await client.post<ECPayCvsMapConfigResponse>('/api/logistics/ecpay/cvs-map', {
    return_url: input.returnUrl,
    flow: input.flow,
  })
  return response.data
}

export async function fetchConvenienceStores(): Promise<ConvenienceStoreLocation[]> {
  const client = createApiClient()
  const response = await client.get<ConvenienceStoreLocationResponse[]>('/api/logistics/ecpay/stores')
  return Array.isArray(response.data)
    ? response.data.map((item) => ({
      id: String(item.store_id ?? '').trim(),
      name: String(item.store_name ?? '').trim(),
      address: String(item.store_address ?? '').trim(),
      phone: String(item.store_phone ?? '').trim(),
      city: String(item.city ?? '').trim(),
      district: String(item.district ?? '').trim(),
    })).filter((item) => item.id && item.name && item.city && item.district)
    : []
}

export async function fetchOrderDetail(id: number): Promise<StoreOrder> {
  const client = createApiClient()
  const response = await client.get<OrderResponse>(`/api/orders/${id}`)
  return {
    id: response.data.id,
    totalAmount: response.data.total_amount,
    status: response.data.status,
    lineId: response.data.line_id,
    phone: response.data.phone,
    convenienceStore: response.data.convenience_store,
    shippingAddress: response.data.shipping_address,
    paymentMethod: response.data.payment_method,
    createdAt: response.data.created_at,
    items: Array.isArray(response.data.items)
      ? response.data.items.map((item) => ({
        id: item.id,
        productId: item.product_id,
        name: item.name ?? `商品 #${item.product_id}`,
        variantName: item.variant_name,
        variantSku: item.variant_sku,
        quantity: item.quantity,
        price: item.price,
      }))
      : [],
  }
}
