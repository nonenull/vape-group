<script setup lang="ts">
import { computed } from 'vue'
import { fetchOrderDetail } from '~/composables/useStoreApi'
import { useStoreSeo } from '~/composables/useStoreSeo'

const tenantStore = useTenantStore()
await tenantStore.initTenant()

const route = useRoute()
const orderId = computed(() => Number(route.query.id ?? 0) || 0)
const orderDetail = orderId.value ? await fetchOrderDetail(orderId.value).catch(() => null) : null
const totalAmount = computed(() => Number(orderDetail?.totalAmount ?? route.query.total ?? 0) || 0)
const paymentMethod = computed(() => String(orderDetail?.paymentMethod ?? route.query.payment ?? ''))
const lineContactUrl = computed(() => tenantStore.platformConfig.lineContactUrl.trim())
const supportText = computed(() => tenantStore.currentTenant?.supportText?.trim() || '如需確認訂單，請聯繫 Line 客服協助處理。')
const shippingFee = computed(() => {
  if (!orderDetail) {
    return 0
  }
  const itemsTotal = orderDetail.items.reduce((sum, item) => sum + item.price * item.quantity, 0)
  return Math.max(0, orderDetail.totalAmount - itemsTotal)
})

const paymentLabel = computed(() => {
  if (paymentMethod.value === 'bank_transfer') {
    return '銀行轉帳'
  }
  if (paymentMethod.value === 'cash_on_delivery') {
    return '貨到付款'
  }
  return '未指定'
})

useStoreSeo({
  title: '訂單完成 | Vape Group 商城',
  description: '訂單已送出成功，請儘快聯繫 Line 客服確認訂單資訊。',
  canonicalPath: '/order-success',
  type: 'website',
  robots: 'noindex,nofollow',
})
</script>

<template>
  <section class="order-success-page">
    <article class="panel receipt-card">
      <p class="eyebrow">Order Receipt</p>
      <h1>訂單已送出</h1>
      <p class="lead">{{ supportText }}</p>

      <div class="receipt-grid">
        <div class="receipt-row">
          <span>訂單編號</span>
          <strong>{{ orderId ? `#${orderId}` : '已建立' }}</strong>
        </div>
        <div class="receipt-row">
          <span>訂單金額</span>
          <strong>NT$ {{ totalAmount.toFixed(2) }}</strong>
        </div>
        <div class="receipt-row">
          <span>付款方式</span>
          <strong>{{ paymentLabel }}</strong>
        </div>
        <div v-if="orderDetail" class="receipt-row">
          <span>7-11 門市</span>
          <strong>{{ orderDetail.convenienceStore || '未填寫' }}</strong>
        </div>
        <div v-if="orderDetail" class="receipt-row">
          <span>運費</span>
          <strong>NT$ {{ shippingFee.toFixed(2) }}</strong>
        </div>
      </div>

      <div v-if="orderDetail?.items.length" class="order-items-card">
        <h2>購物明細</h2>
        <div
          v-for="item in orderDetail.items"
          :key="item.id"
          class="order-item-row"
        >
          <div class="order-item-copy">
            <strong>{{ item.name }}</strong>
            <span>SKU：{{ item.variantSku || '—' }}</span>
            <span v-if="item.variantName">規格：{{ item.variantName }}</span>
          </div>
          <div class="order-item-meta">
            <span>數量 × {{ item.quantity }}</span>
            <strong>NT$ {{ (item.price * item.quantity).toFixed(2) }}</strong>
          </div>
        </div>
      </div>

      <div class="actions">
        <a
          v-if="lineContactUrl"
          class="primary-link"
          :href="lineContactUrl"
          target="_blank"
          rel="noreferrer"
        >
          聯繫客服
        </a>
        <NuxtLink class="secondary-link" to="/products">繼續逛商品</NuxtLink>
      </div>

      <p class="hint">
        請將訂單編號提供給客服，以便更快為您確認付款、庫存與出貨安排。
      </p>
    </article>
  </section>
</template>

<style scoped>
.order-success-page {
  display: grid;
  place-items: center;
  padding: 1rem 0;
}

.receipt-card {
  width: min(100%, 640px);
  display: grid;
  gap: 1rem;
}

.eyebrow,
.lead,
.hint,
.receipt-row span {
  color: var(--wp-text-muted);
}

.eyebrow {
  margin: 0;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.receipt-card h1 {
  margin: 0;
}

.lead,
.hint {
  margin: 0;
  line-height: 1.7;
}

.receipt-grid {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 0.75rem;
  background: linear-gradient(180deg, #ffffff, var(--tenant-surface, #f6f7f7));
  border: 1px solid var(--tenant-border, var(--wp-border));
}

.order-items-card {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 0.75rem;
  border: 1px solid var(--tenant-border, var(--wp-border));
  background: #fff;
}

.order-items-card h2 {
  margin: 0;
  font-size: 1rem;
}

.order-item-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--tenant-border, var(--wp-border));
}

.order-item-row:first-of-type {
  padding-top: 0;
  border-top: none;
}

.order-item-copy,
.order-item-meta {
  display: grid;
  gap: 0.2rem;
}

.order-item-copy span,
.order-item-meta span {
  color: var(--wp-text-muted);
  font-size: 0.84rem;
}

.receipt-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.primary-link,
.secondary-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 42px;
  padding: 0.75rem 1rem;
  border-radius: 999px;
  text-decoration: none;
}

.primary-link {
  background: var(--tenant-accent, var(--wp-blue));
  color: #ffffff;
}

.secondary-link {
  border: 1px solid var(--tenant-border, var(--wp-border));
  color: inherit;
}

@media (max-width: 640px) {
  .receipt-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .order-item-row {
    flex-direction: column;
  }

  .actions {
    display: grid;
  }
}
</style>
