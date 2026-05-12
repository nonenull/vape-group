export interface Product {
  id: number
  name: string
  sku: string
  price: number
  stock: number
  category: string
  status: string
}

export const adminProducts: Product[] = [
  {
    id: 1,
    name: 'Vape Cloud Ultra 500',
    sku: 'VCU-500',
    price: 99.99,
    stock: 42,
    category: 'Pod 系統',
    status: '在线',
  },
  {
    id: 2,
    name: 'Vape Pro Max Kit',
    sku: 'VPM-2026',
    price: 199.9,
    stock: 18,
    category: '套装',
    status: '在线',
  },
  {
    id: 3,
    name: 'Flavor Booster 50ml',
    sku: 'FBR-50',
    price: 49.0,
    stock: 126,
    category: '烟油',
    status: '缺货',
  },
]
