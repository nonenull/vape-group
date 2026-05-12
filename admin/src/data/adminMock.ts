export interface ProductVariantRecord {
  name: string
  sku: string
}

export interface ProductOptionGroupRecord {
  name: string
  values: string[]
}

export interface ProductSkuVariantRecord {
  sku: string
  price?: number | null
  stock?: number | null
  selections: Record<string, string>
}

export interface TenantRecord {
  id: number
  domain: string
  boundDomains: string[]
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

export interface ProductRecord {
  id: number
  sku: string
  baseName: string
  basePrice: number
  baseStockQuantity: number
  category: string
  categoryId?: number | null
  brand?: string
  brandId?: number | null
  previewImage: string
  gallery: string[]
  detailImages: string[]
  status: '上架中' | '草稿' | '缺貨'
  description?: string
  longDescription?: string
  badge?: string
  rating?: number
  reviews?: number
  flavors?: string[]
  variants?: ProductVariantRecord[]
  optionGroups?: ProductOptionGroupRecord[]
  skuVariants?: ProductSkuVariantRecord[]
  isActive?: boolean
  updatedAt: string
}

export interface CategoryRecord {
  id: number
  tenantId: number
  name: string
  parentId?: number | null
  sortOrder: number
}

export interface BrandRecord {
  id: number
  tenantId: number
  name: string
  logoUrl?: string
  description?: string
}

export interface ProductOverrideRecord {
  id: number
  tenantId: number
  productId: number
  customName: string
  customDescription: string
  customPrice: number | null
  customStockQuantity: number | null
  customImages?: string[]
  customDetailImages?: string[]
  seoTitle: string
  seoDescription: string
  isVisible: boolean
}

export interface OrderRecord {
  id: number
  tenantId: number
  orderNo: string
  customerName: string
  totalAmount: number
  status: '待付款' | '已付款' | '已出貨' | '已完成'
  paymentMethod: string
  createdAt: string
  items?: Array<{
    name: string
    sku: string
    variantLabel?: string
    quantity: number
    price: number
  }>
}

export const orderRecords: OrderRecord[] = [
  {
    id: 1,
    tenantId: 1,
    orderNo: 'VG20260507001',
    customerName: '陳先生',
    totalAmount: 1890,
    status: '已付款',
    paymentMethod: '信用卡',
    createdAt: '2026-05-07 09:20',
    items: [
      {
        name: 'NOVA Pod 煙彈',
        sku: 'POD-GRAPE-2',
        variantLabel: '口味：冰葡萄 / 盒裝：2入',
        quantity: 2,
        price: 299,
      },
    ],
  },
  {
    id: 2,
    tenantId: 2,
    orderNo: 'VG20260507002',
    customerName: '林小姐',
    totalAmount: 990,
    status: '待付款',
    paymentMethod: 'ATM 轉帳',
    createdAt: '2026-05-07 10:05',
    items: [
      {
        name: 'NOVA Pod 煙彈',
        sku: 'POD-TOBACCO-4',
        variantLabel: '口味：經典菸草 / 盒裝：4入',
        quantity: 1,
        price: 499,
      },
    ],
  },
  {
    id: 3,
    tenantId: 1,
    orderNo: 'VG20260506087',
    customerName: '王先生',
    totalAmount: 2590,
    status: '已出貨',
    paymentMethod: '信用卡',
    createdAt: '2026-05-06 16:40',
  },
]
