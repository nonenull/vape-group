<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Delete, Edit, Plus, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminAPI } from '@/api/admin'
import { useAdminStore } from '@/stores/admin'
import type { DomainRecord, HomeSectionRecord, TenantRecord } from '@/data/adminMock'

const store = useAdminStore()
const brands = computed(() => store.getBrands())
const domains = ref<DomainRecord[]>([])

const emptyTenant = (): TenantRecord => ({
  id: 0,
  domain: '',
  boundDomains: [],
  npmProxyHostId: null,
  name: '',
  isActive: true,
  theme: '',
  homeTemplate: 'classic',
  homeModuleOrder: ['brand-categories', 'products', 'brands', 'categories'],
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
  heroTitle: '',
  tagline: '',
  announcement: '',
  supportText: '',
  seoTitle: '',
  seoDescription: '',
  homeBanner: {
    enabled: false,
    title: '',
    subtitle: '',
    image: '',
    link: '',
    buttonText: '',
  },
  homeSections: [
    { id: 'brand-categories', type: 'brand_categories', enabled: true, title: '品牌分類', limit: 3 },
    { id: 'products', type: 'hot_products', enabled: true, title: '最近熱賣', limit: 8 },
    { id: 'brands', type: 'other_brands', enabled: true, title: '其他品牌', limit: 6 },
    { id: 'categories', type: 'featured_categories', enabled: true, title: '熱門分類', limit: 6 },
  ],
})

const selectedTenantId = ref(store.tenants[0]?.id ?? 0)
const tenantForm = ref<TenantRecord>(store.tenants[0] ? cloneTenantRecord(store.tenants[0]) : emptyTenant())
const isCreating = ref(false)
const isTenantModalOpen = ref(false)
const isViewMode = ref(false)
const isSaving = ref(false)
const selectedTenantIds = ref<number[]>([])
const selectedDomainToBind = ref('')
const isBindingDomain = ref(false)
const domainBeingRemoved = ref('')
const domainBeingPromoted = ref('')

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
  'brand-categories': '品牌分類',
}

const sectionTypeOptions = [
  { value: 'hot_products', label: '最近熱賣' },
  { value: 'brand_categories', label: '品牌分類' },
  { value: 'other_brands', label: '其他品牌' },
  { value: 'featured_categories', label: '熱門分類' },
] as const

