<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAdminStore } from '@/stores/admin'
import type { BrandRecord } from '@/data/adminMock'

const store = useAdminStore()

const selectedTenantId = ref(0)
const selectedBrandId = ref(0)
const isCreating = ref(false)

const emptyBrand = (): BrandRecord => ({
  id: 0,
  tenantId: selectedTenantId.value,
  name: '',
  logoUrl: '',
  description: '',
})

const brandForm = ref<BrandRecord>(emptyBrand())
const tenantBrands = computed(() => store.getBrandsByTenant(selectedTenantId.value))

watch(selectedTenantId, async (tenantId) => {
  if (!tenantId) return
  await store.fetchBrands(tenantId)
  const fallback = store.getBrandsByTenant(tenantId)[0]
  selectedBrandId.value = fallback?.id ?? 0
  brandForm.value = fallback ? { ...fallback } : emptyBrand()
}, { immediate: true })

watch(selectedBrandId, (brandId) => {
  if (isCreating.value) return
  const brand = tenantBrands.value.find((item) => item.id === brandId)
  if (brand) {
    brandForm.value = { ...brand }
  }
})

function startCreateBrand() {
  isCreating.value = true
  selectedBrandId.value = 0
  brandForm.value = emptyBrand()
}

function selectBrand(brandId: number) {
  isCreating.value = false
  selectedBrandId.value = brandId
}

async function saveBrand() {
  if (!selectedTenantId.value) {
    alert('請先選擇租戶')
    return
  }
  if (!brandForm.value.name.trim()) {
    alert('請先填寫品牌名稱')
    return
  }

  try {
    if (isCreating.value || brandForm.value.id === 0) {
      const created = await store.createBrand(selectedTenantId.value, {
        name: brandForm.value.name,
        logoUrl: brandForm.value.logoUrl,
        description: brandForm.value.description,
      })
      isCreating.value = false
      selectedBrandId.value = created.id
      brandForm.value = { ...created }
      alert('品牌已成功建立')
      return
    }

    await store.updateBrand({ ...brandForm.value, tenantId: selectedTenantId.value })
    alert('品牌已成功儲存')
  } catch (error) {
    console.error('保存品牌失敗:', error)
    alert('保存品牌失敗: ' + (error as Error).message)
  }
}

async function removeBrand() {
  const brand = tenantBrands.value.find((item) => item.id === selectedBrandId.value)
  if (!brand) return
  if (!confirm('確定要刪除此品牌嗎？')) return

  try {
    await store.deleteBrand(selectedTenantId.value, brand.id)
    const fallback = store.getBrandsByTenant(selectedTenantId.value)[0]
    selectedBrandId.value = fallback?.id ?? 0
    brandForm.value = fallback ? { ...fallback } : emptyBrand()
    isCreating.value = false
    alert('品牌已成功刪除')
  } catch (error) {
    console.error('刪除品牌失敗:', error)
    alert('刪除品牌失敗: ' + (error as Error).message)
  }
}

onMounted(() => {
  if (!store.tenants.length) {
    store.fetchTenants()
  }
  if (!selectedTenantId.value && store.tenants[0]) {
    selectedTenantId.value = store.tenants[0].id
  }
})
</script>

