import { defineStore } from 'pinia'
import type { Product, ProductSkuVariant } from '~/types/store'

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

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: [] as CartItem[],
    hydrated: false,
  }),
  getters: {
    itemCount: (state) => state.items.reduce((sum, item) => sum + item.quantity, 0),
    subtotal: (state) => state.items.reduce((sum, item) => sum + item.price * item.quantity, 0),
  },
  actions: {
    hydrate() {
      if (!import.meta.client || this.hydrated) {
        return
      }

      const raw = localStorage.getItem(CART_STORAGE_KEY)
      if (raw) {
        try {
          const parsed = JSON.parse(raw)
          this.items = Array.isArray(parsed) ? parsed : []
        } catch {
          this.items = []
        }
      }
      this.hydrated = true
    },
    persist() {
      if (!import.meta.client) {
        return
      }
      localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(this.items))
    },
    addItem(product: Product, variant?: ProductSkuVariant | null, variantLabel = '') {
      this.hydrate()
      const price = variant?.price ?? product.salePrice ?? product.price
      const resolvedSku = variant?.sku ?? product.sku
      const existing = this.items.find((item) => item.productId === product.id && item.variantSku === resolvedSku)
      if (existing) {
        existing.quantity += 1
        this.persist()
        return
      }

      this.items.push({
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
      this.persist()
    },
    updateQuantity(cartItemId: number, quantity: number) {
      this.hydrate()
      const item = this.items.find((entry) => entry.id === cartItemId)
      if (!item) {
        return
      }
      if (quantity <= 0) {
        this.items = this.items.filter((entry) => entry.id !== cartItemId)
        this.persist()
        return
      }
      item.quantity = quantity
      this.persist()
    },
    removeItem(cartItemId: number) {
      this.hydrate()
      this.items = this.items.filter((entry) => entry.id !== cartItemId)
      this.persist()
    },
    clearCart() {
      this.items = []
      this.persist()
    },
  },
})
