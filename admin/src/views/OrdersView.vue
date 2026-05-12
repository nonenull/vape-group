<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useAdminStore } from '@/stores/admin'

const store = useAdminStore()
const selectedOrderId = ref(store.orders[0]?.id ?? 0)
const selectedOrder = computed(() => store.orders.find((item) => item.id === selectedOrderId.value))
const selectedStatus = ref(selectedOrder.value?.status ?? '待付款')

watch(
  selectedOrder,
  (order) => {
    if (order) {
      selectedStatus.value = order.status
    }
  },
  { immediate: true },
)

function saveStatus() {
  if (selectedOrder.value) {
    store.updateOrderStatus(selectedOrder.value.id, selectedStatus.value)
  }
}
</script>

<template>
  <section class="orders-page">
    <div class="page-heading">
      <div>
        <p class="label">Order Center</p>
        <h2>訂單追蹤</h2>
        <p class="subcopy">對應 `product.md` 的訂單模組，先建立後台追蹤與狀態流轉頁面，方便之後直接接入真實 API。</p>
      </div>
    </div>

    <div class="orders-layout">
      <article class="table-card">
        <table class="order-table">
          <thead>
            <tr>
              <th>訂單編號</th>
              <th>租戶</th>
              <th>客戶</th>
              <th>金額</th>
              <th>狀態</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="order in store.orders"
              :key="order.id"
              :class="{ selected: order.id === selectedOrderId }"
              @click="selectedOrderId = order.id"
            >
              <td>{{ order.orderNo }}</td>
              <td>{{ store.getTenantName(order.tenantId) }}</td>
              <td>{{ order.customerName }}</td>
              <td>NT$ {{ order.totalAmount.toLocaleString() }}</td>
              <td>{{ order.status }}</td>
            </tr>
          </tbody>
        </table>
      </article>

      <article class="detail-card" v-if="selectedOrder">
        <div class="card-heading">
          <h3>訂單詳情</h3>
          <small>{{ selectedOrder.createdAt }}</small>
        </div>
        <div class="detail-grid">
          <div>
            <span>訂單編號</span>
            <strong>{{ selectedOrder.orderNo }}</strong>
          </div>
          <div>
            <span>租戶</span>
            <strong>{{ store.getTenantName(selectedOrder.tenantId) }}</strong>
          </div>
          <div>
            <span>客戶名稱</span>
            <strong>{{ selectedOrder.customerName }}</strong>
          </div>
          <div>
            <span>付款方式</span>
            <strong>{{ selectedOrder.paymentMethod }}</strong>
          </div>
          <div>
            <span>訂單金額</span>
            <strong>NT$ {{ selectedOrder.totalAmount.toLocaleString() }}</strong>
          </div>
        </div>
        <label class="status-control">
          <span>訂單狀態</span>
          <select v-model="selectedStatus">
            <option value="待付款">待付款</option>
            <option value="已付款">已付款</option>
            <option value="已出貨">已出貨</option>
            <option value="已完成">已完成</option>
          </select>
        </label>
        <div v-if="selectedOrder.items?.length" class="order-items">
          <h4>訂購項目</h4>
          <div v-for="(item, index) in selectedOrder.items" :key="`${item.sku}-${index}`" class="order-item-row">
            <div>
              <strong>{{ item.name }}</strong>
              <p>{{ item.variantLabel || '單一規格' }}</p>
            </div>
            <div class="order-item-side">
              <span>{{ item.sku }}</span>
              <small>{{ item.quantity }} 件 · NT$ {{ item.price.toLocaleString() }}</small>
            </div>
          </div>
        </div>
        <div class="actions">
          <button class="primary" type="button" @click="saveStatus">儲存狀態</button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.orders-page {
  display: grid;
  gap: 1rem;
}

.label {
  color: var(--wp-blue);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: 0.75rem;
}

.page-heading h2 {
  margin: 0.35rem 0 0.45rem;
}

.subcopy {
  color: var(--wp-text-muted);
  max-width: 72ch;
}

.orders-layout {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr 360px;
}

.table-card,
.detail-card {
  background: #fff;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
}

.order-table {
  width: 100%;
  border-collapse: collapse;
}

.order-table th,
.order-table td {
  padding: 0.95rem 1rem;
  border-bottom: 1px solid var(--wp-border);
  text-align: left;
}

.order-table tbody tr {
  cursor: pointer;
}

.order-table tbody tr.selected {
  background: #f0f6fc;
}

.detail-card {
  padding: 1rem 1.25rem;
}

.card-heading {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.card-heading small,
.detail-grid span {
  color: var(--wp-text-muted);
}

.detail-grid {
  display: grid;
  gap: 0.9rem;
}

.detail-grid strong {
  display: block;
  margin-top: 0.15rem;
}

.status-control {
  display: grid;
  gap: 0.4rem;
  margin-top: 1rem;
}

.order-items {
  display: grid;
  gap: 0.75rem;
  margin-top: 1rem;
}

.order-item-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 0.95rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
}

.order-item-row p,
.order-item-side span,
.order-item-side small {
  color: var(--wp-text-muted);
}

.order-item-side {
  text-align: right;
}

.status-control span {
  font-weight: 600;
}

select {
  min-height: 2.5rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 0.375rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.primary {
  min-height: 2.5rem;
  padding: 0.65rem 1rem;
  border: 1px solid var(--wp-blue);
  border-radius: 0.375rem;
  background: var(--wp-blue);
  color: #fff;
  font-weight: 600;
}

@media (max-width: 980px) {
  .orders-layout {
    grid-template-columns: 1fr;
  }

  .order-table {
    display: block;
    overflow-x: auto;
  }
}
</style>
