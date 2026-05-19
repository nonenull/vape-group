<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useStoreSeo } from '~/composables/useStoreSeo'
import { createOrder } from '~/composables/useStoreApi'
import { useCheckoutStore } from '~/stores/checkout'

const router = useRouter()
const checkoutStore = useCheckoutStore()
checkoutStore.hydrate()

const shippingFee = computed(() => (checkoutStore.subtotal >= 1200 || !checkoutStore.directItem ? 0 : 90))
const finalTotal = computed(() => checkoutStore.subtotal + shippingFee.value)
const checkoutForm = reactive({
  lineId: '',
  phone: '',
  convenienceStore: '',
  shippingAddress: '',
  paymentMethod: 'cash_on_delivery',
})
const submitting = ref(false)
const orderSuccess = ref<{ id: number; total: number } | null>(null)

useStoreSeo({
  title: '直接下單 | Vape Group 商城',
  description: '立即結算目前選擇的商品並送出訂單。',
  canonicalPath: '/checkout',
  type: 'website',
  robots: 'noindex,nofollow',
})

const checkout = async () => {
  if (!checkoutStore.directItem) {
    return
  }
  if (!checkoutForm.shippingAddress.trim()) {
    window.alert('請填寫收件地址')
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
    const item = checkoutStore.directItem
    const order = await createOrder({
      items: [{
        product_id: item.productId,
        variant_name: item.variantLabel,
        variant_sku: item.variantSku,
        quantity: item.quantity,
      }],
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
    checkoutStore.clearDirectCheckout()
    checkoutForm.lineId = ''
    checkoutForm.phone = ''
    checkoutForm.convenienceStore = ''
    checkoutForm.shippingAddress = ''
  } catch (error) {
    console.error('建立直接下單訂單失敗:', error)
    window.alert('送出訂單失敗，請稍後再試')
  } finally {
    submitting.value = false
  }
}

const goBackToProduct = async () => {
  const productId = checkoutStore.directItem?.productId
  if (productId) {
    await router.push(`/products/${productId}`)
    return
  }
  await router.push('/products')
}
</script>

<template>
  <section class="checkout-page">
    <div class="page-head">
      <p class="breadcrumb">首頁 / 直接下單</p>
      <h1>直接下單</h1>
    </div>

    <div v-if="checkoutStore.directItem" class="checkout-layout">
      <article class="panel checkout-item-card">
        <h2>商品資訊</h2>
        <div class="product-cell">
          <img :src="checkoutStore.directItem.image" :alt="checkoutStore.directItem.name">
          <div class="product-copy">
            <strong>{{ checkoutStore.directItem.name }}</strong>
            <small v-if="checkoutStore.directItem.variantLabel">{{ checkoutStore.directItem.variantLabel }}</small>
            <span>SKU：{{ checkoutStore.directItem.variantSku }}</span>
          </div>
        </div>

        <div class="summary-grid">
          <div class="summary-row">
            <span>單價</span>
            <strong>NT$ {{ checkoutStore.directItem.price.toFixed(2) }}</strong>
          </div>
          <label class="checkout-field">
            <span>數量</span>
            <input
              :value="checkoutStore.directItem.quantity"
              type="number"
              min="1"
              @input="checkoutStore.updateDirectQuantity(Number(($event.target as HTMLInputElement).value))"
            >
          </label>
          <div class="summary-row total">
            <span>商品小計</span>
            <strong>NT$ {{ checkoutStore.subtotal.toFixed(2) }}</strong>
          </div>
        </div>
      </article>

      <aside class="panel summary-card">
        <h2>結算資訊</h2>
        <div v-if="orderSuccess" class="success-box">
          <strong>訂單已建立</strong>
          <p>訂單編號 #{{ orderSuccess.id }}，金額 NT$ {{ orderSuccess.total.toFixed(2) }}</p>
        </div>
        <div class="summary-row">
          <span>商品小計</span>
          <strong>NT$ {{ checkoutStore.subtotal.toFixed(2) }}</strong>
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
          <input v-model="checkoutForm.lineId" type="text" placeholder="填寫顧客 Line ID">
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
          <span>收件地址</span>
          <textarea v-model="checkoutForm.shippingAddress" rows="3" placeholder="填寫詳細收件地址" />
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
        <button class="secondary" type="button" @click="goBackToProduct">返回商品頁面</button>
      </aside>
    </div>

    <div v-else class="panel empty-state">
      <h2>目前沒有直接下單商品</h2>
      <p>請先回商品詳情頁選擇規格，再使用直接下單。</p>
      <NuxtLink class="primary-link" to="/products">去逛商品</NuxtLink>
    </div>
  </section>
</template>

<style scoped>
.checkout-page {
  display: grid;
  gap: 1rem;
}

.breadcrumb,
.page-head p,
.empty-state p,
.product-copy small,
.product-copy span {
  color: var(--wp-text-muted);
}

.page-head h1 {
  margin: 0.35rem 0 0.45rem;
}

.checkout-layout {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr 340px;
  align-items: start;
}

.checkout-item-card,
.summary-card,
.summary-grid,
.success-box,
.checkout-field {
  display: grid;
  gap: 1rem;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.product-cell img {
  width: 72px;
  height: 72px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
}

.product-copy {
  display: grid;
  gap: 0.2rem;
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

@media (max-width: 960px) {
  .checkout-layout {
    grid-template-columns: 1fr;
  }
}
</style>
