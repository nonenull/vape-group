<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Delete, Edit, Plus, Upload, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminAPI, resolveAssetURL } from '@/api/admin'
import { useAdminStore } from '@/stores/admin'
import type { BrandRecord } from '@/data/adminMock'

const store = useAdminStore()

const selectedBrandId = ref(0)
const isCreating = ref(false)
const isBrandModalOpen = ref(false)
const isViewMode = ref(false)
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
  isViewMode.value = false
  isBrandModalOpen.value = true
  brandForm.value = emptyBrand()
}

function openViewBrand(brandId: number) {
  isCreating.value = false
  isViewMode.value = true
  selectedBrandId.value = brandId
  const brand = brands.value.find((item) => item.id === brandId)
  if (brand) {
    hydrateBrandForm(brand)
  }
  isBrandModalOpen.value = true
}

function openEditBrand(brandId: number) {
  const brand = brands.value.find((item) => item.id === brandId)
  if (!brand) return
  isCreating.value = false
  isViewMode.value = false
  selectedBrandId.value = brandId
  hydrateBrandForm(brand)
  isBrandModalOpen.value = true
}

function closeBrandModal() {
  isBrandModalOpen.value = false
  isViewMode.value = false
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
    ElMessage.success('品牌 Logo 已上傳')
  } catch (error) {
    console.error('上傳品牌 Logo 失敗:', error)
    ElMessage.error('上傳品牌 Logo 失敗: ' + (error as Error).message)
  } finally {
    isUploadingLogo.value = false
    target.value = ''
  }
}

async function saveBrand() {
  if (!brandForm.value.name.trim()) {
    ElMessage.warning('請先填寫品牌名稱')
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
      isViewMode.value = false
      isBrandModalOpen.value = false
      selectedBrandId.value = created.id
      brandForm.value = cloneBrand(created)
      ElMessage.success('品牌已成功建立')
      return
    }

    await store.updateBrand({ ...brandForm.value })
    isViewMode.value = false
    isBrandModalOpen.value = false
    ElMessage.success('品牌已成功儲存')
  } catch (error) {
    console.error('保存品牌失敗:', error)
    ElMessage.error('保存品牌失敗: ' + (error as Error).message)
  }
}

async function removeBrandById(brandId: number) {
  const brand = brands.value.find((item) => item.id === brandId)
  if (!brand) return
  try {
    await ElMessageBox.confirm(`確定要刪除此品牌「${brand.name}」嗎？`, '刪除品牌', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    await store.deleteBrand(brand.id)
    const fallback = store.getBrands()[0]
    selectedBrandId.value = fallback?.id ?? 0
    brandForm.value = fallback ? cloneBrand(fallback) : emptyBrand()
    isBrandModalOpen.value = false
    isViewMode.value = false
    isCreating.value = false
    ElMessage.success('品牌已成功刪除')
  } catch (error) {
    console.error('刪除品牌失敗:', error)
    ElMessage.error('刪除品牌失敗: ' + (error as Error).message)
  }
}

onMounted(() => {
  if (!store.brands.length) {
    store.fetchBrands()
  }
})
</script>

<template>
  <section class="brand-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div>
            <span class="title">品牌管理</span>
            <p class="subcopy">維護全租戶共享的品牌資料，包含 Logo 與說明。</p>
          </div>
          <div class="actions">
            <el-button type="success" :icon="Plus" @click="startCreateBrand">新增品牌</el-button>
          </div>
        </div>
      </template>

      <el-table :data="brands" stripe style="width: 100%">
        <el-table-column label="品牌" min-width="200">
          <template #default="{ row }">
            <div class="brand-name-cell">
              <strong>{{ row.name }}</strong>
              <small>ID {{ row.id }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Logo" width="120">
          <template #default="{ row }">
            <img
              v-if="row.logoUrl"
              :src="displayImage(row.logoUrl)"
              :alt="row.name || 'brand logo'"
              class="brand-table-logo"
            />
            <span v-else class="text-muted">無 Logo</span>
          </template>
        </el-table-column>
        <el-table-column label="品牌描述" min-width="280">
          <template #default="{ row }">
            <span class="brand-description-col">{{ row.description || '尚未填寫品牌描述' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button size="small" :icon="View" @click="openViewBrand(row.id)">查看</el-button>
              <el-button size="small" type="primary" :icon="Edit" @click="openEditBrand(row.id)">編輯</el-button>
              <el-button size="small" type="danger" :icon="Delete" @click="removeBrandById(row.id)">刪除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="isBrandModalOpen"
      :title="isCreating ? '新增品牌' : isViewMode ? '查看品牌' : '修改品牌'"
      width="720px"
      :close-on-click-modal="false"
    >
      <div class="brand-preview" v-if="brandForm.logoUrl">
        <img :src="displayImage(brandForm.logoUrl)" :alt="brandForm.name || 'brand logo'" />
      </div>

      <el-form label-width="100px">
        <el-form-item label="品牌名稱">
          <el-input v-model="brandForm.name" :disabled="isViewMode" />
        </el-form-item>
        <el-form-item label="Logo URL">
          <el-input v-model="brandForm.logoUrl" :disabled="isViewMode" />
          <small>可直接貼上圖片網址，或使用下方按鈕上傳圖片。</small>
        </el-form-item>
        <el-form-item label="上傳 Logo">
          <input type="file" accept="image/*" :disabled="isUploadingLogo || isViewMode" @change="uploadBrandLogo" />
          <small>{{ isUploadingLogo ? '圖片上傳中...' : '支援 JPG、PNG、WEBP 等常見圖片格式。' }}</small>
        </el-form-item>
        <el-form-item label="品牌描述">
          <el-input v-model="brandForm.description" :disabled="isViewMode" type="textarea" :rows="5" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="closeBrandModal">取消</el-button>
        <el-button v-if="!isViewMode" type="primary" @click="saveBrand">
          {{ isCreating ? '建立品牌' : '儲存品牌' }}
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.brand-list {
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

.brand-name-cell strong {
  display: block;
}

.brand-name-cell small,
.brand-description-col,
.text-muted,
.el-form-item small {
  color: #909399;
  font-size: 12px;
}

.brand-table-logo {
  display: block;
  width: 88px;
  height: 56px;
  object-fit: contain;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  padding: 0.35rem;
}

.brand-preview {
  margin-bottom: 1rem;
  padding: 1rem;
  border: 1px solid #dcdfe6;
  border-radius: 12px;
  background: #f5f7fa;
}

.brand-preview img {
  display: block;
  max-width: 100%;
  max-height: 120px;
  object-fit: contain;
}

@media (max-width: 900px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .actions {
    width: 100%;
  }
}
</style>
