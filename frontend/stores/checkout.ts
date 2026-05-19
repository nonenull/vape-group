import { defineStore } from 'pinia'
import type { Product, ProductSkuVariant } from '~/types/store'

const DIRECT_CHECKOUT_STORAGE_KEY = 'vape_group_direct_checkout'

export interface CheckoutItem {
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

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    directItem: null as CheckoutItem | null,
    hydrated: false,
  }),
  getters: {
    hasDirectItem: (state) => Boolean(state.directItem),
    subtotal: (state) => (state.directItem ? state.directItem.price * state.directItem.quantity : 0),
  },
  actions: {
    hydrate() {
      if (!import.meta.client || this.hydrated) {
        return
      }

      const raw = localStorage.getItem(DIRECT_CHECKOUT_STORAGE_KEY)
      if (raw) {
        try {
          const parsed = JSON.parse(raw)
          this.directItem = parsed && typeof parsed === 'object' ? parsed as CheckoutItem : null
        } catch {
          this.directItem = null
        }
      }
      this.hydrated = true
    },
    persist() {
      if (!import.meta.client) {
        return
      }
      if (this.directItem) {
        localStorage.setItem(DIRECT_CHECKOUT_STORAGE_KEY, JSON.stringify(this.directItem))
        return
      }
      localStorage.removeItem(DIRECT_CHECKOUT_STORAGE_KEY)
    },
    startDirectCheckout(product: Product, variant?: ProductSkuVariant | null, variantLabel = '') {
      this.hydrate()
      const price = variant?.price ?? product.salePrice ?? product.price
      const resolvedSku = variant?.sku ?? product.sku

      this.directItem = {
        productId: product.id,
        name: product.name,
        sku: resolvedSku,
        variantLabel,
        variantSku: resolvedSku,
        variantStock: variant?.stock ?? null,
        price,
        quantity: 1,
        image: product.image,
      }
      this.persist()
    },
    updateDirectQuantity(quantity: number) {
      this.hydrate()
      if (!this.directItem) {
        return
      }
      this.directItem.quantity = quantity <= 0 ? 1 : quantity
      this.persist()
    },
    clearDirectCheckout() {
      this.directItem = null
      this.persist()
    },
  },
})
