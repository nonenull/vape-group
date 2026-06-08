<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useStoreSeo } from '~/composables/useStoreSeo'
import { createOrder } from '~/composables/useStoreApi'
import { useCheckoutStore } from '~/stores/checkout'
import { useTenantStore } from '~/stores/tenant'

const router = useRouter()
const checkoutStore = useCheckoutStore()
const tenantStore = useTenantStore()
const DIRECT_CHECKOUT_FORM_STORAGE_KEY = 'vape_group_direct_checkout_form'
checkoutStore.hydrate()

await tenantStore.initTenant()

const shippingFee = computed(() => {
  const threshold = tenantStore.platformConfig.freeShippingThreshold || 0
  const fee = tenantStore.platformConfig.shippingFee || 0
  return checkoutStore.subtotal >= threshold || !checkoutStore.directItem ? 0 : fee
})
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

const persistCheckoutForm = () => {
  if (!import.meta.client) {
    return
  }

  sessionStorage.setItem(DIRECT_CHECKOUT_FORM_STORAGE_KEY, JSON.stringify({
    lineId: checkoutForm.lineId,
    phone: checkoutForm.phone,
    convenienceStore: checkoutForm.convenienceStore,
    shippingAddress: checkoutForm.shippingAddress,
    paymentMethod: checkoutForm.paymentMethod,
  }))
}

const hydrateCheckoutForm = () => {
  if (!import.meta.client) {
    return
  }

  const raw = sessionStorage.getItem(DIRECT_CHECKOUT_FORM_STORAGE_KEY)
  if (!raw) {
    return
  }

  try {
    const parsed = JSON.parse(raw)
    checkoutForm.lineId = typeof parsed.lineId === 'string' ? parsed.lineId : ''
    checkoutForm.phone = typeof parsed.phone === 'string' ? parsed.phone : ''
    checkoutForm.convenienceStore = typeof parsed.convenienceStore === 'string' ? parsed.convenienceStore : ''
    checkoutForm.shippingAddress = typeof parsed.shippingAddress === 'string' ? parsed.shippingAddress : ''
    checkoutForm.paymentMethod = typeof parsed.paymentMethod === 'string' ? parsed.paymentMethod : 'cash_on_delivery'
  } catch {
    sessionStorage.removeItem(DIRECT_CHECKOUT_FORM_STORAGE_KEY)
  }
}

const clearCheckoutFormPersistence = () => {
  if (!import.meta.client) {
    return
  }

  sessionStorage.removeItem(DIRECT_CHECKOUT_FORM_STORAGE_KEY)
}

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
    const paymentMethod = checkoutForm.paymentMethod
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
      payment_method: paymentMethod,
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
    clearCheckoutFormPersistence()
    await router.push({
      path: '/order-success',
      query: {
        id: String(order.id),
        total: String(order.total_amount),
        payment: paymentMethod,
      },
    })
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

if (import.meta.client) {
  hydrateCheckoutForm()
}

watch(
  [
    () => checkoutForm.lineId,
    () => checkoutForm.phone,
    () => checkoutForm.convenienceStore,
    () => checkoutForm.shippingAddress,
    () => checkoutForm.paymentMethod,
  ],
  () => {
    persistCheckoutForm()
  },
)
</script>

<template>
  <section class="checkout-page">
    <div class="page-head">
      <p class="breadcrumb">首頁 / 直接下單</p>
      <h1>直接下單</h1>
    </div>

    <div v-if="checkoutStore.directItem" class="checkout-layout">
      <div class="checkout-main-column">
        <article class="panel checkout-item-card">
          <div class="section-head">
            <div>
              <h2>商品資訊</h2>
              <p>確認這次要直接下單的商品、規格與數量。</p>
            </div>
          </div>
          <div class="checkout-item-overview">
            <div class="product-cell product-cell-featured">
              <img :src="checkoutStore.directItem.image" :alt="checkoutStore.directItem.name">
              <div class="product-copy">
                <strong>{{ checkoutStore.directItem.name }}</strong>
                <small v-if="checkoutStore.directItem.variantLabel">{{ checkoutStore.directItem.variantLabel }}</small>
                <span>SKU：{{ checkoutStore.directItem.variantSku }}</span>
              </div>
            </div>

            <div class="checkout-metric-grid">
              <div class="checkout-metric-item">
                <span>單價</span>
                <strong>NT$ {{ checkoutStore.directItem.price.toFixed(2) }}</strong>
              </div>
              <label class="checkout-metric-item checkout-metric-quantity">
                <span>數量</span>
                <input
                  class="quantity-input"
                  :value="checkoutStore.directItem.quantity"
                  type="number"
                  min="1"
                  @input="checkoutStore.updateDirectQuantity(Number(($event.target as HTMLInputElement).value))"
                >
              </label>
              <div class="checkout-metric-item checkout-metric-total">
                <span>商品小計</span>
                <strong>NT$ {{ checkoutStore.subtotal.toFixed(2) }}</strong>
              </div>
            </div>
          </div>
        </article>

        <article class="panel customer-info-card">
          <div class="section-head">
            <div>
              <h2>客戶資訊</h2>
              <p>填寫聯絡資料與收件資訊，方便客服確認與安排出貨。</p>
            </div>
          </div>
          <div class="customer-form-grid">
            <label class="checkout-field">
              <span>Line ID</span>
              <input v-model="checkoutForm.lineId" type="text" placeholder="填寫顧客 Line ID">
            </label>
            <label class="checkout-field">
              <span>聯絡電話</span>
              <input v-model="checkoutForm.phone" type="tel" placeholder="填寫收件人電話">
            </label>
            <label class="checkout-field customer-form-full">
              <span>7-11 門市</span>
              <input v-model="checkoutForm.convenienceStore" type="text" placeholder="填寫 7-11 門市名稱">
            </label>
            <label class="checkout-field customer-form-full">
              <span>收件地址</span>
              <textarea v-model="checkoutForm.shippingAddress" rows="3" placeholder="填寫詳細收件地址" />
            </label>
            <label class="checkout-field customer-form-full">
              <span>付款方式</span>
              <div class="checkout-fixed-value">7-11 貨到付款</div>
            </label>
          </div>
        </article>
      </div>

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
  display: flex;
  gap: 1rem;
  align-items: start;
}

