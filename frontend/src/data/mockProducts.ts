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

export interface Product {
  id: number
  name: string
  sku: string
  price: number
  salePrice?: number
  category: string
  rating: number
  reviews: number
  stock: number
  image: string
  gallery: string[]
  detailImages: string[]
  badge?: string
  description: string
  longDescription: string
  flavors: string[]
  variants: ProductVariant[]
  optionGroups: ProductOptionGroup[]
  skuVariants: ProductSkuVariant[]
  specs: Array<{ label: string; value: string }>
}

export interface TenantInfo {
  id: number
  domain: string
  name: string
  isActive: boolean
  theme: string
  previewImage: string
  logoImage: string
  accentColor?: string
  surfaceColor?: string
  heroTitle?: string
  tagline?: string
  announcement?: string
  supportText?: string
  seoTitle: string
  seoDescription: string
}
