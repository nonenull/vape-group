<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { adminAPI, resolveAssetURL } from '@/api/admin'
import { useAdminStore } from '@/stores/admin'
import type { BrandRecord } from '@/data/adminMock'

const store = useAdminStore()

const selectedBrandId = ref(0)
const isCreating = ref(false)
const isBrandModalOpen = ref(false)
const isUploadingLogo = ref(false)

const emptyBrand = (): BrandRecord => ({
  id: 0,
  name: '',
  logoUrl: '',
  description: '',
})

function cloneBrand(brand: BrandRecord): BrandRecord {
  return { ...brand }
}

const brandForm = ref<BrandRecord>(emptyBrand())
const brands = computed(() => store.getBrands())
const selectedBrand = computed(() =>
  brands.value.find((item) => item.id === selectedBrandId.value) ?? null,
)

function hydrateBrandForm(brand: BrandRecord) {
  brandForm.value = cloneBrand(brand)
}

watch(brands, (items) => {
  if (isCreating.value) return
  const fallback = items[0]
  if (!selectedBrandId.value || !items.some((item) => item.id === selectedBrandId.value)) {
    selectedBrandId.value = fallback?.id ?? 0
    brandForm.value = fallback ? cloneBrand(fallback) : emptyBrand()
  }
}, { immediate: true })

watch(selectedBrandId, (brandId) => {
  if (isCreating.value) return
  const brand = brands.value.find((item) => item.id === brandId)
  if (brand) {
    hydrateBrandForm(brand)
  }
})

function startCreateBrand() {
  isCreating.value = true
  isBrandModalOpen.value = true
  brandForm.value = emptyBrand()
}

function selectBrand(brandId: number) {
  isCreating.value = false
  selectedBrandId.value = brandId
}

function openEditBrand(brandId: number) {
  const brand = brands.value.find((item) => item.id === brandId)
  if (!brand) return
  isCreating.value = false
  selectedBrandId.value = brandId
  hydrateBrandForm(brand)
  isBrandModalOpen.value = true
}

function closeBrandModal() {
  isBrandModalOpen.value = false
  if (selectedBrand.value) {
    hydrateBrandForm(selectedBrand.value)
  } else {
    brandForm.value = emptyBrand()
  }
  isCreating.value = false
  isUploadingLogo.value = false
}

const displayImage = (value: string) => resolveAssetURL(value)

async function uploadBrandLogo(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) {
    return
  }

  isUploadingLogo.value = true
  try {
    const uploaded = await adminAPI.uploadImage(file)
    brandForm.value.logoUrl = uploaded.url
  } catch (error) {
    console.error('上傳品牌 Logo 失敗:', error)
    alert('上傳品牌 Logo 失敗: ' + (error as Error).message)
  } finally {
    isUploadingLogo.value = false
    target.value = ''
  }
}

async function saveBrand() {
  if (!brandForm.value.name.trim()) {
    alert('請先填寫品牌名稱')
    return
  }

  try {
    if (isCreating.value || brandForm.value.id === 0) {
      const created = await store.createBrand({
        name: brandForm.value.name,
        logoUrl: brandForm.value.logoUrl,
        description: brandForm.value.description,
      })
      isCreating.value = false
      isBrandModalOpen.value = false
      selectedBrandId.value = created.id
      brandForm.value = cloneBrand(created)
      alert('品牌已成功建立')
      return
    }

    await store.updateBrand({ ...brandForm.value })
    isBrandModalOpen.value = false
    alert('品牌已成功儲存')
  } catch (error) {
    console.error('保存品牌失敗:', error)
    alert('保存品牌失敗: ' + (error as Error).message)
  }
}

async function removeBrand() {
  if (!selectedBrand.value) return
  await removeBrandById(selectedBrand.value.id)
}

async function removeBrandById(brandId: number) {
  const brand = brands.value.find((item) => item.id === brandId)
  if (!brand) return
  if (!confirm('確定要刪除此品牌嗎？')) return

  try {
    await store.deleteBrand(brand.id)
    const fallback = store.getBrands()[0]
    selectedBrandId.value = fallback?.id ?? 0
    brandForm.value = fallback ? cloneBrand(fallback) : emptyBrand()
    isBrandModalOpen.value = false
    isCreating.value = false
    alert('品牌已成功刪除')
  } catch (error) {
    console.error('刪除品牌失敗:', error)
    alert('刪除品牌失敗: ' + (error as Error).message)
  }
}

onMounted(() => {
  if (!store.brands.length) {
    store.fetchBrands()
  }
})
</script>

