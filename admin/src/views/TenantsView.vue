<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Delete, Edit, Plus, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminAPI } from '@/api/admin'
import { useAdminStore } from '@/stores/admin'
import type { DomainRecord, TenantRecord } from '@/data/adminMock'

const store = useAdminStore()
const brands = computed(() => store.getBrands())
const domains = ref<DomainRecord[]>([])

const emptyTenant = (): TenantRecord => ({
  id: 0,
  domain: '',
  boundDomains: [],
  name: '',
  isActive: true,
  theme: '',
  homeTemplate: 'classic',
  homeModuleOrder: ['categories', 'brands', 'products'],
  primaryBrandId: null,
  previewImage: '/src/assets/logo.svg',
  logoImage: '/src/assets/logo.svg',
  accentColor: '#2271b1',
  accentStrongColor: '#135e96',
  surfaceColor: '#f6f7f7',
  pageBgColor: '#f3f6f9',
  cardBgColor: '#ffffff',
  textColor: '#1d2327',
  mutedTextColor: '#646970',
  borderColor: '#dcdcde',
  heroBgColor: '#f0f6fc',
  tagBgColor: '#f0f6fc',
  seoTitle: '',
  seoDescription: '',
})

const selectedTenantId = ref(store.tenants[0]?.id ?? 0)
const tenantForm = ref<TenantRecord>(store.tenants[0] ? { ...store.tenants[0] } : emptyTenant())
const isCreating = ref(false)
const isTenantModalOpen = ref(false)
const isViewMode = ref(false)
const boundDomainsInput = ref('')
const isSaving = ref(false)
const selectedTenantIds = ref<number[]>([])

const themePresets = [
  {
    key: 'ocean',
    label: 'Ocean',
    theme: 'ocean',
    homeTemplate: 'classic',
    accentColor: '#2271b1',
    accentStrongColor: '#135e96',
    surfaceColor: '#f6f7f7',
    pageBgColor: '#f3f6f9',
    cardBgColor: '#ffffff',
    textColor: '#1d2327',
    mutedTextColor: '#646970',
    borderColor: '#dcdcde',
    heroBgColor: '#eaf3fb',
    tagBgColor: '#f0f6fc',
    primaryBrandId: null,
  },
  {
    key: 'sunset',
    label: 'Sunset',
    theme: 'sunset',
    homeTemplate: 'campaign',
    accentColor: '#c75b12',
    accentStrongColor: '#9d4308',
    surfaceColor: '#fff4eb',
    pageBgColor: '#fff8f2',
    cardBgColor: '#fffdfb',
    textColor: '#2e1b10',
    mutedTextColor: '#7d5b4b',
    borderColor: '#f1d6c4',
    heroBgColor: '#ffe8d8',
    tagBgColor: '#ffefe2',
    primaryBrandId: null,
  },
  {
    key: 'forest',
    label: 'Forest',
    theme: 'forest',
    homeTemplate: 'brand-first',
    accentColor: '#2f6f4f',
    accentStrongColor: '#1f5238',
    surfaceColor: '#f2f7f3',
    pageBgColor: '#f6fbf7',
    cardBgColor: '#ffffff',
    textColor: '#173124',
    mutedTextColor: '#5f7b6d',
    borderColor: '#d6e5da',
    heroBgColor: '#e2f0e6',
    tagBgColor: '#eaf5ed',
    primaryBrandId: null,
  },
] as const

const moduleLabels: Record<string, string> = {
  categories: '热门分类',
  brands: '精选品牌',
  products: '本季热卖',
}

const selectedTenant = computed(() => store.tenants.find((item) => item.id === selectedTenantId.value) ?? null)

const availablePrimaryDomains = computed(() => {
  const currentTenantId = tenantForm.value.id
  const usedByOthers = new Set(
    store.tenants
      .filter((tenant) => tenant.id !== currentTenantId)
      .map((tenant) => tenant.domain.trim().toLowerCase())
      .filter(Boolean),
  )
  return domains.value.filter((domain) => !usedByOthers.has(domain.domainName.trim().toLowerCase()))
})

const availableBoundDomains = computed(() => {
  const currentTenantId = tenantForm.value.id
  const usedByOthers = new Set<string>()
  for (const tenant of store.tenants) {
    if (tenant.id === currentTenantId) continue
    if (tenant.domain.trim()) usedByOthers.add(tenant.domain.trim().toLowerCase())
    for (const bound of tenant.boundDomains) {
      if (bound.trim()) usedByOthers.add(bound.trim().toLowerCase())
    }
  }
  return domains.value.filter((domain) => !usedByOthers.has(domain.domainName.trim().toLowerCase()))
})

