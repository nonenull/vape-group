<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Delete, Edit, Plus, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAdminStore } from '@/stores/admin'
import type { CategoryRecord } from '@/data/adminMock'

const store = useAdminStore()

const selectedCategoryId = ref(0)
const isCreating = ref(false)
const isCategoryModalOpen = ref(false)
const isViewMode = ref(false)

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

function sortCategories(items: CategoryRecord[]) {
  return [...items].sort((a, b) => {
    if (a.sortOrder !== b.sortOrder) {
      return a.sortOrder - b.sortOrder
    }
    return a.id - b.id
  })
}

function hydrateCategoryForm(category: CategoryRecord) {
  categoryForm.value = cloneCategory(category)
}

const categoriesByParent = computed(() => {
  const byParent = new Map<number | null, CategoryRecord[]>()
  for (const category of categories.value) {
    const key = category.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(category)
    byParent.set(key, bucket)
  }

  for (const [key, bucket] of byParent.entries()) {
    byParent.set(key, sortCategories(bucket))
  }

  return byParent
})

const categoryTreeList = computed(() => {
  const result: Array<CategoryRecord & {
    depth: number
    path: string
    childCount: number
  }> = []

  const walk = (parentId: number | null, depth: number, path: string) => {
    const items = categoriesByParent.value.get(parentId) ?? []
    items.forEach((item) => {
      const currentPath = path ? `${path} / ${item.name}` : item.name
      result.push({
        ...item,
        depth,
        path: currentPath,
        childCount: (categoriesByParent.value.get(item.id) ?? []).length,
      })
      walk(item.id, depth + 1, currentPath)
    })
  }

  walk(null, 0, '')
  return result
})

