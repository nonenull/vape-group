<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAdminStore } from '@/stores/admin'
import type { CategoryRecord } from '@/data/adminMock'

const store = useAdminStore()

const selectedTenantId = ref(0)
const selectedCategoryId = ref(0)
const isCreating = ref(false)

const emptyCategory = (): CategoryRecord => ({
  id: 0,
  tenantId: selectedTenantId.value,
  name: '',
  parentId: null,
  sortOrder: 0,
})

const categoryForm = ref<CategoryRecord>(emptyCategory())

const tenantCategories = computed(() => store.getCategoriesByTenant(selectedTenantId.value))
const selectedCategory = computed(() =>
  tenantCategories.value.find((item) => item.id === selectedCategoryId.value) ?? null,
)

const categoryTreeOptions = computed(() => {
  const byParent = new Map<number | null, CategoryRecord[]>()
  for (const category of tenantCategories.value) {
    const key = category.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(category)
    byParent.set(key, bucket)
  }

  const result: Array<{ id: number; label: string }> = []
  const walk = (parentId: number | null, depth: number) => {
    const items = byParent.get(parentId) ?? []
    for (const item of items) {
      if (item.id === categoryForm.value.id) continue
      result.push({
        id: item.id,
        label: `${'— '.repeat(depth)}${item.name}`,
      })
      walk(item.id, depth + 1)
    }
  }
  walk(null, 0)
  return result
})

watch(selectedTenantId, async (tenantId) => {
  if (!tenantId) return
  await store.fetchCategories(tenantId)
  const fallback = store.getCategoriesByTenant(tenantId)[0]
  selectedCategoryId.value = fallback?.id ?? 0
  categoryForm.value = fallback ? { ...fallback } : emptyCategory()
}, { immediate: true })

watch(selectedCategoryId, (categoryId) => {
  if (isCreating.value) return
  const category = tenantCategories.value.find((item) => item.id === categoryId)
  if (category) {
    categoryForm.value = { ...category }
  }
})

function startCreateCategory() {
  isCreating.value = true
  selectedCategoryId.value = 0
  categoryForm.value = emptyCategory()
}

function selectCategory(categoryId: number) {
  isCreating.value = false
  selectedCategoryId.value = categoryId
}

async function saveCategory() {
  if (!selectedTenantId.value) {
    alert('請先選擇租戶')
    return
  }
  if (!categoryForm.value.name.trim()) {
    alert('請先填寫分類名稱')
    return
  }

  try {
    if (isCreating.value || categoryForm.value.id === 0) {
      const created = await store.createCategory(selectedTenantId.value, {
        name: categoryForm.value.name,
        parentId: categoryForm.value.parentId,
        sortOrder: categoryForm.value.sortOrder,
      })
      isCreating.value = false
      selectedCategoryId.value = created.id
      categoryForm.value = { ...created }
      alert('分類已成功建立')
      return
    }

    await store.updateCategory({ ...categoryForm.value, tenantId: selectedTenantId.value })
    alert('分類已成功儲存')
  } catch (error) {
    console.error('保存分類失敗:', error)
    alert('保存分類失敗: ' + (error as Error).message)
  }
}

async function removeCategory() {
  if (!selectedCategory.value) return
  if (!confirm('確定要刪除此分類嗎？')) return

  try {
    await store.deleteCategory(selectedTenantId.value, selectedCategory.value.id)
    const fallback = store.getCategoriesByTenant(selectedTenantId.value)[0]
    selectedCategoryId.value = fallback?.id ?? 0
    categoryForm.value = fallback ? { ...fallback } : emptyCategory()
    isCreating.value = false
    alert('分類已成功刪除')
  } catch (error) {
    console.error('刪除分類失敗:', error)
    alert('刪除分類失敗: ' + (error as Error).message)
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
  <section class="categories-page">
    <div class="page-heading">
      <div>
        <p class="label">Category Center</p>
        <h2>分類管理</h2>
        <p class="subcopy">依租戶維護多級分類，方便商品在不同站點使用各自的分類樹。</p>
      </div>
      <button class="primary" type="button" @click="startCreateCategory">新增分類</button>
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
        <h3>分類清單</h3>
        <div class="item-list">
          <button
            v-for="category in tenantCategories"
            :key="category.id"
            type="button"
            class="item-row"
            :class="{ selected: category.id === selectedCategoryId && !isCreating }"
            @click="selectCategory(category.id)"
          >
            <div>
              <strong>{{ category.name }}</strong>
              <p>排序 {{ category.sortOrder }} · {{ category.parentId ? `父分類 #${category.parentId}` : '頂層分類' }}</p>
            </div>
          </button>
        </div>
      </article>

      <article class="panel editor-card">
        <div class="card-heading">
          <h3>{{ isCreating ? '新增分類' : '分類設定' }}</h3>
          <small v-if="!isCreating">ID {{ categoryForm.id }}</small>
        </div>

        <div class="form-grid">
          <label class="full">
            <span>分類名稱</span>
            <input v-model="categoryForm.name" />
          </label>
          <label>
            <span>父分類</span>
            <select v-model="categoryForm.parentId">
              <option :value="null">無，作為頂層分類</option>
              <option v-for="option in categoryTreeOptions" :key="option.id" :value="option.id">{{ option.label }}</option>
            </select>
          </label>
          <label>
            <span>排序值</span>
            <input v-model.number="categoryForm.sortOrder" type="number" min="0" />
          </label>
        </div>

        <div class="actions">
          <button v-if="!isCreating" class="danger" type="button" @click="removeCategory">刪除分類</button>
          <button class="primary" type="button" @click="saveCategory">{{ isCreating ? '建立分類' : '儲存分類' }}</button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.categories-page { display: grid; gap: 1rem; }
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
.form-grid { display: grid; gap: 0.875rem; grid-template-columns: repeat(2, 1fr); }
.full { grid-column: 1 / -1; }
input, select { width: 100%; min-height: 2.5rem; padding: 0.65rem 0.75rem; border: 1px solid var(--wp-border-strong); border-radius: 0.375rem; background: #fff; }
.actions { display: flex; justify-content: space-between; gap: 0.75rem; margin-top: 1rem; }
.primary, .danger { min-height: 2.5rem; padding: 0.65rem 1rem; border-radius: 0.375rem; font-weight: 600; border: 1px solid transparent; }
.primary { background: var(--wp-blue); color: #fff; border-color: var(--wp-blue); }
.danger { background: #fff; color: var(--wp-red); border-color: rgba(214, 54, 56, 0.3); }
@media (max-width: 900px) { .workspace-grid, .form-grid { grid-template-columns: 1fr; } }
</style>
