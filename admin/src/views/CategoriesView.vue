<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAdminStore } from '@/stores/admin'
import type { CategoryRecord } from '@/data/adminMock'

const store = useAdminStore()

const selectedCategoryId = ref(0)
const isCreating = ref(false)
const isCategoryModalOpen = ref(false)

const emptyCategory = (): CategoryRecord => ({
  id: 0,
  name: '',
  parentId: null,
  sortOrder: 0,
})

function cloneCategory(category: CategoryRecord): CategoryRecord {
  return { ...category }
}

const categoryForm = ref<CategoryRecord>(emptyCategory())

const categories = computed(() => store.getCategories())
const selectedCategory = computed(() =>
  categories.value.find((item) => item.id === selectedCategoryId.value) ?? null,
)
const categoryNameMap = computed(() => new Map(categories.value.map((item) => [item.id, item.name])))

function hydrateCategoryForm(category: CategoryRecord) {
  categoryForm.value = cloneCategory(category)
}

const categoryTreeList = computed(() => {
  const byParent = new Map<number | null, CategoryRecord[]>()
  for (const category of categories.value) {
    const key = category.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(category)
    byParent.set(key, bucket)
  }

  const result: Array<CategoryRecord & { depth: number; path: string }> = []
  const walk = (parentId: number | null, depth: number, path: string) => {
    const items = byParent.get(parentId) ?? []
    for (const item of items) {
      const currentPath = path ? `${path} / ${item.name}` : item.name
      result.push({
        ...item,
        depth,
        path: currentPath,
      })
      walk(item.id, depth + 1, currentPath)
    }
  }

  walk(null, 0, '')
  return result
})

