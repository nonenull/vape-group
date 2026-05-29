<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useStoreSeo } from '~/composables/useStoreSeo'
import { createOrder, fetchProducts } from '~/composables/useStoreApi'
import { useCartStore } from '~/stores/cart'
import { useTenantStore } from '~/stores/tenant'

const router = useRouter()
const cartStore = useCartStore()
const tenantStore = useTenantStore()
cartStore.hydrate()
const { products } = await fetchProducts(1, 200)
await tenantStore.initTenant()

const shippingFee = computed(() => {
  const threshold = tenantStore.platformConfig.freeShippingThreshold || 0
  const fee = tenantStore.platformConfig.shippingFee || 0
  return cartStore.subtotal >= threshold || cartStore.itemCount === 0 ? 0 : fee
})
const finalTotal = computed(() => cartStore.subtotal + shippingFee.value)
const checkoutForm = reactive({
  lineId: '',
  phone: '',
  convenienceStore: '',
  paymentMethod: 'cash_on_delivery',
})
const submitting = ref(false)
const orderSuccess = ref<{ id: number; total: number } | null>(null)
const productMap = new Map(products.map((product) => [product.id, product]))
const cartVariantOptionsMap = computed(() =>
  new Map(
    cartStore.items.map((item) => {
      const product = productMap.get(item.productId)
      const options = product?.skuVariants.map((variant) => {
        const variantLabel = product.optionGroups
          .map((group) => {
            const value = variant.selections[group.name]
            return value ? `${group.name}：${value}` : ''
          })
          .filter(Boolean)
          .join(' / ')

        return {
          sku: variant.sku,
          label: variantLabel || variant.sku,
          stock: variant.stock == null ? product.stock : variant.stock,
          rawStock: variant.stock ?? null,
          price: variant.price ?? product.salePrice ?? product.price,
        }
      }) ?? []

      return [item.id, options]
    }),
  ),
)
const cartValidation = computed(() =>
  cartStore.items.map((item) => {
    const product = productMap.get(item.productId)
    if (!product) {
      return {
        cartItemId: item.id,
        isInvalid: true,
        message: '商品已不存在，請移除後再下單。',
      }
    }

    const matchedVariant = product.skuVariants.find((variant) => variant.sku === item.variantSku) ?? null
    const availableStock = matchedVariant?.stock == null ? product.stock : matchedVariant.stock

    if (item.quantity > availableStock) {
      return {
        cartItemId: item.id,
        isInvalid: true,
        message: availableStock > 0
          ? `目前庫存僅剩 ${availableStock} 件，請調整數量。`
          : '目前庫存不足，請移除或更換規格。',
      }
    }

    return {
      cartItemId: item.id,
      isInvalid: false,
      message: '',
    }
  }),
)
const cartValidationMap = computed(() => new Map(cartValidation.value.map((item) => [item.cartItemId, item])))
const hasInvalidCartItems = computed(() => cartValidation.value.some((item) => item.isInvalid))

const updateCartItemVariant = (cartItemId: number, nextSku: string) => {
  const options = cartVariantOptionsMap.value.get(cartItemId) ?? []
  const selected = options.find((option) => option.sku === nextSku)
  if (!selected) {
    return
  }

  cartStore.updateVariant(cartItemId, {
    sku: selected.sku,
    variantLabel: selected.label,
    variantStock: selected.rawStock,
    price: selected.price,
  })
}

useStoreSeo({
  title: `購物車${cartStore.itemCount ? ` (${cartStore.itemCount})` : ''} | Vape Group 商城`,
  description: '檢視購物車商品、填寫顧客資料並送出訂單。',
  canonicalPath: '/cart',
  type: 'website',
  robots: 'noindex,nofollow',
})

