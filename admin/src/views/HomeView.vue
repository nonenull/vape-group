<script setup lang="ts">
import { computed } from 'vue'
import { useAdminStore } from '@/stores/admin'

const store = useAdminStore()

const paidOrders = computed(() => store.orders.filter((order) => order.status === '已付款').length)
const lowStockProducts = computed(() =>
  store.products.filter((product) => product.baseStockQuantity <= 20),
)
const grossRevenue = computed(() =>
  store.orders.reduce((sum, order) => {
    if (order.status === '已付款' || order.status === '已出貨' || order.status === '已完成') {
      return sum + order.totalAmount
    }
    return sum
  }, 0),
)
</script>

<template>
  <section class="dashboard-page">
    <div class="welcome-card panel">
      <div>
        <p class="eyebrow">Dashboard</p>
        <h2>多租戶營運總覽</h2>
        <p>依照產品文件的方向，這裡集中顯示租戶、商品、覆寫與訂單狀態，方便管理員快速判斷哪個站點需要調整。</p>
      </div>
      <div class="status-callout">
        <span>租戶狀態</span>
        <strong>{{ store.activeTenants.length }}/{{ store.tenants.length }}</strong>
        <small>啟用中的站點可正常接收前台流量</small>
      </div>
    </div>

    <div class="stats-grid">
      <article class="panel stat-card">
        <span>全域商品</span>
        <strong>{{ store.products.length }}</strong>
        <p>集中維護 SKU、基礎價格與庫存</p>
      </article>
      <article class="panel stat-card">
        <span>租戶覆寫</span>
        <strong>{{ store.visibleOverrides.length }}</strong>
        <p>可見的租戶客製資料組合</p>
      </article>
      <article class="panel stat-card">
        <span>已付款訂單</span>
        <strong>{{ paidOrders }}</strong>
        <p>目前付款已確認，可進入出貨流程</p>
      </article>
      <article class="panel stat-card">
        <span>營收快照</span>
        <strong>NT$ {{ grossRevenue.toLocaleString() }}</strong>
        <p>已付款、已出貨與已完成訂單加總</p>
      </article>
    </div>

    <div class="dashboard-grid">
      <article class="panel">
        <div class="panel-heading">
          <h3>低庫存提醒</h3>
          <RouterLink to="/products">前往商品管理</RouterLink>
        </div>
        <div class="warning-list">
          <div v-for="product in lowStockProducts" :key="product.id" class="warning-item">
            <strong>{{ product.baseName }}</strong>
            <span>{{ product.sku }}</span>
            <small>剩餘 {{ product.baseStockQuantity }} 件</small>
          </div>
        </div>
      </article>

      <article class="panel">
        <div class="panel-heading">
          <h3>最近訂單</h3>
          <RouterLink to="/orders">檢視全部</RouterLink>
        </div>
        <div class="order-list">
          <div v-for="order in store.orders.slice(0, 3)" :key="order.id" class="order-item">
            <div>
              <strong>{{ order.orderNo }}</strong>
              <p>{{ store.getTenantName(order.tenantId) }} · {{ order.customerName }}</p>
            </div>
            <div class="order-side">
              <span>{{ order.status }}</span>
              <small>NT$ {{ order.totalAmount.toLocaleString() }}</small>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.dashboard-page {
  display: grid;
  gap: 1rem;
}

.panel {
  background: var(--wp-surface);
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1.25rem;
}

.welcome-card {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.eyebrow,
.stat-card span {
  color: var(--wp-blue);
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.welcome-card h2,
.panel-heading h3 {
  margin: 0.4rem 0 0.5rem;
}

.welcome-card p:last-child,
.stat-card p,
.warning-item span,
.warning-item small,
.order-item p,
.order-side small {
  color: var(--wp-text-muted);
}

.status-callout {
  min-width: 240px;
  padding: 1rem;
  border-radius: 0.5rem;
  background: linear-gradient(180deg, #f0f6fc, #fff);
  border: 1px solid var(--wp-border);
}

.status-callout span,
.status-callout small {
  display: block;
  color: var(--wp-text-muted);
}

.status-callout strong {
  display: block;
  margin: 0.45rem 0;
  color: var(--wp-green);
  font-size: 1.6rem;
}

.stats-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(4, 1fr);
}

.dashboard-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1.1fr 0.9fr;
}

.stat-card strong {
  display: block;
  margin: 0.4rem 0;
  font-size: 1.5rem;
}

.panel-heading {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.panel-heading a {
  color: var(--wp-blue);
  font-size: 0.8125rem;
  font-weight: 600;
}

.warning-list,
.order-list {
  display: grid;
  gap: 0.75rem;
}

.warning-item,
.order-item {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
}

.warning-item {
  align-items: center;
}

.order-side {
  text-align: right;
}

@media (max-width: 960px) {
  .welcome-card,
  .stats-grid,
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .welcome-card {
    flex-direction: column;
  }
}
</style>
