<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useSeo } from '@/composables/useSeo'
import { orderAPI } from '@/api/orders'
import { useCartStore } from '@/stores/cart'

const cartStore = useCartStore()
const shippingFee = computed(() => (cartStore.subtotal >= 1200 || cartStore.itemCount === 0 ? 0 : 90))
const finalTotal = computed(() => cartStore.subtotal + shippingFee.value)
const checkoutForm = reactive({
  lineId: '',
  phone: '',
  convenienceStore: '',
  shippingAddress: '',
  paymentMethod: 'cash_on_delivery',
})
const submitting = ref(false)
const orderSuccess = ref<{ id: number; total: number } | null>(null)

useSeo(computed(() => ({
  title: `購物車${cartStore.itemCount ? ` (${cartStore.itemCount})` : ''} | Vape Group 商城`,
  description: '檢視購物車商品、填寫顧客資料並送出訂單。',
  canonicalPath: '/cart',
  type: 'website' as const,
  robots: 'noindex,nofollow',
})))

const checkout = async () => {
  if (!cartStore.items.length) {
    return
  }
  if (!checkoutForm.shippingAddress.trim()) {
    alert('請填寫收件地址')
    return
  }
  if (!checkoutForm.lineId.trim()) {
    alert('請填寫 Line ID')
    return
  }
  if (!checkoutForm.phone.trim()) {
    alert('請填寫聯絡電話')
    return
  }
  if (!checkoutForm.convenienceStore.trim()) {
    alert('請填寫 7-11 門市')
    return
  }

  submitting.value = true
  try {
    const order = await orderAPI.createOrder({
      items: cartStore.items.map((item) => ({
        product_id: item.productId,
        variant_name: item.variantLabel,
        variant_sku: item.variantSku,
        quantity: item.quantity,
      })),
      line_id: checkoutForm.lineId.trim(),
      phone: checkoutForm.phone.trim(),
      convenience_store: checkoutForm.convenienceStore.trim(),
      shipping_address: checkoutForm.shippingAddress.trim(),
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
    checkoutForm.shippingAddress = ''
  } catch (error) {
    console.error('建立訂單失敗:', error)
    alert('送出訂單失敗，請稍後再試')
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
      <p>現在已具備前台購物車管理流程，可調整數量、刪除商品並預覽訂單金額。</p>
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
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in cartStore.items" :key="item.id">
              <td>
                <div class="product-cell">
                  <img :src="item.image" :alt="item.name" />
                  <div class="product-copy">
                    <span>{{ item.name }}</span>
                    <small v-if="item.variantLabel">{{ item.variantLabel }}</small>
                  </div>
                </div>
              </td>
              <td>NT$ {{ item.price.toFixed(2) }}</td>
              <td>
                <input
                  :value="item.quantity"
                  type="number"
                  min="1"
                  @input="cartStore.updateQuantity(item.id, Number(($event.target as HTMLInputElement).value))"
                />
              </td>
              <td>NT$ {{ (item.price * item.quantity).toFixed(2) }}</td>
              <td>
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
        <label class="checkout-field">
          <span>Line ID</span>
          <input v-model="checkoutForm.lineId" type="text" placeholder="填寫顧客 Line ID" />
        </label>
        <label class="checkout-field">
          <span>聯絡電話</span>
          <input v-model="checkoutForm.phone" type="tel" placeholder="填寫收件人電話" />
        </label>
        <label class="checkout-field">
          <span>7-11 門市</span>
          <input v-model="checkoutForm.convenienceStore" type="text" placeholder="例如：臺北車站門市" />
        </label>
        <label class="checkout-field">
          <span>收件地址</span>
          <textarea v-model="checkoutForm.shippingAddress" rows="3" placeholder="填寫詳細收件地址"></textarea>
        </label>
        <label class="checkout-field">
          <span>付款方式</span>
          <select v-model="checkoutForm.paymentMethod">
            <option value="cash_on_delivery">貨到付款</option>
            <option value="bank_transfer">銀行轉帳</option>
          </select>
        </label>
        <button class="primary" type="button" :disabled="submitting" @click="checkout">
          {{ submitting ? '送出中...' : '送出訂單' }}
        </button>
        <button class="secondary" type="button" @click="cartStore.clearCart">清空購物車</button>
      </aside>
    </div>

    <div v-else class="panel empty-state">
      <h2>購物車目前是空的</h2>
      <p>可以先從商品目錄挑選想展示的商品，測試加入購物車與金額計算流程。</p>
      <RouterLink class="primary-link" to="/products">去逛商品</RouterLink>
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

.panel {
  background: var(--wp-surface);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1.25rem;
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
}

.product-cell span {
  min-width: 0;
  line-height: 1.4;
}

.product-copy {
  display: grid;
  gap: 0.2rem;
}

.product-copy small {
  color: var(--wp-text-muted);
}

.product-cell img {
  width: 56px;
  height: 56px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.cart-table input {
  width: 88px;
  min-height: 2.4rem;
  padding: 0.55rem 0.65rem;
  border-radius: 0.375rem;
  border: 1px solid var(--wp-border-strong);
}

.summary-card h2 {
  margin-bottom: 1rem;
}

.summary-card {
  min-width: 0;
}

.success-box {
  margin-bottom: 1rem;
  padding: 0.9rem 1rem;
  border-radius: 0.5rem;
  background: #edf7ed;
  border: 1px solid #b7dfb9;
  color: #195c1c;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--wp-border);
}

.checkout-field {
  display: grid;
  gap: 0.4rem;
  margin-top: 1rem;
}

.checkout-field span {
  font-weight: 600;
}

.checkout-field input,
.checkout-field textarea,
.checkout-field select {
  width: 100%;
  min-height: 2.7rem;
  padding: 0.7rem 0.8rem;
  border-radius: 0.375rem;
  border: 1px solid var(--wp-border-strong);
  background: #fff;
}

.summary-row.total {
  font-size: 1.05rem;
}

.primary,
.secondary,
.primary-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-height: 2.7rem;
  margin-top: 1rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid var(--wp-border-strong);
}

.primary,
.primary-link {
  background: var(--tenant-accent, var(--wp-blue));
  color: #fff;
  border-color: var(--tenant-accent, var(--wp-blue));
}

.secondary {
  background: #fff;
  color: var(--wp-heading);
}

.link-button {
  border: 0;
  padding: 0;
  background: transparent;
  color: var(--tenant-accent, var(--wp-blue));
  font-weight: 600;
  cursor: pointer;
}

.empty-state {
  text-align: center;
  padding: 2.5rem;
}

@media (max-width: 960px) {
  .cart-layout {
    grid-template-columns: 1fr;
  }

  .cart-table-card,
  .summary-card {
    min-width: 0;
  }

  .cart-table {
    display: block;
    overflow-x: auto;
  }
}

@media (max-width: 640px) {
  .panel {
    padding: 1rem;
  }

  .cart-table th,
  .cart-table td {
    padding: 0.75rem 0.45rem;
  }

  .cart-table th:first-child,
  .cart-table td:first-child {
    min-width: 180px;
  }

  .product-cell {
    gap: 0.6rem;
    align-items: flex-start;
  }

  .product-cell img {
    width: 48px;
    height: 48px;
    flex: 0 0 48px;
  }

  .product-cell span {
    min-width: 120px;
    font-size: 0.9rem;
    word-break: break-word;
  }

  .summary-card {
    width: 100%;
    overflow-x: hidden;
  }

  .summary-row {
    gap: 0.75rem;
    padding: 0.65rem 0;
    font-size: 0.92rem;
  }

  .checkout-field input,
  .checkout-field textarea,
  .checkout-field select,
  .primary,
  .secondary,
  .primary-link {
    min-width: 0;
  }
}
</style>