watch(
  selectedTenantId,
  (tenantId) => {
    if (isCreating.value) return
    const tenant = store.tenants.find((item) => item.id === tenantId)
    if (tenant) {
      tenantForm.value = { ...tenant, boundDomains: [...tenant.boundDomains] }
      boundDomainsInput.value = tenant.boundDomains.join('\n')
    }
  },
  { immediate: true },
)

function startCreateTenant() {
  isCreating.value = true
  isViewMode.value = false
  selectedTenantId.value = 0
  tenantForm.value = emptyTenant()
  boundDomainsInput.value = ''
  isTenantModalOpen.value = true
}

function openEditTenant(tenantId: number) {
  isCreating.value = false
  isViewMode.value = false
  selectedTenantId.value = tenantId
  const tenant = store.tenants.find((item) => item.id === tenantId)
  if (tenant) {
    tenantForm.value = { ...tenant, boundDomains: [...tenant.boundDomains] }
    boundDomainsInput.value = tenant.boundDomains.join('\n')
  }
  isTenantModalOpen.value = true
}

function openViewTenant(tenantId: number) {
  isCreating.value = false
  isViewMode.value = true
  selectedTenantId.value = tenantId
  const tenant = store.tenants.find((item) => item.id === tenantId)
  if (tenant) {
    tenantForm.value = { ...tenant, boundDomains: [...tenant.boundDomains] }
    boundDomainsInput.value = tenant.boundDomains.join('\n')
  }
  isTenantModalOpen.value = true
}

function closeTenantModal() {
  isTenantModalOpen.value = false
  isCreating.value = false
  isViewMode.value = false
  const fallback = selectedTenant.value ?? store.tenants[0] ?? null
  if (fallback) {
    selectedTenantId.value = fallback.id
    tenantForm.value = { ...fallback, boundDomains: [...fallback.boundDomains] }
    boundDomainsInput.value = fallback.boundDomains.join('\n')
    return
  }
  selectedTenantId.value = 0
  tenantForm.value = emptyTenant()
  boundDomainsInput.value = ''
}

function handleSelectionChange(rows: TenantRecord[]) {
  selectedTenantIds.value = rows.map((row) => row.id)
}

function appendBoundDomain(domainName: string) {
  const normalized = domainName.trim().toLowerCase()
  if (!normalized) return
  const current = boundDomainsInput.value
    .split('\n')
    .map((item) => item.trim().toLowerCase())
    .filter(Boolean)
  if (current.includes(normalized) || normalized === tenantForm.value.domain.trim().toLowerCase()) {
    return
  }
  boundDomainsInput.value = [...current, normalized].join('\n')
}

async function saveTenant() {
  if (!tenantForm.value.name.trim()) {
    ElMessage.warning('請先填寫租戶名稱')
    return
  }
  if (!tenantForm.value.domain.trim()) {
    ElMessage.warning('請先填寫主域名')
    return
  }

  tenantForm.value.boundDomains = boundDomainsInput.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
  tenantForm.value.homeModuleOrder = (tenantForm.value.homeModuleOrder ?? []).filter(Boolean)

  isSaving.value = true
  try {
    if (isCreating.value || tenantForm.value.id === 0) {
      const created = await store.createTenant({
        domain: tenantForm.value.domain,
        boundDomains: tenantForm.value.boundDomains,
        name: tenantForm.value.name,
        isActive: tenantForm.value.isActive,
        theme: tenantForm.value.theme,
        homeTemplate: tenantForm.value.homeTemplate,
        homeModuleOrder: tenantForm.value.homeModuleOrder,
        primaryBrandId: tenantForm.value.primaryBrandId,
        previewImage: tenantForm.value.previewImage,
        logoImage: tenantForm.value.logoImage,
        accentColor: tenantForm.value.accentColor,
        accentStrongColor: tenantForm.value.accentStrongColor,
        surfaceColor: tenantForm.value.surfaceColor,
        pageBgColor: tenantForm.value.pageBgColor,
        cardBgColor: tenantForm.value.cardBgColor,
        textColor: tenantForm.value.textColor,
        mutedTextColor: tenantForm.value.mutedTextColor,
        borderColor: tenantForm.value.borderColor,
        heroBgColor: tenantForm.value.heroBgColor,
        tagBgColor: tenantForm.value.tagBgColor,
        heroTitle: tenantForm.value.heroTitle,
        tagline: tenantForm.value.tagline,
        announcement: tenantForm.value.announcement,
        supportText: tenantForm.value.supportText,
        seoTitle: tenantForm.value.seoTitle,
        seoDescription: tenantForm.value.seoDescription,
      })
      isCreating.value = false
      isTenantModalOpen.value = false
      selectedTenantId.value = created.id
      tenantForm.value = { ...created, boundDomains: [...created.boundDomains] }
      boundDomainsInput.value = created.boundDomains.join('\n')
      ElMessage.success('租戶已成功建立')
      return
    }

    await store.updateTenant({ ...tenantForm.value, boundDomains: [...tenantForm.value.boundDomains] })
    isTenantModalOpen.value = false
    ElMessage.success('租戶設定已成功儲存')
  } catch (error) {
    console.error('保存租戶失敗:', error)
    ElMessage.error('保存租戶失敗: ' + (error as Error).message)
  } finally {
    isSaving.value = false
  }
}

