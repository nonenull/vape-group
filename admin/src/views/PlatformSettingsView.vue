<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAdminStore } from '@/stores/admin'
import type { PlatformConfigRecord } from '@/data/adminMock'

const store = useAdminStore()
const isSaving = ref(false)
const form = ref<PlatformConfigRecord>({
  id: 0,
  lineContactUrl: '',
  featuredCategoryIds: [],
  featuredBrandIds: [],
})

const categories = computed(() => store.getCategories())
const brands = computed(() => store.getBrands())
const categoryLookup = computed(() => new Map(categories.value.map((item) => [item.id, item])))
const brandLookup = computed(() => new Map(brands.value.map((item) => [item.id, item])))
const categoryOptions = computed(() => {
  const byParent = new Map<number | null, typeof categories.value>()
  for (const category of categories.value) {
    const key = category.parentId ?? null
    const bucket = byParent.get(key) ?? []
    bucket.push(category)
    byParent.set(key, bucket)
  }

  const result: Array<{ id: number; label: string }> = []
  const walk = (parentId: number | null, depth: number, path: string[]) => {
    const items = (byParent.get(parentId) ?? [])
      .slice()
      .sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id)
    for (const item of items) {
      const currentPath = [...path, item.name]
      result.push({
        id: item.id,
        label: `${'— '.repeat(depth)}${currentPath.join(' / ')}`,
      })
      walk(item.id, depth + 1, currentPath)
    }
  }

  walk(null, 0, [])
  return result
})

onMounted(async () => {
  if (!store.categories.length) {
    await store.fetchCategories()
  }
  if (!store.brands.length) {
    await store.fetchBrands()
  }
  if (!store.platformConfig.id) {
    await store.fetchPlatformConfig()
  }
  form.value = { ...store.platformConfig }
})

function toggleFeaturedCategory(categoryId: number) {
  const current = new Set(form.value.featuredCategoryIds)
  if (current.has(categoryId)) {
    current.delete(categoryId)
  } else {
    current.add(categoryId)
  }
  form.value.featuredCategoryIds = Array.from(current)
}

function isFeaturedCategory(categoryId: number) {
  return form.value.featuredCategoryIds.includes(categoryId)
}

function toggleFeaturedBrand(brandId: number) {
  const current = new Set(form.value.featuredBrandIds)
  if (current.has(brandId)) {
    current.delete(brandId)
  } else {
    current.add(brandId)
  }
  form.value.featuredBrandIds = Array.from(current)
}

function isFeaturedBrand(brandId: number) {
  return form.value.featuredBrandIds.includes(brandId)
}