const categoryTreeOptions = computed(() => {
  const byParent = new Map<number | null, CategoryRecord[]>()
  for (const category of categories.value) {
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

watch(
  categories,
  (items) => {
    if (isCreating.value) {
      return
    }
    const fallback = items[0]
    if (!selectedCategoryId.value || !items.some((item) => item.id === selectedCategoryId.value)) {
      selectedCategoryId.value = fallback?.id ?? 0
      categoryForm.value = fallback ? cloneCategory(fallback) : emptyCategory()
    }
  },
  { immediate: true },
)

watch(selectedCategoryId, (categoryId) => {
  if (isCreating.value) return
  const category = categories.value.find((item) => item.id === categoryId)
  if (category) {
    hydrateCategoryForm(category)
  }
})

function startCreateCategory() {
  isCreating.value = true
  isCategoryModalOpen.value = true
  categoryForm.value = emptyCategory()
}

function selectCategory(categoryId: number) {
  isCreating.value = false
  selectedCategoryId.value = categoryId
}

function openEditCategory(categoryId: number) {
  const category = categories.value.find((item) => item.id === categoryId)
  if (!category) return
  isCreating.value = false
  selectedCategoryId.value = categoryId
  hydrateCategoryForm(category)
  isCategoryModalOpen.value = true
}

function closeCategoryModal() {
  isCategoryModalOpen.value = false
  if (selectedCategory.value) {
    hydrateCategoryForm(selectedCategory.value)
  } else {
    categoryForm.value = emptyCategory()
  }
  isCreating.value = false
}

async function saveCategory() {
  if (!categoryForm.value.name.trim()) {
    alert('請先填寫分類名稱')
    return
  }

  try {
    if (isCreating.value || categoryForm.value.id === 0) {
      const created = await store.createCategory({
        name: categoryForm.value.name,
        parentId: categoryForm.value.parentId,
        sortOrder: categoryForm.value.sortOrder,
      })
      isCreating.value = false
      isCategoryModalOpen.value = false
      selectedCategoryId.value = created.id
      categoryForm.value = cloneCategory(created)
      alert('分類已成功建立')
      return
    }

    await store.updateCategory({ ...categoryForm.value })
    isCategoryModalOpen.value = false
    alert('分類已成功儲存')
  } catch (error) {
    console.error('保存分類失敗:', error)
    alert('保存分類失敗: ' + (error as Error).message)
  }
}

async function removeCategory() {
  if (!selectedCategory.value) return
  await removeCategoryById(selectedCategory.value.id)
}

async function removeCategoryById(categoryId: number) {
  const category = categories.value.find((item) => item.id === categoryId)
  if (!category) return
  if (!confirm('確定要刪除此分類嗎？')) return

  try {
    await store.deleteCategory(category.id)
    const fallback = store.getCategories()[0]
    selectedCategoryId.value = fallback?.id ?? 0
    categoryForm.value = fallback ? cloneCategory(fallback) : emptyCategory()
    isCategoryModalOpen.value = false
    isCreating.value = false
    alert('分類已成功刪除')
  } catch (error) {
    console.error('刪除分類失敗:', error)
    alert('刪除分類失敗: ' + (error as Error).message)
  }
}

onMounted(() => {
  if (!store.categories.length) {
    store.fetchCategories()
  }
})
</script>

<template>
  <section class="categories-page">
    <div class="page-heading">
      <div>
        <p class="label">Category Center</p>
        <h2>分類管理</h2>
        <p class="subcopy">維護一套全租戶共用的分類樹，所有站點同步使用同一份分類資料。</p>
      </div>
      <button class="primary" type="button" @click="startCreateCategory">新增分類</button>
    </div>

    <div class="workspace-grid">
      <article class="panel list-card">
        <div class="card-heading compact">
          <h3>共享分類清單</h3>
          <small>{{ categories.length }} 個分類</small>
        </div>
        <div class="item-list">
          <article
            v-for="category in categoryTreeList"
            :key="category.id"
            class="item-row"
            :class="{ selected: category.id === selectedCategoryId && !isCreating }"
            @click="selectCategory(category.id)"
          >
            <div class="item-copy">
              <strong
                class="tree-label"
                :style="{ '--depth': category.depth }"
              >
                <span v-if="category.depth > 0" class="tree-branch" aria-hidden="true"></span>
                {{ category.name }}
              </strong>
              <p>
                排序 {{ category.sortOrder }} ·
                {{ category.parentId ? `父分類 ${categoryNameMap.get(category.parentId) ?? `#${category.parentId}`}` : '頂層分類' }}
              </p>
              <small class="tree-path">{{ category.path }}</small>
            </div>
            <div class="row-actions" @click.stop>
              <button class="secondary small-button" type="button" @click="openEditCategory(category.id)">編輯</button>
              <button class="danger small-button" type="button" @click="removeCategoryById(category.id)">刪除</button>
            </div>
          </article>
        </div>
      </article>

      <article class="panel overview-card">
        <div class="card-heading">
          <h3>分類概要</h3>
          <button
            v-if="selectedCategory"
            class="secondary"
            type="button"
            @click="openEditCategory(selectedCategory.id)"
          >
            修改分類
          </button>
        </div>

        <div v-if="selectedCategory" class="overview-grid">
          <div class="overview-copy">
            <div class="overview-title">
              <h4>{{ selectedCategory.name }}</h4>
              <small>ID {{ selectedCategory.id }}</small>
            </div>
            <dl class="overview-details">
              <div>
                <dt>父分類</dt>
                <dd>{{ selectedCategory.parentId ? `#${selectedCategory.parentId}` : '頂層分類' }}</dd>
              </div>
              <div>
                <dt>排序值</dt>
                <dd>{{ selectedCategory.sortOrder }}</dd>
              </div>
            </dl>
            <div class="overview-actions">
              <button class="danger" type="button" @click="removeCategoryById(selectedCategory.id)">刪除分類</button>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <p>請先從左側選擇分類，再查看概要或進行編輯。</p>
        </div>
      </article>
    </div>

    <div
      v-if="isCategoryModalOpen"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="closeCategoryModal"
    >
      <article class="modal-card">
        <div class="card-heading modal-heading">
          <div>
            <h3>{{ isCreating ? '新增分類' : '修改分類' }}</h3>
            <small v-if="!isCreating">ID {{ categoryForm.id }}</small>
          </div>
          <button class="secondary" type="button" @click="closeCategoryModal">關閉</button>
        </div>

        <div class="modal-body">
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
        </div>

        <div class="actions modal-actions">
          <button v-if="!isCreating" class="danger" type="button" @click="removeCategory">刪除分類</button>
          <div class="action-group">
            <button class="secondary" type="button" @click="closeCategoryModal">取消</button>
            <button class="primary" type="button" @click="saveCategory">{{ isCreating ? '建立分類' : '儲存分類' }}</button>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.categories-page {
  display: grid;
  gap: 1rem;
}

.page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.label {
  color: var(--wp-blue);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: 0.75rem;
}

.subcopy {
  color: var(--wp-text-muted);
  max-width: 72ch;
}

.panel,
.modal-card {
  background: #fff;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
}

.panel,
.modal-card {
  padding: 1rem 1.25rem;
}

.workspace-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 0.9fr 1.1fr;
}

.card-heading {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.card-heading.compact {
  margin-bottom: 0.75rem;
}

.card-heading small {
  color: var(--wp-text-muted);
}

.item-list {
  display: grid;
  gap: 0.75rem;
}

.item-row {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  padding: 0.9rem 1rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
  cursor: pointer;
}

.item-row.selected {
  border-color: var(--wp-blue);
  background: #f0f6fc;
}

.item-copy strong {
  display: block;
  margin-bottom: 0.15rem;
}

.item-copy p,
.tree-path {
  color: var(--wp-text-muted);
}

.tree-label {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding-left: calc(var(--depth) * 1.25rem);
}

.tree-branch {
  width: 0.9rem;
  height: 0.9rem;
  border-left: 1px solid var(--wp-border-strong);
  border-bottom: 1px solid var(--wp-border-strong);
  border-bottom-left-radius: 0.35rem;
  flex: 0 0 auto;
}

.tree-path {
  display: block;
  margin-top: 0.2rem;
  font-size: 0.78rem;
}

.row-actions {
  display: flex;
  gap: 0.45rem;
}

.overview-grid,
.overview-copy {
  display: grid;
  gap: 1rem;
}

.overview-title h4 {
  margin: 0;
  font-size: 1.15rem;
}

.overview-title small {
  color: var(--wp-text-muted);
}

.overview-details {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.overview-details div {
  padding: 0.75rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.625rem;
  background: var(--wp-surface-soft);
}

.overview-details dt {
  margin: 0 0 0.25rem;
  color: var(--wp-text-muted);
  font-size: 0.8rem;
}

.overview-details dd {
  margin: 0;
  font-weight: 600;
}

.overview-actions {
  display: flex;
}

.empty-state {
  padding: 1rem;
  border: 1px dashed var(--wp-border-strong);
  border-radius: 0.625rem;
  color: var(--wp-text-muted);
  background: var(--wp-surface-soft);
}

.form-grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: repeat(2, 1fr);
}

.form-grid label {
  display: grid;
  gap: 0.4rem;
}

.full {
  grid-column: 1 / -1;
}

input,
select {
  width: 100%;
  min-height: 2.5rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 0.375rem;
  background: #fff;
}

.actions {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 1rem;
}

.action-group,
.modal-actions {
  display: flex;
  gap: 0.75rem;
}

.primary,
.danger,
.secondary {
  min-height: 2.5rem;
  padding: 0.65rem 1rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid transparent;
}

.small-button {
  min-height: 2rem;
  padding: 0.4rem 0.75rem;
}

.primary {
  background: var(--wp-blue);
  color: #fff;
  border-color: var(--wp-blue);
}

.secondary {
  background: #fff;
  color: var(--wp-blue);
  border-color: rgba(34, 113, 177, 0.24);
}

.danger {
  background: #fff;
  color: var(--wp-red);
  border-color: rgba(214, 54, 56, 0.3);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  overflow-y: auto;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
}

.modal-card {
  width: min(720px, 100%);
  height: min(calc(100vh - 3rem), 100%);
  max-height: calc(100vh - 3rem);
  overflow: hidden;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  gap: 0;
}

.modal-heading {
  margin-bottom: 0;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid var(--wp-border);
}

.modal-body {
  min-height: 0;
  overflow-y: auto;
  scrollbar-gutter: stable;
  overscroll-behavior: contain;
  padding-top: 0.85rem;
}

@media (max-width: 900px) {
  .workspace-grid,
  .form-grid,
  .overview-details {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 780px) {
  .page-heading,
  .actions,
  .modal-actions,
  .action-group,
  .row-actions {
    flex-direction: column;
  }

  .item-row {
    grid-template-columns: 1fr;
  }

  .modal-overlay {
    padding: 0.75rem;
  }

  .modal-card {
    height: min(calc(100vh - 1.5rem), 100%);
    max-height: calc(100vh - 1.5rem);
  }
}
</style>