function cloneTenantRecord(input: TenantRecord): TenantRecord {
  return {
    ...input,
    boundDomains: [...input.boundDomains],
    homeBanner: input.homeBanner ? { ...input.homeBanner } : undefined,
    homeSections: input.homeSections?.map((item) => ({ ...item })) ?? [],
  }
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

const availableDomainsToBind = computed(() => {
  const selected = new Set(
    [
      tenantForm.value.domain,
      ...(tenantForm.value.boundDomains ?? []),
    ]
      .map((item) => item.trim().toLowerCase())
      .filter(Boolean),
  )

  return availableBoundDomains.value.filter((domain) => !selected.has(domain.domainName.trim().toLowerCase()))
})

const domainManagementRows = computed(() => [
  {
    domain: tenantForm.value.domain,
    role: '主域名',
    isPrimary: true,
  },
  ...tenantForm.value.boundDomains.map((domain) => ({
    domain,
    role: '绑定域名',
    isPrimary: false,
  })),
])

const boundDomainOptions = computed(() => {
  const optionMap = new Map<string, DomainRecord>()

  for (const domain of availableBoundDomains.value) {
    optionMap.set(domain.domainName.trim().toLowerCase(), domain)
  }

  for (const domainName of tenantForm.value.boundDomains) {
    const normalized = domainName.trim().toLowerCase()
    if (!normalized || optionMap.has(normalized)) continue
    optionMap.set(normalized, {
      id: -optionMap.size - 1,
      domainName: normalized,
      registrar: 'manual',
    })
  }

  return Array.from(optionMap.values())
})

function normalizeBoundDomains(values: string[]) {
  const primaryDomain = tenantForm.value.domain.trim().toLowerCase()
  const seen = new Set<string>()

  return values
    .map((item) => item.trim().toLowerCase())
    .filter((item) => {
      if (!item || item === primaryDomain || seen.has(item)) {
        return false
      }
      seen.add(item)
      return true
    })
}

watch(
  selectedTenantId,
  (tenantId) => {
    if (isCreating.value) return
    const tenant = store.tenants.find((item) => item.id === tenantId)
    if (tenant) {
      tenantForm.value = cloneTenantRecord(tenant)
    }
  },
  { immediate: true },
)

watch(
  () => tenantForm.value.domain,
  () => {
    tenantForm.value.boundDomains = normalizeBoundDomains(tenantForm.value.boundDomains)
  },
)

function startCreateTenant() {
  isCreating.value = true
  isViewMode.value = false
  selectedTenantId.value = 0
  tenantForm.value = emptyTenant()
  isTenantModalOpen.value = true
}

function openEditTenant(tenantId: number) {
  isCreating.value = false
  isViewMode.value = false
  selectedTenantId.value = tenantId
  const tenant = store.tenants.find((item) => item.id === tenantId)
  if (tenant) {
    tenantForm.value = cloneTenantRecord(tenant)
  }
  isTenantModalOpen.value = true
}

function openViewTenant(tenantId: number) {
  isCreating.value = false
  isViewMode.value = true
  selectedTenantId.value = tenantId
  const tenant = store.tenants.find((item) => item.id === tenantId)
  if (tenant) {
    tenantForm.value = cloneTenantRecord(tenant)
  }
  isTenantModalOpen.value = true
}

function closeTenantModal() {
  isTenantModalOpen.value = false
  isCreating.value = false
  isViewMode.value = false
  selectedDomainToBind.value = ''
  const fallback = selectedTenant.value ?? store.tenants[0] ?? null
  if (fallback) {
    selectedTenantId.value = fallback.id
    tenantForm.value = cloneTenantRecord(fallback)
    return
  }
  selectedTenantId.value = 0
  tenantForm.value = emptyTenant()
}

function handleSelectionChange(rows: TenantRecord[]) {
  selectedTenantIds.value = rows.map((row) => row.id)
}

function syncTenantForm(updated: TenantRecord) {
  if (selectedTenantId.value === updated.id || tenantForm.value.id === updated.id) {
    tenantForm.value = cloneTenantRecord(updated)
    selectedTenantId.value = updated.id
  }
}

function showNpmSyncMessage(action: string, npmResult: any) {
  if (!npmResult) {
    ElMessage.success(`${action}成功`)
    return
  }

  if (npmResult.status === 'success' && npmResult.npm_updated) {
    const domains = Array.isArray(npmResult.updated_domains) ? npmResult.updated_domains.join(', ') : ''
    ElMessage.success(`${action}成功，NPM 已同步${domains ? `：${domains}` : ''}`)
    return
  }

  if (npmResult.status === 'warning') {
    ElMessage.warning(`${action}成功，但 NPM 未更新：${npmResult.message || '未找到代理主机'}`)
    return
  }

  ElMessage.warning(`${action}成功，但 NPM 同步失败：${npmResult.message || '未知错误'}`)
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

  tenantForm.value.boundDomains = normalizeBoundDomains(tenantForm.value.boundDomains)
  tenantForm.value.homeModuleOrder = (tenantForm.value.homeModuleOrder ?? []).filter(Boolean)
  tenantForm.value.homeSections = (tenantForm.value.homeSections ?? [])
    .map((item, index): HomeSectionRecord => ({
      id: item.id?.trim() || `${item.type || 'section'}-${index + 1}`,
      type: item.type?.trim() || 'hot_products',
      enabled: item.enabled !== false,
      title: item.title?.trim() || '',
      limit: Number.isFinite(Number(item.limit)) ? Math.max(0, Number(item.limit)) : 0,
    }))
    .filter((item) => item.type)

  isSaving.value = true
  try {
    if (isCreating.value || tenantForm.value.id === 0) {
      const created = await store.createTenant({
        domain: tenantForm.value.domain,
        boundDomains: tenantForm.value.boundDomains,
        npmProxyHostId: tenantForm.value.npmProxyHostId,
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
        homeBanner: tenantForm.value.homeBanner,
        homeSections: tenantForm.value.homeSections,
      })
      isCreating.value = false
      isTenantModalOpen.value = false
      selectedTenantId.value = created.id
      tenantForm.value = cloneTenantRecord(created)
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

async function bindDomainToTenant() {
  if (isCreating.value || !tenantForm.value.id) {
    ElMessage.warning('请先创建租户，再绑定域名')
    return
  }
  if (!selectedDomainToBind.value) {
    ElMessage.warning('请选择要绑定的域名')
    return
  }

  isBindingDomain.value = true
  try {
    const result = await store.addTenantDomain(tenantForm.value.id, selectedDomainToBind.value)
    syncTenantForm(result.tenant)
    selectedDomainToBind.value = ''
    showNpmSyncMessage('绑定域名', result.npmResult)
  } catch (error) {
    console.error('绑定域名失败:', error)
    ElMessage.error('绑定域名失败: ' + (error as Error).message)
  } finally {
    isBindingDomain.value = false
  }
}

async function removeBoundDomain(domain: string) {
  if (!tenantForm.value.id) return

  try {
    await ElMessageBox.confirm(`确定要删除绑定域名 "${domain}" 吗？`, '删除域名', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  domainBeingRemoved.value = domain
  try {
    const result = await store.removeTenantDomain(tenantForm.value.id, domain)
    syncTenantForm(result.tenant)
    showNpmSyncMessage('删除绑定域名', result.npmResult)
  } catch (error) {
    console.error('删除绑定域名失败:', error)
    ElMessage.error('删除绑定域名失败: ' + (error as Error).message)
  } finally {
    domainBeingRemoved.value = ''
  }
}

async function promoteDomainToPrimary(domain: string) {
  if (!tenantForm.value.id) return

  try {
    await ElMessageBox.confirm(`确定将 "${domain}" 设为主域名吗？`, '设为主域名', {
      type: 'info',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  domainBeingPromoted.value = domain
  try {
    const result = await store.setTenantPrimaryDomain(tenantForm.value.id, domain)
    syncTenantForm(result.tenant)
    ElMessage.success(`主域名已切换为 ${result.tenant.domain}`)
  } catch (error) {
    console.error('设置主域名失败:', error)
    ElMessage.error('设置主域名失败: ' + (error as Error).message)
  } finally {
    domainBeingPromoted.value = ''
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
      tenantForm.value = fallback ? cloneTenantRecord(fallback) : emptyTenant()
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
    tenantForm.value = fallback ? cloneTenantRecord(fallback) : emptyTenant()
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

function ensureHomeSections() {
  if (!Array.isArray(tenantForm.value.homeSections) || tenantForm.value.homeSections.length === 0) {
    tenantForm.value.homeSections = emptyTenant().homeSections?.map((item) => ({ ...item })) ?? []
  }
}

function moveSection(index: number, direction: -1 | 1) {
  ensureHomeSections()
  const sections = [...(tenantForm.value.homeSections ?? [])]
  const target = index + direction
  if (target < 0 || target >= sections.length) return
  const currentValue = sections[index]
  const targetValue = sections[target]
  if (!currentValue || !targetValue) return
  sections[index] = targetValue
  sections[target] = currentValue
  tenantForm.value.homeSections = sections
}

function addSection() {
  ensureHomeSections()
  const nextIndex = (tenantForm.value.homeSections?.length ?? 0) + 1
  tenantForm.value.homeSections = [
    ...(tenantForm.value.homeSections ?? []),
    {
      id: `section-${nextIndex}`,
      type: 'brand_categories',
      enabled: true,
      title: `首頁模組 ${nextIndex}`,
      limit: 3,
    },
  ]
}

function removeSection(index: number) {
  ensureHomeSections()
  tenantForm.value.homeSections = (tenantForm.value.homeSections ?? []).filter((_, itemIndex) => itemIndex !== index)
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
    tenantForm.value = cloneTenantRecord(fallback)
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
      <el-form label-width="110px">
        <div class="form-grid">
          <el-form-item label="租戶名稱">
            <el-input v-model="tenantForm.name" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item label="主域名">
            <el-select
              v-model="tenantForm.domain"
              :disabled="isViewMode"
              filterable
              clearable
              placeholder="請選擇主域名"
            >
              <el-option
                v-for="domain in availablePrimaryDomains"
                :key="domain.id"
                :label="domain.domainName"
                :value="domain.domainName"
              />
            </el-select>
          </el-form-item>
          <el-form-item v-if="isCreating" class="full" label="綁定網域">
            <el-select
              v-model="tenantForm.boundDomains"
              :disabled="isViewMode"
              multiple
              filterable
              clearable
              default-first-option
              placeholder="請選擇綁定網域"
            >
              <el-option
                v-for="domain in boundDomainOptions"
                :key="domain.id"
                :label="domain.domainName"
                :value="domain.domainName"
                :disabled="domain.domainName === tenantForm.domain"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="NPM Host ID">
            <el-input-number
              v-model="tenantForm.npmProxyHostId"
              :disabled="isViewMode"
              :min="1"
              :step="1"
              controls-position="right"
              placeholder="可选"
            />
          </el-form-item>
          <el-form-item v-if="!isCreating" class="full" label="域名管理">
            <div class="domain-manager">
              <div class="domain-manager-toolbar">
                <el-select
                  v-model="selectedDomainToBind"
                  :disabled="isViewMode || isBindingDomain"
                  filterable
                  clearable
                  placeholder="选择域名库中的域名进行绑定"
                >
                  <el-option
                    v-for="domain in availableDomainsToBind"
                    :key="domain.id"
                    :label="domain.domainName"
                    :value="domain.domainName"
                  />
                </el-select>
                <el-button
                  type="primary"
                  :loading="isBindingDomain"
                  :disabled="isViewMode || !selectedDomainToBind"
                  @click="bindDomainToTenant"
                >
                  绑定域名
                </el-button>
              </div>
              <el-table
                :data="domainManagementRows"
                border
                size="small"
                class="domain-manager-table"
                empty-text="暂无域名"
              >
                <el-table-column prop="domain" label="域名" min-width="280">
                  <template #default="{ row }">
                    <div class="domain-cell">
                      <strong>{{ row.domain }}</strong>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column prop="role" label="类型" width="120">
                  <template #default="{ row }">
                    <el-tag :type="row.isPrimary ? 'success' : 'info'" effect="light" size="small">
                      {{ row.role }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" min-width="220">
                  <template #default="{ row }">
                    <div v-if="!row.isPrimary" class="domain-actions">
                      <el-button
                        size="small"
                        :loading="domainBeingPromoted === row.domain"
                        :disabled="isViewMode"
                        @click="promoteDomainToPrimary(row.domain)"
                      >
                        设为主域名
                      </el-button>
                      <el-button
                        size="small"
                        type="danger"
                        :loading="domainBeingRemoved === row.domain"
                        :disabled="isViewMode"
                        @click="removeBoundDomain(row.domain)"
                      >
                        删除
                      </el-button>
                    </div>
                    <span v-else class="domain-manager-empty">当前生效中</span>
                  </template>
                </el-table-column>
              </el-table>
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
          <el-form-item class="full" label="Banner 設定">
            <div class="full-stack">
              <el-switch
                v-model="tenantForm.homeBanner!.enabled"
                :disabled="isViewMode"
                active-text="啟用 Banner"
                inactive-text="停用 Banner"
              />
              <div class="section-grid">
                <el-input v-model="tenantForm.homeBanner!.title" :disabled="isViewMode" placeholder="Banner 標題" />
                <el-input v-model="tenantForm.homeBanner!.buttonText" :disabled="isViewMode" placeholder="按鈕文字" />
                <el-input v-model="tenantForm.homeBanner!.subtitle" :disabled="isViewMode" placeholder="Banner 副標題" />
                <el-input v-model="tenantForm.homeBanner!.link" :disabled="isViewMode" placeholder="跳轉連結，如 /products" />
              </div>
              <el-input v-model="tenantForm.homeBanner!.image" :disabled="isViewMode" placeholder="Banner 圖片 URL" />
            </div>
          </el-form-item>
          <el-form-item class="full" label="旧首页模块顺序">
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
          <el-form-item class="full" label="首頁模組配置">
            <div class="module-list">
              <div
                v-for="(section, index) in tenantForm.homeSections"
                :key="section.id"
                class="module-item module-editor"
              >
                <div class="module-editor-grid">
                  <el-input v-model="section.title" :disabled="isViewMode" placeholder="模組標題" />
                  <el-select v-model="section.type" :disabled="isViewMode">
                    <el-option
                      v-for="option in sectionTypeOptions"
                      :key="option.value"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                  <el-input-number
                    v-model="section.limit"
                    :disabled="isViewMode"
                    :min="0"
                    :max="24"
                    controls-position="right"
                  />
                  <el-switch v-model="section.enabled" :disabled="isViewMode" active-text="啟用" inactive-text="停用" />
                </div>
                <div class="module-actions">
                  <el-button size="small" :disabled="isViewMode || index === 0" @click="moveSection(index, -1)">上移</el-button>
                  <el-button
                    size="small"
                    :disabled="isViewMode || index === (tenantForm.homeSections ?? []).length - 1"
                    @click="moveSection(index, 1)"
                  >
                    下移
                  </el-button>
                  <el-button size="small" type="danger" :disabled="isViewMode" @click="removeSection(index)">刪除</el-button>
                </div>
              </div>
              <el-button v-if="!isViewMode" plain @click="addSection">新增首頁模組</el-button>
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
            <div class="preview-field">
              <el-input v-model="tenantForm.previewImage" :disabled="isViewMode" />
              <div class="preview-panel inline">
                <div class="preview-frame">
                  <img :src="tenantForm.previewImage" :alt="tenantForm.name || 'tenant preview'" />
                </div>
                <div class="logo-chip">
                  <img :src="tenantForm.logoImage" :alt="tenantForm.name || 'tenant logo'" />
                </div>
              </div>
            </div>
          </el-form-item>
          <el-form-item class="full" label="Logo URL">
            <el-input v-model="tenantForm.logoImage" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="首页主标题">
            <el-input v-model="tenantForm.heroTitle" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="站點標語">
            <el-input v-model="tenantForm.tagline" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="公告文案">
            <el-input v-model="tenantForm.announcement" :disabled="isViewMode" />
          </el-form-item>
          <el-form-item class="full" label="客服提示">
            <el-input v-model="tenantForm.supportText" :disabled="isViewMode" />
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
}

.preview-panel.inline {
  grid-template-columns: minmax(180px, 280px) 64px;
  align-items: start;
  flex: 0 0 auto;
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

.preview-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  width: 100%;
  align-items: start;
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

.full-stack {
  display: grid;
  gap: 0.75rem;
  width: 100%;
}

.domain-manager {
  display: grid;
  gap: 0.75rem;
  width: 100%;
}

.domain-manager-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
}

.domain-manager-table {
  width: 100%;
}

.domain-cell strong,
.domain-manager-empty {
  font-size: 12px;
}

.domain-manager-empty {
  color: #909399;
}

.domain-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.section-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
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

.module-editor {
  align-items: stretch;
}

.module-editor-grid {
  display: grid;
  gap: 0.75rem;
  flex: 1;
  grid-template-columns: minmax(0, 1.3fr) minmax(0, 1fr) 140px 130px;
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

  .section-grid,
  .module-editor-grid,
  .domain-manager-toolbar,
  .preview-field,
  .preview-panel.inline {
    grid-template-columns: 1fr;
  }

  .preview-panel.inline {
    justify-items: start;
  }
}
</style>