<template>
  <section class="brands-page">
    <div class="page-heading">
      <div>
        <p class="label">Brand Center</p>
        <h2>品牌管理</h2>
        <p class="subcopy">維護全租戶共享的品牌資料，包含 Logo 與說明，供所有商品建立時快速選用。</p>
      </div>
      <button class="primary" type="button" @click="startCreateBrand">新增品牌</button>
    </div>

    <div class="workspace-grid">
      <article class="panel list-card">
        <div class="card-heading compact">
          <h3>品牌清單</h3>
          <small>{{ brands.length }} 個品牌</small>
        </div>
        <div class="item-list">
          <article
            v-for="brand in brands"
            :key="brand.id"
            class="item-row"
            :class="{ selected: brand.id === selectedBrandId && !isCreating }"
            @click="selectBrand(brand.id)"
          >
            <div class="item-copy">
              <strong>{{ brand.name }}</strong>
              <p>{{ brand.description || '尚未填寫品牌描述' }}</p>
            </div>
            <div class="row-actions" @click.stop>
              <button class="secondary small-button" type="button" @click="openEditBrand(brand.id)">編輯</button>
              <button class="danger small-button" type="button" @click="removeBrandById(brand.id)">刪除</button>
            </div>
          </article>
        </div>
      </article>

      <article class="panel overview-card">
        <div class="card-heading">
          <h3>品牌概要</h3>
          <button
            v-if="selectedBrand"
            class="secondary"
            type="button"
            @click="openEditBrand(selectedBrand.id)"
          >
            修改品牌
          </button>
        </div>

        <div v-if="selectedBrand" class="overview-grid">
          <div class="brand-preview" v-if="selectedBrand.logoUrl">
            <img :src="displayImage(selectedBrand.logoUrl)" :alt="selectedBrand.name || 'brand logo'" />
          </div>
          <div class="overview-copy">
            <div class="overview-title">
              <h4>{{ selectedBrand.name }}</h4>
              <small>ID {{ selectedBrand.id }}</small>
            </div>
            <p class="overview-description">{{ selectedBrand.description || '尚未填寫品牌描述。' }}</p>
            <div class="overview-actions">
              <button class="danger" type="button" @click="removeBrandById(selectedBrand.id)">刪除品牌</button>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <p>請先從左側選擇品牌，再查看概要或進行編輯。</p>
        </div>
      </article>
    </div>

    <div
      v-if="isBrandModalOpen"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="closeBrandModal"
    >
      <article class="modal-card">
        <div class="card-heading modal-heading">
          <div>
            <h3>{{ isCreating ? '新增品牌' : '修改品牌' }}</h3>
            <small v-if="!isCreating">ID {{ brandForm.id }}</small>
          </div>
          <button class="secondary" type="button" @click="closeBrandModal">關閉</button>
        </div>

        <div class="modal-body">
          <div class="brand-preview" v-if="brandForm.logoUrl">
            <img :src="displayImage(brandForm.logoUrl)" :alt="brandForm.name || 'brand logo'" />
          </div>

          <div class="form-grid">
            <label class="full">
              <span>品牌名稱</span>
              <input v-model="brandForm.name" />
            </label>
            <label class="full">
              <span>Logo URL</span>
              <input v-model="brandForm.logoUrl" />
              <small>可直接貼上圖片網址，或使用下方按鈕上傳圖片。</small>
            </label>
            <label class="full upload-field">
              <span>上傳 Logo</span>
              <input type="file" accept="image/*" :disabled="isUploadingLogo" @change="uploadBrandLogo" />
              <small>{{ isUploadingLogo ? '圖片上傳中...' : '支援 JPG、PNG、WEBP 等常見圖片格式。' }}</small>
            </label>
            <label class="full">
              <span>品牌描述</span>
              <textarea v-model="brandForm.description" rows="5"></textarea>
            </label>
          </div>
        </div>

        <div class="actions modal-actions">
          <button v-if="!isCreating" class="danger" type="button" @click="removeBrand">刪除品牌</button>
          <div class="action-group">
            <button class="secondary" type="button" @click="closeBrandModal">取消</button>
            <button class="primary" type="button" @click="saveBrand">{{ isCreating ? '建立品牌' : '儲存品牌' }}</button>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.brands-page {
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

.tenant-switch label,
.form-grid label {
  display: grid;
  gap: 0.4rem;
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
.overview-description,
.form-grid small {
  color: var(--wp-text-muted);
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

.brand-preview {
  margin-bottom: 1rem;
  padding: 1rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.75rem;
  background: var(--wp-surface-soft);
}

.brand-preview img {
  display: block;
  max-width: 100%;
  max-height: 120px;
  object-fit: contain;
}

.form-grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: repeat(2, 1fr);
}

.upload-field input[type='file'] {
  padding: 0.55rem 0.65rem;
}

.full {
  grid-column: 1 / -1;
}

input,
select,
textarea {
  width: 100%;
  min-height: 2.5rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 0.375rem;
  background: #fff;
}

textarea {
  resize: vertical;
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
  .form-grid {
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