async function savePlatformConfig() {
  isSaving.value = true
  try {
    const updated = await store.updatePlatformConfig({
      id: form.value.id,
      lineContactUrl: form.value.lineContactUrl.trim(),
      featuredCategoryIds: form.value.featuredCategoryIds,
      featuredBrandIds: form.value.featuredBrandIds,
    })
    form.value = { ...updated }
    alert('平台配置已成功儲存')
  } catch (error) {
    console.error('儲存平台配置失敗:', error)
    alert('儲存平台配置失敗: ' + (error as Error).message)
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <section class="platform-settings-page">
    <div class="page-heading">
      <div>
        <p class="label">Platform Settings</p>
        <h2>平台配置</h2>
        <p class="subcopy">統一設定前台全站共用的客服超連結，右下角 LINE 懸浮按鈕會直接使用這個網址跳轉。</p>
      </div>
    </div>

    <article class="settings-card">
      <div class="card-heading">
        <h3>客服連結</h3>
        <small>全站共用</small>
      </div>

      <div class="form-grid">
        <label class="full">
          <span>LINE 超連結</span>
          <input
            v-model="form.lineContactUrl"
            type="url"
            placeholder="例如：https://line.me/R/ti/p/@your-line-id"
          />
          <small>前台右下角客服按鈕將直接跳轉到這個網址。</small>
        </label>
      </div>

      <div class="divider"></div>

      <div class="card-heading">
        <h3>首頁熱門分類</h3>
        <small>可多選顯示</small>
      </div>

      <div class="featured-category-panel">
        <p class="helper-text">
          勾選後，前台首頁會以左右滑動的方塊卡片顯示這些分類；未勾選時，首頁會自動使用有商品的分類作為備用。
        </p>

        <div v-if="categoryOptions.length" class="category-checklist">
          <label
            v-for="option in categoryOptions"
            :key="option.id"
            class="category-pill"
            :class="{ active: isFeaturedCategory(option.id) }"
          >
            <input
              :checked="isFeaturedCategory(option.id)"
              type="checkbox"
              @change="toggleFeaturedCategory(option.id)"
            >
            <span>{{ option.label }}</span>
          </label>
        </div>
        <p v-else class="helper-text">暂无可配置分类，请先到分类管理中建立分类。</p>

        <div v-if="form.featuredCategoryIds.length" class="selected-summary">
          已选择：
          {{ form.featuredCategoryIds.map((id) => categoryLookup.get(id)?.name ?? `#${id}`).join('、') }}
        </div>
      </div>

      <div class="divider"></div>

      <div class="card-heading">
        <h3>首頁品牌卡片</h3>
        <small>可多選顯示</small>
      </div>

      <div class="featured-category-panel">
        <p class="helper-text">
          勾選後，前台首頁會顯示品牌卡片並帶出品牌 Logo；未勾選時，首頁會自動使用已有 Logo 的品牌作為備用。
        </p>

        <div v-if="brands.length" class="category-checklist">
          <label
            v-for="brand in brands"
            :key="brand.id"
            class="category-pill"
            :class="{ active: isFeaturedBrand(brand.id) }"
          >
            <input
              :checked="isFeaturedBrand(brand.id)"
              type="checkbox"
              @change="toggleFeaturedBrand(brand.id)"
            >
            <span>{{ brand.name }}</span>
          </label>
        </div>
        <p v-else class="helper-text">暂无可配置品牌，请先到品牌管理中建立品牌。</p>

        <div v-if="form.featuredBrandIds.length" class="selected-summary">
          已选择：
          {{ form.featuredBrandIds.map((id) => brandLookup.get(id)?.name ?? `#${id}`).join('、') }}
        </div>
      </div>

      <div class="actions">
        <button class="primary" type="button" :disabled="isSaving" @click="savePlatformConfig">
          {{ isSaving ? '儲存中...' : '儲存平台配置' }}
        </button>
      </div>
    </article>
  </section>
</template>

<style scoped>
.platform-settings-page {
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

.page-heading h2 {
  margin: 0.35rem 0 0.45rem;
}

.subcopy,
.form-grid small,
.card-heading small,
.helper-text {
  color: var(--wp-text-muted);
}

.settings-card {
  background: #fff;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1rem 1.25rem;
}

.card-heading {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.form-grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: repeat(2, 1fr);
}

label {
  display: grid;
  gap: 0.4rem;
}

label span {
  font-weight: 600;
}

.full {
  grid-column: 1 / -1;
}

input {
  width: 100%;
  min-height: 2.5rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 0.375rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
}

.divider {
  height: 1px;
  margin: 1.25rem 0;
  background: var(--wp-border);
}

.featured-category-panel {
  display: grid;
  gap: 0.9rem;
}

.category-checklist {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.category-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  min-height: 2.75rem;
  padding: 0.7rem 0.85rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.875rem;
  background: #f8fafc;
  cursor: pointer;
  transition: border-color 0.2s ease, background 0.2s ease, transform 0.2s ease;
}

.category-pill input {
  width: auto;
  min-height: auto;
  margin: 0;
}

.category-pill span {
  font-weight: 600;
}

.category-pill.active {
  border-color: var(--wp-blue);
  background: rgba(34, 113, 177, 0.08);
  transform: translateY(-1px);
}

.selected-summary {
  font-size: 0.95rem;
  color: var(--wp-heading);
}

.primary {
  min-height: 2.5rem;
  padding: 0.65rem 1rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid var(--wp-blue);
  background: var(--wp-blue);
  color: #fff;
}

@media (max-width: 960px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