.checkout-main-column {
  display: grid;
  gap: 1rem;
  flex: 1 1 auto;
  min-width: 0;
}

.checkout-item-card,
.success-box,
.checkout-field {
  display: grid;
  gap: 1rem;
}

.customer-info-card {
  display: grid;
  gap: 1rem;
}

.customer-form-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.customer-form-full {
  grid-column: 1 / -1;
}

.summary-card {
  display: grid;
  gap: 0.75rem;
  flex: 0 0 340px;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.product-cell-featured {
  align-items: flex-start;
  padding: 0.2rem 0 0.1rem;
}

.product-cell img {
  width: 84px;
  height: 84px;
  object-fit: cover;
  border-radius: 0.75rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.product-copy {
  display: grid;
  gap: 0.28rem;
}

.product-copy strong {
  color: var(--wp-heading);
  font-size: 1rem;
  line-height: 1.45;
}

.checkout-item-overview {
  display: grid;
  gap: 1rem;
}

.checkout-metric-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checkout-metric-item {
  display: grid;
  gap: 0.45rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--wp-surface-soft) 55%, #fff);
}

.checkout-metric-item span {
  color: var(--wp-text-muted);
  font-size: 0.82rem;
}

.checkout-metric-item strong {
  color: var(--wp-heading);
  font-size: 1.02rem;
  line-height: 1.35;
}

.checkout-metric-total {
  background: color-mix(in srgb, var(--tenant-accent, var(--wp-blue)) 9%, #fff);
}

.checkout-metric-total strong {
  font-size: 1.12rem;
}

.checkout-metric-quantity input {
  width: 100%;
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: start;
}

.section-head h2,
.section-head p {
  margin: 0;
}

.section-head p {
  color: var(--wp-text-muted);
  line-height: 1.6;
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

.checkout-fixed-value {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: var(--wp-surface-soft);
  color: var(--wp-text);
  padding: 0.8rem 0.9rem;
}

@media (min-width: 961px) {
  .checkout-item-card {
    gap: 1.2rem;
  }

  .checkout-item-overview {
    grid-template-columns: minmax(0, 1.15fr) minmax(360px, 0.95fr);
    align-items: stretch;
  }

  .product-cell-featured {
    gap: 1rem;
    min-height: 100%;
    padding: 1rem;
    border: 1px solid var(--wp-border);
    border-radius: 0.95rem;
    background: linear-gradient(180deg, #fbfdff 0%, #f2f6fa 100%);
  }

  .product-cell img {
    width: 112px;
    height: 112px;
    border-radius: 0.95rem;
  }

  .product-copy {
    align-content: center;
    min-width: 0;
  }

  .product-copy strong {
    font-size: 1.08rem;
  }

  .checkout-metric-grid {
    height: 100%;
  }

  .checkout-metric-item {
    align-content: space-between;
    min-height: 100%;
    padding: 1rem 1.05rem;
  }

  .customer-info-card {
    gap: 1.1rem;
  }

  .customer-form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1rem;
  }

  .summary-card {
    gap: 0.9rem;
    position: sticky;
    top: 1rem;
  }
}

@media (max-width: 960px) {
  .checkout-layout {
    flex-direction: column;
  }

  .checkout-metric-grid {
    grid-template-columns: 1fr;
  }

  .customer-form-grid {
    grid-template-columns: 1fr;
  }

  .customer-form-full {
    grid-column: auto;
  }
}

@media (max-width: 640px) {
  .checkout-item-card,
  .customer-info-card {
    gap: 0.9rem;
  }

  .product-cell-featured {
    gap: 0.8rem;
  }

  .product-cell img {
    width: 72px;
    height: 72px;
  }

  .checkout-metric-item {
    padding: 0.85rem 0.9rem;
  }

  .summary-card {
    gap: 1rem;
    padding: 1rem;
    border-radius: 0.9rem;
  }

  .summary-card h2 {
    margin: 0;
    font-size: 1.1rem;
  }

  .summary-row {
    align-items: flex-start;
    gap: 0.85rem;
    padding: 0.15rem 0;
  }

  .summary-row span {
    font-size: 0.88rem;
    line-height: 1.45;
    color: var(--wp-text-muted);
  }

  .summary-row strong {
    font-size: 0.98rem;
    line-height: 1.35;
    text-align: right;
  }

  .summary-row.total {
    margin-top: 0.15rem;
    padding-top: 1rem;
    font-size: 1.08rem;
  }

  .summary-row.total span {
    font-size: 0.92rem;
  }

  .summary-row.total strong {
    font-size: 1.2rem;
  }

  .success-box {
    gap: 0.45rem;
    padding: 0.95rem 1rem;
    border-radius: 0.8rem;
  }

  .summary-card .primary,
  .summary-card .secondary {
    min-height: 46px;
    padding: 0.85rem 1rem;
    font-size: 0.95rem;
  }
}
</style>