<template>
  <section class="brands-page">
    <div class="page-heading">
      <div>
        <p class="label">Brand Center</p>
        <h2>品牌管理</h2>
        <p class="subcopy">依租戶維護品牌資料，包含 Logo 與說明，供商品建立時快速選用。</p>
      </div>
      <button class="primary" type="button" @click="startCreateBrand">新增品牌</button>
    </div>

    <div class="tenant-switch panel">
      <label>
        <span>管理租戶</span>
        <select v-model.number="selectedTenantId">
          <option v-for="tenant in store.tenants" :key="tenant.id" :value="tenant.id">{{ tenant.name }}</option>
        </select>
      </label>
    </div>

    <div class="workspace-grid">
      <article class="panel list-card">
        <h3>品牌清單</h3>
        <div class="item-list">
          <button
            v-for="brand in tenantBrands"
            :key="brand.id"
            type="button"
            class="item-row"
            :class="{ selected: brand.id === selectedBrandId && !isCreating }"
            @click="selectBrand(brand.id)"
          >
            <div>
              <strong>{{ brand.name }}</strong>
              <p>{{ brand.description || '尚未填寫品牌描述' }}</p>
            </div>
          </button>
        </div>
      </article>

      <article class="panel editor-card">
        <div class="card-heading">
          <h3>{{ isCreating ? '新增品牌' : '品牌設定' }}</h3>
          <small v-if="!isCreating">ID {{ brandForm.id }}</small>
        </div>

        <div class="brand-preview" v-if="brandForm.logoUrl">
          <img :src="brandForm.logoUrl" :alt="brandForm.name || 'brand logo'" />
        </div>

        <div class="form-grid">
          <label class="full">
            <span>品牌名稱</span>
            <input v-model="brandForm.name" />
          </label>
          <label class="full">
            <span>Logo URL</span>
            <input v-model="brandForm.logoUrl" />
          </label>
          <label class="full">
            <span>品牌描述</span>
            <textarea v-model="brandForm.description" rows="5"></textarea>
          </label>
        </div>

        <div class="actions">
          <button v-if="!isCreating" class="danger" type="button" @click="removeBrand">刪除品牌</button>
          <button class="primary" type="button" @click="saveBrand">{{ isCreating ? '建立品牌' : '儲存品牌' }}</button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.brands-page { display: grid; gap: 1rem; }
.page-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; }
.label { color: var(--wp-blue); font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; font-size: 0.75rem; }
.subcopy { color: var(--wp-text-muted); max-width: 72ch; }
.panel { background: #fff; border: 1px solid var(--wp-border); border-radius: 0.5rem; box-shadow: var(--wp-shadow); padding: 1rem 1.25rem; }
.tenant-switch label, .form-grid label { display: grid; gap: 0.4rem; }
.workspace-grid { display: grid; gap: 1rem; grid-template-columns: 0.9fr 1.1fr; }
.item-list { display: grid; gap: 0.75rem; }
.item-row { padding: 0.9rem 1rem; border: 1px solid var(--wp-border); border-radius: 0.5rem; background: var(--wp-surface-soft); text-align: left; }
.item-row.selected { border-color: var(--wp-blue); background: #f0f6fc; }
.item-row p { color: var(--wp-text-muted); }
.card-heading { display: flex; justify-content: space-between; gap: 1rem; align-items: center; margin-bottom: 1rem; }
.card-heading small { color: var(--wp-text-muted); }
.brand-preview { margin-bottom: 1rem; padding: 1rem; border: 1px solid var(--wp-border); border-radius: 0.75rem; background: var(--wp-surface-soft); }
.brand-preview img { max-height: 120px; object-fit: contain; }
.form-grid { display: grid; gap: 0.875rem; grid-template-columns: repeat(2, 1fr); }
.full { grid-column: 1 / -1; }
input, select, textarea { width: 100%; min-height: 2.5rem; padding: 0.65rem 0.75rem; border: 1px solid var(--wp-border-strong); border-radius: 0.375rem; background: #fff; }
textarea { resize: vertical; }
.actions { display: flex; justify-content: space-between; gap: 0.75rem; margin-top: 1rem; }
.primary, .danger { min-height: 2.5rem; padding: 0.65rem 1rem; border-radius: 0.375rem; font-weight: 600; border: 1px solid transparent; }
.primary { background: var(--wp-blue); color: #fff; border-color: var(--wp-blue); }
.danger { background: #fff; color: var(--wp-red); border-color: rgba(214, 54, 56, 0.3); }
@media (max-width: 900px) { .workspace-grid, .form-grid { grid-template-columns: 1fr; } }
</style>
