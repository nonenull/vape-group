export interface ProductVariant {
  name: string
  sku: string
}

export interface ProductOptionGroup {
  name: string
  values: string[]
}

export interface ProductSkuVariant {
  sku: string
  price?: number | null
  stock?: number | null
  selections: Record<string, string>
}

export interface ProductSpec {
  label: string
  value: string
}

export interface Product {
  id: number
  name: string
  sku: string
  slug?: string
  price: number
  salePrice?: number
  category: string
  categoryId?: number | null
  categoryIds?: number[]
  brand?: string
  rating: number
  reviews: number
  stock: number
  image: string
  gallery: string[]
  detailImages: string[]
  badge?: string
  description: string
  longDescription: string
  specificationHtml: string
  flavors: string[]
  variants: ProductVariant[]
  optionGroups: ProductOptionGroup[]
  skuVariants: ProductSkuVariant[]
  seoTitle?: string
  seoDescription?: string
  isVisible: boolean
  specs: ProductSpec[]
}

export interface TenantInfo {
  id: number
  domain: string
  name: string
  isActive: boolean
  theme: string
  homeTemplate?: string
  homeModuleOrder?: string[]
  primaryBrandId?: number | null
  previewImage: string
  logoImage: string
  accentColor?: string
  accentStrongColor?: string
  surfaceColor?: string
  pageBgColor?: string
  cardBgColor?: string
  textColor?: string
  mutedTextColor?: string
  borderColor?: string
  heroBgColor?: string
  tagBgColor?: string
  heroTitle?: string
  tagline?: string
  announcement?: string
  supportText?: string
  seoTitle: string
  seoDescription: string
}

export interface PlatformConfig {
  id: number
  lineContactUrl: string
  faqHtml: string
  shippingFee: number
  freeShippingThreshold: number
  featuredCategoryIds: number[]
  featuredBrandIds: number[]
}

export interface Category {
  id: number
  name: string
  parentId?: number | null
  sortOrder: number
}

export interface Brand {
  id: number
  name: string
  logoUrl?: string
  description?: string
}

export interface StoreOrderItem {
  id: number
  productId: number
  name: string
  variantName: string
  variantSku: string
  quantity: number
  price: number
}

export interface StoreOrder {
  id: number
  totalAmount: number
  status: string
  lineId: string
  phone: string
  convenienceStore: string
  shippingAddress: string
  paymentMethod: string
  items: StoreOrderItem[]
  createdAt: string
}

export interface ProductListApiItem {
  id: number
  sku: string
  slug?: string
  base_name?: string
  base_price?: number
  base_stock_quantity?: number
  base_images?: string[]
  detail_images?: string[]
  specifications?: Record<string, any>
  is_active?: boolean
  category?: string
  category_id?: number | null
  category_ids?: number[]
  brand?: string
  preview_image?: string
  gallery?: string[]
  status?: string
  description?: string
  long_description?: string
  specification_html?: string
  badge?: string
  rating?: number
  reviews?: number
  flavors?: string[]
  variants?: Array<{ name: string; sku: string }>
  option_groups?: Array<{ name: string; values: string[] }>
  sku_variants?: Array<{ sku: string; price?: number | null; stock?: number | null; selections: Record<string, string> }>
  custom_name?: string
  custom_description?: string
  custom_price?: number
  custom_stock_quantity?: number
  custom_images?: string[]
  custom_detail_images?: string[]
  seo_title?: string
  seo_description?: string
  is_visible?: boolean
}

export interface ProductListResponse {
  data: ProductListApiItem[]
  total: number
  page: number
  limit: number
}
