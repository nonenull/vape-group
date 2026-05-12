import apiClient from './client'

export interface Product {
  id: number
  sku: string
  base_name: string
  base_price: number
  base_stock_quantity: number
  base_images: string[]
  specifications: Record<string, any>
  is_active: boolean
  category?: string
  preview_image?: string
  gallery?: string[]
  status?: string
  description?: string
  long_description?: string
  badge?: string
  rating?: number
  reviews?: number
  flavors?: string[]
  created_at: string
  updated_at: string
}

export interface ProductWithOverride extends Product {
  custom_name?: string
  custom_description?: string
  custom_price?: number
  custom_stock_quantity?: number
  custom_images?: string[]
  variants?: Array<{ name: string; sku: string }>
  option_groups?: Array<{ name: string; values: string[] }>
  sku_variants?: Array<{ sku: string; price?: number | null; stock?: number | null; selections: Record<string, string> }>
  seo_title?: string
  seo_description?: string
}

export interface ProductListResponse {
  data: ProductWithOverride[]
  total: number
  page: number
  limit: number
}

export const productAPI = {
  // 获取商品列表
  getProducts(page = 1, limit = 20): Promise<ProductListResponse> {
    return apiClient.get('/api/products', {
      params: { page, limit },
    })
  },

  // 获取商品详情
  getProductDetail(id: number): Promise<ProductWithOverride> {
    return apiClient.get(`/api/products/${id}`)
  },

  // 创建商品（管理员）
  createProduct(data: Partial<Product>) {
    return apiClient.post('/api/products', data)
  },

  // 更新商品（管理员）
  updateProduct(id: number, data: Partial<Product>) {
    return apiClient.put(`/api/products/${id}`, data)
  },

  // 删除商品（管理员）
  deleteProduct(id: number) {
    return apiClient.delete(`/api/products/${id}`)
  },

  // 设置租户商品覆盖数据
  setProductOverrides(id: number, overrides: any) {
    return apiClient.put(`/api/products/${id}/overrides`, overrides)
  },

  // 获取租户商品覆盖数据
  getProductOverrides(id: number) {
    return apiClient.get(`/api/products/${id}/overrides`)
  },
}
