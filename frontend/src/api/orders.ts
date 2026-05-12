import apiClient from './client'

export interface CreateOrderItemPayload {
  product_id: number
  variant_name?: string
  variant_sku?: string
  quantity: number
}

export interface CreateOrderPayload {
  items: CreateOrderItemPayload[]
  line_id: string
  phone: string
  convenience_store: string
  shipping_address: string
  payment_method: string
}

export interface OrderItemResponse {
  id: number
  product_id: number
  variant_name: string
  variant_sku: string
  quantity: number
  price: number
}

export interface OrderResponse {
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
  items: OrderItemResponse[]
  created_at: string
  updated_at: string
}

export const orderAPI = {
  createOrder(payload: CreateOrderPayload): Promise<OrderResponse> {
    return apiClient.post('/api/orders', payload)
  },
}
