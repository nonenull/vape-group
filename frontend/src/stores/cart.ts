import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { Product, ProductSkuVariant } from '@/data/mockProducts'

const CART_STORAGE_KEY = 'vape_group_cart'

export interface CartItem {
  id: number
  productId: number
  name: string
  sku: string
  variantLabel: string
  variantSku: string
  variantStock: number | null
  price: number
  quantity: number
  image: string
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>(loadCart())

  function loadCart(): CartItem[] {
    const raw = localStorage.getItem(CART_STORAGE_KEY)
    if (!raw) {
      return []
    }

    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }

  function persistCart() {
    localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(items.value))
  }

  function addItem(product: Product, variant?: ProductSkuVariant | null, variantLabel = '') {
    const price = variant?.price ?? product.salePrice ?? product.price
    const resolvedSku = variant?.sku ?? product.sku
    const existing = items.value.find((item) => item.productId === product.id && item.variantSku === resolvedSku)
    if (existing) {
      existing.quantity += 1
      persistCart()
      return
    }

    items.value.push({
      id: Date.now(),
      productId: product.id,
      name: product.name,
      sku: resolvedSku,
      variantLabel,
      variantSku: resolvedSku,
      variantStock: variant?.stock ?? null,
      price,
      quantity: 1,
      image: product.image,
    })
    persistCart()
  }

  function updateQuantity(cartItemId: number, quantity: number) {
    const item = items.value.find((entry) => entry.id === cartItemId)
    if (!item) {
      return
    }
    if (quantity <= 0) {
      items.value = items.value.filter((entry) => entry.id !== cartItemId)
      persistCart()
      return
    }
    item.quantity = quantity
    persistCart()
  }

  function removeItem(cartItemId: number) {
    items.value = items.value.filter((entry) => entry.id !== cartItemId)
    persistCart()
  }

  function clearCart() {
    items.value = []
    persistCart()
  }

  const itemCount = computed(() => items.value.reduce((sum, item) => sum + item.quantity, 0))
  const subtotal = computed(() => items.value.reduce((sum, item) => sum + item.price * item.quantity, 0))

  return {
    items,
    itemCount,
    subtotal,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
  }
})
