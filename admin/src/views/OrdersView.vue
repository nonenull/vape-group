<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useAdminStore } from '@/stores/admin'

const store = useAdminStore()
const selectedOrderId = ref(store.orders[0]?.id ?? 0)
const selectedOrder = computed(() => store.orders.find((item) => item.id === selectedOrderId.value))
const selectedStatus = ref(selectedOrder.value?.status ?? '已下单')

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
    ElMessage.success('訂單狀態已更新')
  }
}

onMounted(async () => {
  await store.fetchOrders()
  if (!store.tenants.length) {
    await store.fetchTenants()
  }
  selectedOrderId.value = store.orders[0]?.id ?? 0
})
</script>

<template>
  <section class="orders-page">
    <div class="page-heading">
      <div>
        <p class="label">Order Center</p>
        <h2>訂單追蹤</h2>
        <p class="subcopy">對應訂單模組，集中查看訂單狀態、客戶資訊與商品明細。</p>
      </div>
    </div>

    <div class="orders-layout">
      <el-card>
        <template #header>
          <div class="card-header">
            <span class="title">訂單列表</span>
            <small>{{ store.orders.length }} 筆訂單</small>
          </div>
        </template>

        <el-table
          :data="store.orders"
          stripe
          highlight-current-row
          @current-change="(row: typeof store.orders[number] | undefined) => selectedOrderId = row?.id ?? 0"
        >
          <el-table-column prop="orderNo" label="訂單編號" min-width="150" />
          <el-table-column label="租戶" min-width="120">
            <template #default="{ row }">
              {{ store.getTenantName(row.tenantId) }}
            </template>
          </el-table-column>
          <el-table-column prop="customerName" label="客戶" min-width="120" />
          <el-table-column label="金額" min-width="120">
            <template #default="{ row }">
              NT$ {{ row.totalAmount.toLocaleString() }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="狀態" width="110" />
        </el-table>
      </el-card>

      <el-card v-if="selectedOrder" class="detail-card">
        <template #header>
          <div class="card-header">
            <span class="title">訂單詳情</span>
            <small>{{ selectedOrder.createdAt }}</small>
          </div>
        </template>

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

        <div class="status-panel">
          <span class="status-label">訂單狀態</span>
          <el-select v-model="selectedStatus" style="width: 100%">
            <el-option value="已下单" label="已下单" />
            <el-option value="已出貨" label="已出貨" />
            <el-option value="已完成" label="已完成" />
          </el-select>
        </div>

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
          <el-button type="primary" @click="saveStatus">儲存狀態</el-button>
        </div>
      </el-card>
    </div>
  </section>
</template>

<style scoped>
.orders-page {
  display: grid;
  gap: 1rem;
}

.page-heading h2 {
  margin: 0.35rem 0 0.45rem;
}

.label {
  color: var(--wp-blue);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: 0.75rem;
}

.subcopy,
.card-header small,
.detail-grid span,
.order-item-row p,
.order-item-side span,
.order-item-side small {
  color: #909399;
}

.orders-layout {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr 360px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.card-header .title {
  font-size: 1rem;
  font-weight: 700;
}

.detail-card :deep(.el-card__body) {
  display: grid;
  gap: 1rem;
}

.detail-grid {
  display: grid;
  gap: 0.9rem;
}

.detail-grid strong {
  display: block;
  margin-top: 0.15rem;
}

.status-panel {
  display: grid;
  gap: 0.45rem;
}

.status-label {
  font-weight: 600;
}

.order-items {
  display: grid;
  gap: 0.75rem;
}

.order-item-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 0.95rem;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #f5f7fa;
}

.order-item-side {
  text-align: right;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 980px) {
  .orders-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .card-header,
  .order-item-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .order-item-side {
    text-align: left;
  }
}
</style>