const categoryTreeOptions = computed(() => {
  const result: Array<{ id: number; label: string }> = []
  const walk = (parentId: number | null, depth: number) => {
    const items = categoriesByParent.value.get(parentId) ?? []
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

const selectedCategoryPath = computed(() => {
  if (!selectedCategory.value) return ''

  const segments: string[] = []
  let current: CategoryRecord | null = selectedCategory.value

  while (current) {
    segments.unshift(current.name)
    current = current.parentId
      ? categories.value.find((item) => item.id === current?.parentId) ?? null
      : null
  }

  return segments.join(' / ')
})

const selectedCategoryChildren = computed(() => {
  if (!selectedCategory.value) return []
  return categoriesByParent.value.get(selectedCategory.value.id) ?? []
})

watch(
  categories,
  (items) => {
    if (isCreating.value) return
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
  isViewMode.value = false
  isCategoryModalOpen.value = true
  categoryForm.value = emptyCategory()
}

function openViewCategory(categoryId: number) {
  const category = categories.value.find((item) => item.id === categoryId)
  if (!category) return
  isCreating.value = false
  isViewMode.value = true
  selectedCategoryId.value = categoryId
  hydrateCategoryForm(category)
  isCategoryModalOpen.value = true
}

function openEditCategory(categoryId: number) {
  const category = categories.value.find((item) => item.id === categoryId)
  if (!category) return
  isCreating.value = false
  isViewMode.value = false
  selectedCategoryId.value = categoryId
  hydrateCategoryForm(category)
  isCategoryModalOpen.value = true
}

function closeCategoryModal() {
  isCategoryModalOpen.value = false
  isViewMode.value = false
  if (selectedCategory.value) {
    hydrateCategoryForm(selectedCategory.value)
  } else {
    categoryForm.value = emptyCategory()
  }
  isCreating.value = false
}

async function saveCategory() {
  if (!categoryForm.value.name.trim()) {
    ElMessage.warning('請先填寫分類名稱')
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
      ElMessage.success('分類已成功建立')
      return
    }

    await store.updateCategory({ ...categoryForm.value })
    isCategoryModalOpen.value = false
    ElMessage.success('分類已成功儲存')
  } catch (error) {
    console.error('保存分類失敗:', error)
    ElMessage.error('保存分類失敗: ' + (error as Error).message)
  }
}

async function removeCategoryById(categoryId: number) {
  const category = categories.value.find((item) => item.id === categoryId)
  if (!category) return
  try {
    await ElMessageBox.confirm(`確定要刪除此分類「${category.name}」嗎？`, '刪除分類', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    await store.deleteCategory(category.id)
    const fallback = store.getCategories()[0]
    selectedCategoryId.value = fallback?.id ?? 0
    categoryForm.value = fallback ? cloneCategory(fallback) : emptyCategory()
    isCategoryModalOpen.value = false
    isCreating.value = false
    ElMessage.success('分類已成功刪除')
  } catch (error) {
    console.error('刪除分類失敗:', error)
    ElMessage.error('刪除分類失敗: ' + (error as Error).message)
  }
}

onMounted(() => {
  if (!store.categories.length) {
    store.fetchCategories()
  }
})
</script>

<template>
  <section class="category-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div>
            <span class="title">分類管理</span>
            <p class="subcopy">維護全租戶共用的分類樹，所有站點同步使用同一份分類資料。</p>
          </div>
          <div class="actions">
            <el-button type="success" :icon="Plus" @click="startCreateCategory">新增分類</el-button>
          </div>
        </div>
      </template>

      <el-table :data="categoryTreeList" stripe style="width: 100%">
        <el-table-column label="分類" min-width="280">
          <template #default="{ row }">
            <div class="category-name-cell" :style="{ paddingLeft: `${row.depth * 18}px` }">
              <strong>{{ row.name }}</strong>
              <small>{{ row.path }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="父分類" min-width="180">
          <template #default="{ row }">
            {{ row.parentId ? (categoryNameMap.get(row.parentId) ?? `#${row.parentId}`) : '頂層分類' }}
          </template>
        </el-table-column>
        <el-table-column prop="sortOrder" label="排序值" width="100" />
        <el-table-column label="子分類" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.childCount" size="small" type="info">{{ row.childCount }} 個</el-tag>
            <span v-else class="text-muted">0</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button size="small" :icon="View" @click="openViewCategory(row.id)">查看</el-button>
              <el-button size="small" type="primary" :icon="Edit" @click="openEditCategory(row.id)">編輯</el-button>
              <el-button size="small" type="danger" :icon="Delete" @click="removeCategoryById(row.id)">刪除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="isCategoryModalOpen"
      :title="isCreating ? '新增分類' : isViewMode ? '查看分類' : '修改分類'"
      width="720px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedCategory && !isCreating" class="overview-box">
        <div class="overview-title">
          <strong>{{ selectedCategory.name }}</strong>
          <small>ID {{ selectedCategory.id }}</small>
        </div>
        <div class="overview-grid">
          <div><span>父分類</span><strong>{{ selectedCategory.parentId ? (categoryNameMap.get(selectedCategory.parentId) ?? `#${selectedCategory.parentId}`) : '頂層分類' }}</strong></div>
          <div><span>排序值</span><strong>{{ selectedCategory.sortOrder }}</strong></div>
          <div><span>分類路徑</span><strong>{{ selectedCategoryPath }}</strong></div>
          <div><span>直屬子分類</span><strong>{{ selectedCategoryChildren.length }} 個</strong></div>
        </div>
      </div>

      <el-form label-width="100px">
        <el-form-item label="分類名稱">
          <el-input v-model="categoryForm.name" :disabled="isViewMode" />
        </el-form-item>
        <el-form-item label="父分類">
          <el-select v-model="categoryForm.parentId" :disabled="isViewMode" style="width: 100%">
            <el-option :value="null" label="無，作為頂層分類" />
            <el-option v-for="option in categoryTreeOptions" :key="option.id" :value="option.id" :label="option.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序值">
          <el-input-number v-model="categoryForm.sortOrder" :disabled="isViewMode" :min="0" style="width: 100%" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="closeCategoryModal">取消</el-button>
        <el-button v-if="!isViewMode" type="primary" @click="saveCategory">
          {{ isCreating ? '建立分類' : '儲存分類' }}
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.category-list {
  display: grid;
  gap: 1rem;
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

.subcopy {
  margin: 0.25rem 0 0;
  color: #909399;
  font-size: 13px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.category-name-cell strong {
  display: block;
}

.category-name-cell small,
.text-muted,
.overview-title small {
  color: #909399;
  font-size: 12px;
}

.overview-box {
  margin-bottom: 1rem;
  padding: 1rem;
  border: 1px solid #dcdfe6;
  border-radius: 12px;
  background: #f5f7fa;
}

.overview-title {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.85rem;
}

.overview-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.overview-grid div {
  display: grid;
  gap: 0.25rem;
}

.overview-grid span {
  color: #909399;
  font-size: 12px;
}

@media (max-width: 900px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .actions {
    width: 100%;
  }

  .overview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