async function removeTenantById(tenantId: number) {
  const tenant = store.tenants.find((item) => item.id === tenantId)
  if (!tenant) return
  try {
    await ElMessageBox.confirm(`確定要刪除租戶「${tenant.name}」嗎？`, '刪除租戶', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    await store.deleteTenant(tenantId)
    selectedTenantIds.value = selectedTenantIds.value.filter((id) => id !== tenantId)
    if (selectedTenantId.value === tenantId) {
      const fallback = store.tenants[0]
      selectedTenantId.value = fallback?.id ?? 0
      tenantForm.value = fallback ? { ...fallback, boundDomains: [...fallback.boundDomains] } : emptyTenant()
      boundDomainsInput.value = fallback?.boundDomains.join('\n') ?? ''
      isTenantModalOpen.value = false
      isViewMode.value = false
      isCreating.value = false
    }
    ElMessage.success('租戶已成功刪除')
  } catch (error) {
    console.error('刪除租戶失敗:', error)
    ElMessage.error('刪除租戶失敗: ' + (error as Error).message)
  }
}

async function removeSelectedTenants() {
  if (!selectedTenantIds.value.length) {
    ElMessage.warning('請先勾選至少一個租戶')
    return
  }
  try {
    await ElMessageBox.confirm(`確定要刪除選中的 ${selectedTenantIds.value.length} 個租戶嗎？`, '批量刪除租戶', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    for (const tenantId of [...selectedTenantIds.value]) {
      await store.deleteTenant(tenantId)
    }
    selectedTenantIds.value = []
    isCreating.value = false
    isTenantModalOpen.value = false
    const fallback = store.tenants[0]
    selectedTenantId.value = fallback?.id ?? 0
    tenantForm.value = fallback ? { ...fallback, boundDomains: [...fallback.boundDomains] } : emptyTenant()
    boundDomainsInput.value = fallback?.boundDomains.join('\n') ?? ''
    ElMessage.success('已成功刪除所選租戶')
  } catch (error) {
    console.error('批量刪除租戶失敗:', error)
    ElMessage.error('批量刪除租戶失敗: ' + (error as Error).message)
  }
}

function applyThemePreset(presetKey: string) {
  const preset = themePresets.find((item) => item.key === presetKey)
  if (!preset) return

  tenantForm.value.theme = preset.theme
  tenantForm.value.homeTemplate = preset.homeTemplate
  tenantForm.value.accentColor = preset.accentColor
  tenantForm.value.accentStrongColor = preset.accentStrongColor
  tenantForm.value.surfaceColor = preset.surfaceColor
  tenantForm.value.pageBgColor = preset.pageBgColor
  tenantForm.value.cardBgColor = preset.cardBgColor
  tenantForm.value.textColor = preset.textColor
  tenantForm.value.mutedTextColor = preset.mutedTextColor
  tenantForm.value.borderColor = preset.borderColor
  tenantForm.value.heroBgColor = preset.heroBgColor
  tenantForm.value.tagBgColor = preset.tagBgColor
}

function moveModule(index: number, direction: -1 | 1) {
  const modules = [...(tenantForm.value.homeModuleOrder ?? [])]
  const target = index + direction
  if (target < 0 || target >= modules.length) return
  const currentValue = modules[index]
  const targetValue = modules[target]
  if (!currentValue || !targetValue) return
  modules[index] = targetValue
  modules[target] = currentValue
  tenantForm.value.homeModuleOrder = modules
}

onMounted(async () => {
  if (!store.tenants.length) {
    await store.fetchTenants()
  }
  if (!store.brands.length) {
    await store.fetchBrands()
  }
  domains.value = await adminAPI.getDomains()

  const fallback = store.tenants[0]
  if (fallback && !selectedTenant.value) {
    selectedTenantId.value = fallback.id
    tenantForm.value = { ...fallback, boundDomains: [...fallback.boundDomains] }
    boundDomainsInput.value = fallback.boundDomains.join('\n')
  }
})
</script>

<template>
  <section class="tenant-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div>
            <span class="title">租戶管理</span>
            <p class="subcopy">多租戶站點、主域名、綁定域名與主題配置。</p>
          </div>
          <div class="actions">
            <el-button type="danger" :icon="Delete" :disabled="selectedTenantIds.length === 0" @click="removeSelectedTenants">
              批量刪除
            </el-button>
            <el-button type="primary" :icon="Plus" @click="startCreateTenant">新增租戶網站</el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="store.tenants"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column label="租戶" min-width="220">
          <template #default="{ row }">
            <div class="tenant-cell">
              <img :src="row.previewImage" :alt="row.name" class="tenant-thumb" />
              <div class="tenant-copy">
                <strong>{{ row.name }}</strong>
                <small>ID {{ row.id }}</small>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="domain" label="主域名" min-width="180" />
        <el-table-column label="綁定網域" min-width="240">
          <template #default="{ row }">
            <span class="bound-domains-text">
              {{ row.boundDomains.length ? row.boundDomains.join(' / ') : '—' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="theme" label="主題" width="120" />
        <el-table-column label="主品牌" min-width="140">
          <template #default="{ row }">
            {{ brands.find((brand) => brand.id === row.primaryBrandId)?.name || '未设置' }}
          </template>
        </el-table-column>
        <el-table-column label="狀態" width="110">
          <template #default="{ row }">
            <el-tag :type="row.isActive ? 'success' : 'danger'" effect="light">
              {{ row.isActive ? '啟用中' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button size="small" :icon="View" @click="openViewTenant(row.id)">查看</el-button>
              <el-button size="small" type="primary" :icon="Edit" @click="openEditTenant(row.id)">編輯</el-button>
              <el-button size="small" type="danger" :icon="Delete" @click="removeTenantById(row.id)">刪除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="isTenantModalOpen"
      :title="isCreating ? '新增租戶網站' : isViewMode ? '查看租戶網站' : '租戶編輯'"
      width="960px"
      :close-on-click-modal="false"
    >
      <div class="preview-panel">
        <div class="preview-frame">
          <img :src="tenantForm.previewImage" :alt="tenantForm.name || 'tenant preview'" />
        </div>
        <div class="logo-chip">
          <img :src="tenantForm.logoImage" :alt="tenantForm.name || 'tenant logo'" />
        </div>
      </div>

      <el-form label-width="110px">
        <div class="form-grid">
          <el-form-item label="租戶名稱">
            <el-input v-model="tenantForm.name" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item label="主域名">
            <el-select v-model="tenantForm.domain" :disabled="isViewMode" placeholder="請選擇主域名">
              <el-option
                v-for="domain in availablePrimaryDomains"
                :key="domain.id"
                :label="domain.domainName"
                :value="domain.domainName"
              />
            </el-select>
          </el-form-item>
          <el-form-item class="full" label="綁定網域">
            <el-input
              v-model="boundDomainsInput"
              :disabled="isViewMode"
              type="textarea"
              :rows="4"
              placeholder="每行一個綁定網域"
            />
            <div v-if="availableBoundDomains.length" class="domain-chip-list">
              <el-button
                v-for="domain in availableBoundDomains"
                :key="domain.id"
                size="small"
                text
                bg
                :disabled="isViewMode"
                @click="appendBoundDomain(domain.domainName)"
              >
                {{ domain.domainName }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item class="full" label="主題風格">
            <el-select v-model="tenantForm.theme" :disabled="isViewMode">
              <el-option value="ocean" label="Ocean" />
              <el-option value="sunset" label="Sunset" />
              <el-option value="forest" label="Forest" />
              <el-option value="classic" label="Classic" />
            </el-select>
          </el-form-item>
          <el-form-item class="full" label="主题预设">
            <div class="preset-list">
              <el-button
                v-for="preset in themePresets"
                :key="preset.key"
                size="small"
                :disabled="isViewMode"
                @click="applyThemePreset(preset.key)"
              >
                {{ preset.label }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="首頁模板">
            <el-select v-model="tenantForm.homeTemplate" :disabled="isViewMode">
              <el-option value="classic" label="Classic" />
              <el-option value="brand-first" label="Brand First" />
              <el-option value="campaign" label="Campaign" />
            </el-select>
          </el-form-item>
          <el-form-item label="主品牌">
            <el-select v-model="tenantForm.primaryBrandId" :disabled="isViewMode">
              <el-option :value="null" label="未设置" />
              <el-option v-for="brand in brands" :key="brand.id" :label="brand.name" :value="brand.id" />
            </el-select>
          </el-form-item>
          <el-form-item class="full" label="首页模块顺序">
            <div class="module-list">
              <div
                v-for="(moduleKey, index) in tenantForm.homeModuleOrder"
                :key="moduleKey"
                class="module-item"
              >
                <strong>{{ moduleLabels[moduleKey] ?? moduleKey }}</strong>
                <div class="module-actions">
                  <el-button size="small" :disabled="isViewMode || index === 0" @click="moveModule(index, -1)">上移</el-button>
                  <el-button
                    size="small"
                    :disabled="isViewMode || index === (tenantForm.homeModuleOrder ?? []).length - 1"
                    @click="moveModule(index, 1)"
                  >
                    下移
                  </el-button>
                </div>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="主色"><el-input v-model="tenantForm.accentColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="主色加深"><el-input v-model="tenantForm.accentStrongColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="浅色背景"><el-input v-model="tenantForm.surfaceColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="页面背景"><el-input v-model="tenantForm.pageBgColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="卡片背景"><el-input v-model="tenantForm.cardBgColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="主文字"><el-input v-model="tenantForm.textColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="弱文字"><el-input v-model="tenantForm.mutedTextColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="边框色"><el-input v-model="tenantForm.borderColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="横幅背景"><el-input v-model="tenantForm.heroBgColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item label="标签背景"><el-input v-model="tenantForm.tagBgColor" :disabled="isViewMode" type="color" /></el-form-item>
          <el-form-item class="full" label="站點預覽圖 URL">
            <el-input v-model="tenantForm.previewImage" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="Logo URL">
            <el-input v-model="tenantForm.logoImage" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="SEO 標題">
            <el-input v-model="tenantForm.seoTitle" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="SEO 描述">
            <el-input v-model="tenantForm.seoDescription" :disabled="isViewMode" type="textarea" :rows="4" />
          </el-form-item>
          <el-form-item class="full" label="站點狀態">
            <el-switch v-model="tenantForm.isActive" :disabled="isViewMode" active-text="啟用" inactive-text="停用" />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="closeTenantModal">取消</el-button>
        <el-button v-if="!isViewMode" type="primary" :loading="isSaving" @click="saveTenant">
          {{ isCreating ? '建立租戶' : '儲存租戶設定' }}
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.tenant-list {
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

.tenant-cell {
  display: grid;
  grid-template-columns: 68px 1fr;
  gap: 0.85rem;
  align-items: center;
}

.tenant-thumb {
  width: 68px;
  height: 52px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #dcdfe6;
  background: #fff;
}

.tenant-copy small,
.bound-domains-text {
  color: #909399;
  font-size: 12px;
}

.bound-domains-text {
  line-height: 1.45;
}

.preview-panel {
  display: grid;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.preview-frame {
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid #dcdfe6;
  background: #f5f7fa;
  min-height: 180px;
}

.preview-frame img {
  width: 100%;
  height: 180px;
  object-fit: cover;
}

.logo-chip {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  border: 1px solid #dcdfe6;
  background: #fff;
  overflow: hidden;
}

.logo-chip img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.form-grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: repeat(2, 1fr);
}

.form-grid :deep(.el-form-item) {
  margin-bottom: 0;
}

.form-grid .full {
  grid-column: 1 / -1;
}

.domain-chip-list,
.preset-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.module-list {
  display: grid;
  gap: 0.6rem;
  width: 100%;
}

.module-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 0.9rem;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #f5f7fa;
}

.module-actions {
  display: flex;
  gap: 0.5rem;
}

@media (max-width: 900px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .actions {
    width: 100%;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