const checkout = async () => {
  if (!cartStore.items.length) {
    return
  }
  if (hasInvalidCartItems.value) {
    window.alert('購物車中有庫存不足的商品，請先調整後再下單')
    return
  }
  if (!checkoutForm.lineId.trim()) {
    window.alert('請填寫 Line ID')
    return
  }
  if (!checkoutForm.phone.trim()) {
    window.alert('請填寫聯絡電話')
    return
  }
  if (!checkoutForm.convenienceStore.trim()) {
    window.alert('請填寫 7-11 門市')
    return
  }

  submitting.value = true
  try {
    const order = await createOrder({
      items: cartStore.items.map((item) => ({
        product_id: item.productId,
        variant_name: item.variantLabel,
        variant_sku: item.variantSku,
        quantity: item.quantity,
      })),
      line_id: checkoutForm.lineId.trim(),
      phone: checkoutForm.phone.trim(),
      convenience_store: checkoutForm.convenienceStore.trim(),
      shipping_address: checkoutForm.convenienceStore.trim(),
      payment_method: checkoutForm.paymentMethod,
    })

    orderSuccess.value = {
      id: order.id,
      total: order.total_amount,
    }
    cartStore.clearCart()
    checkoutForm.lineId = ''
    checkoutForm.phone = ''
    checkoutForm.convenienceStore = ''
    await router.push({
      path: '/order-success',
      query: {
        id: String(order.id),
        total: String(order.total_amount),
        payment: checkoutForm.paymentMethod,
      },
    })
  } catch (error) {
    console.error('建立訂單失敗:', error)
    window.alert('送出訂單失敗，請稍後再試')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="cart-page">
    <div class="page-head">
      <p class="breadcrumb">首頁 / 購物車</p>
      <h1>購物車</h1>
    </div>

    <div v-if="cartStore.items.length" class="cart-layout">
      <article class="panel cart-table-card">
        <table class="cart-table">
          <thead>
            <tr>
              <th>商品</th>
              <th>單價</th>
              <th>數量</th>
              <th>小計</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in cartStore.items" :key="item.id">
              <td class="product-column" data-label="商品">
                <div class="product-cell">
                  <img :src="item.image" :alt="item.name">
                  <div class="product-copy">
                    <span class="product-name">{{ item.name }}</span>
                    <small v-if="item.variantLabel">{{ item.variantLabel }}</small>
                    <label
                      v-if="(cartVariantOptionsMap.get(item.id)?.length ?? 0) > 0"
                      class="variant-switcher"
                    >
                      <span>修改規格</span>
                      <select
                        :value="item.variantSku"
                        @change="updateCartItemVariant(item.id, ($event.target as HTMLSelectElement).value)"
                      >
                        <option
                          v-for="option in cartVariantOptionsMap.get(item.id)"
                          :key="option.sku"
                          :value="option.sku"
                        >
                          {{ option.label }}｜庫存 {{ option.stock }}
                        </option>
                      </select>
                    </label>
                    <small
                      v-if="cartValidationMap.get(item.id)?.isInvalid"
                      class="stock-warning"
                    >
                      {{ cartValidationMap.get(item.id)?.message }}
                    </small>
                  </div>
                </div>
              </td>
              <td class="price-column" data-label="單價">NT$ {{ item.price.toFixed(2) }}</td>
              <td class="quantity-column" data-label="數量">
                <input
                  class="quantity-input"
                  :value="item.quantity"
                  type="number"
                  min="1"
                  @input="cartStore.updateQuantity(item.id, Number(($event.target as HTMLInputElement).value))"
                >
              </td>
              <td class="subtotal-column" data-label="小計">NT$ {{ (item.price * item.quantity).toFixed(2) }}</td>
              <td class="action-column">
                <button class="link-button" type="button" @click="cartStore.removeItem(item.id)">移除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </article>

      <aside class="panel summary-card">
        <h2>訂單摘要</h2>
        <div v-if="orderSuccess" class="success-box">
          <strong>訂單已建立</strong>
          <p>訂單編號 #{{ orderSuccess.id }}，金額 NT$ {{ orderSuccess.total.toFixed(2) }}</p>
        </div>
        <div class="summary-row">
          <span>商品件數</span>
          <strong>{{ cartStore.itemCount }}</strong>
        </div>
        <div class="summary-row">
          <span>商品小計</span>
          <strong>NT$ {{ cartStore.subtotal.toFixed(2) }}</strong>
        </div>
        <div class="summary-row">
          <span>運費</span>
          <strong>NT$ {{ shippingFee.toFixed(2) }}</strong>
        </div>
        <div class="summary-row total">
          <span>總計</span>
          <strong>NT$ {{ finalTotal.toFixed(2) }}</strong>
        </div>
        <div v-if="hasInvalidCartItems" class="summary-warning">
          購物車中有庫存不足的商品，請先調整數量或移除後再下單。
        </div>
        <label class="checkout-field">
          <span>Line ID</span>
          <input v-model="checkoutForm.lineId" type="text" placeholder="填寫您的 Line ID，方便客服聯繫">
        </label>
        <label class="checkout-field">
          <span>聯絡電話</span>
          <input v-model="checkoutForm.phone" type="tel" placeholder="填寫收件人電話">
        </label>
        <label class="checkout-field">
          <span>7-11 門市</span>
          <input v-model="checkoutForm.convenienceStore" type="text" placeholder="例如：臺北車站門市">
        </label>
        <label class="checkout-field">
          <span>付款方式</span>
          <div class="checkout-fixed-value">7-11 貨到付款</div>
        </label>
        <button class="primary" type="button" :disabled="submitting || hasInvalidCartItems" @click="checkout">
          {{ submitting ? '送出中...' : '送出訂單' }}
        </button>
        <button class="secondary" type="button" @click="cartStore.clearCart">清空購物車</button>
      </aside>
    </div>

    <div v-else class="panel empty-state">
      <h2>購物車目前是空的</h2>
      <p>可以先從商品目錄挑選想展示的商品，測試加入購物車與金額計算流程。</p>
      <NuxtLink class="primary-link" to="/products">去逛商品</NuxtLink>
    </div>
  </section>
</template>

<style scoped>
.cart-page {
  display: grid;
  gap: 1rem;
}

.breadcrumb,
.page-head p,
.empty-state p {
  color: var(--wp-text-muted);
}

.page-head h1 {
  margin: 0.35rem 0 0.45rem;
}

.cart-layout {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr 340px;
  align-items: start;
}

.cart-table {
  width: 100%;
  border-collapse: collapse;
}

.cart-table th,
.cart-table td {
  padding: 0.95rem 0.8rem;
  border-bottom: 1px solid var(--wp-border);
  text-align: left;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.product-copy {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
}

.variant-switcher {
  display: grid;
  gap: 0.3rem;
  margin-top: 0.2rem;
}

.variant-switcher span {
  color: var(--wp-text-muted);
  font-size: 0.78rem;
}

.variant-switcher select {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  border-radius: 0.45rem;
  border: 1px solid var(--wp-border);
  background: #ffffff;
  padding: 0.45rem 0.6rem;
  font-size: 0.82rem;
}

.product-name {
  display: -webkit-box;
  overflow: hidden;
  line-height: 1.4;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.product-copy small {
  color: var(--wp-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-copy .stock-warning {
  color: #c62828;
  white-space: normal;
}

.product-cell img {
  width: 56px;
  height: 56px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.quantity-input {
  width: 72px;
  text-align: center;
}

.summary-card,
.success-box,
.checkout-field {
  display: grid;
  gap: 0.75rem;
}

.success-box {
  padding: 0.85rem 1rem;
  border-radius: 0.5rem;
  background: rgba(70, 180, 80, 0.08);
  color: #1f6f2a;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.summary-row.total {
  padding-top: 0.85rem;
  border-top: 1px solid var(--wp-border);
  font-size: 1.05rem;
}

.summary-warning {
  color: #c62828;
  font-size: 0.9rem;
  line-height: 1.5;
}

.checkout-field span {
  font-size: 0.85rem;
  color: var(--wp-text-muted);
}

.checkout-field input,
.checkout-field textarea,
.checkout-field select {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
  padding: 0.8rem 0.9rem;
}

.checkout-fixed-value {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: var(--wp-surface-soft);
  padding: 0.8rem 0.9rem;
  color: var(--wp-heading);
}

@media (max-width: 960px) {
  .cart-layout {
    grid-template-columns: 1fr;
  }

  .cart-table-card {
    overflow-x: auto;
  }
}

@media (max-width: 640px) {
  .cart-table-card {
    overflow: visible;
  }

  .cart-table thead {
    display: none;
  }

  .cart-table,
  .cart-table tbody,
  .cart-table tr,
  .cart-table td {
    display: block;
    width: 100%;
  }

  .cart-table tr {
    padding: 0.85rem 0;
    border-bottom: 1px solid var(--wp-border);
  }

  .cart-table tbody tr:last-child {
    border-bottom: none;
  }

  .cart-table td {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.35rem 0;
    border-bottom: none;
    font-size: 0.92rem;
  }

  .cart-table td::before {
    content: attr(data-label);
    color: var(--wp-text-muted);
    font-size: 0.78rem;
    flex: 0 0 auto;
  }

  .cart-table td.product-column {
    display: block;
    padding-bottom: 0.65rem;
  }

  .cart-table td.product-column::before,
  .cart-table td.action-column::before {
    content: none;
  }

  .cart-table td.action-column {
    justify-content: flex-end;
    padding-top: 0.15rem;
  }

  .product-cell {
    align-items: flex-start;
    gap: 0.65rem;
  }

  .product-copy,
  .variant-switcher {
    width: 100%;
    min-width: 0;
  }

  .variant-switcher select {
    font-size: 0.8rem;
  }

  .product-cell img {
    width: 48px;
    height: 48px;
  }

  .product-name {
    font-size: 0.94rem;
  }

  .quantity-input {
    width: 64px;
    padding: 0.45rem 0.35rem;
  }
}
</style>
