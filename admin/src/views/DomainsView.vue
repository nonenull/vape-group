<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Delete, Edit, Plus, Refresh, Search, Warning, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminAPI } from '@/api/admin'
import type { DomainDnsRecord, DomainRecord } from '@/data/adminMock'

const domains = ref<DomainRecord[]>([])
const loading = ref(false)
const isSaving = ref(false)
const isCreating = ref(false)
const isViewMode = ref(false)
const isDomainModalOpen = ref(false)
const isDnsModalOpen = ref(false)
const selectedDomainId = ref(0)
const selectedDomainIds = ref<number[]>([])
const searchKeyword = ref('')
const syncing = ref(false)
const checkingDns = ref(false)
const dnsRecords = ref<DomainDnsRecord[]>([])

const emptyDomain = (): DomainRecord => ({
  id: 0,
  domainName: '',
  registrar: 'manual',
  expireDate: null,
  dnsRecords: [],
  isBlocked: false,
  lastCheckIp: null,
  lastCheckedAt: null,
})

const domainForm = ref<DomainRecord>(emptyDomain())

const selectedDomain = computed(() =>
  domains.value.find((item) => item.id === selectedDomainId.value) ?? null,
)

function cloneDomain(domain: DomainRecord): DomainRecord {
  return { ...domain, dnsRecords: [...(domain.dnsRecords ?? [])] }
}

async function loadDomains() {
  loading.value = true
  try {
    domains.value = await adminAPI.getDomains(searchKeyword.value.trim())
  } finally {
    loading.value = false
  }
}

watch(domains, (items) => {
  const existingIds = new Set(items.map((item) => item.id))
  selectedDomainIds.value = selectedDomainIds.value.filter((id) => existingIds.has(id))

  if (isCreating.value) return

  const fallback = items[0]
  if (!selectedDomainId.value || !items.some((item) => item.id === selectedDomainId.value)) {
    selectedDomainId.value = fallback?.id ?? 0
    domainForm.value = fallback ? cloneDomain(fallback) : emptyDomain()
  }
}, { immediate: true })

watch(selectedDomainId, (domainId) => {
  if (isCreating.value) return
  const domain = domains.value.find((item) => item.id === domainId)
  if (domain) {
    domainForm.value = cloneDomain(domain)
  }
})

function startCreateDomain() {
  isCreating.value = true
  isViewMode.value = false
  selectedDomainId.value = 0
  domainForm.value = emptyDomain()
  isDomainModalOpen.value = true
}

function openViewDomain(domainId: number) {
  const domain = domains.value.find((item) => item.id === domainId)
  if (!domain) return
  isCreating.value = false
  isViewMode.value = true
  selectedDomainId.value = domainId
  domainForm.value = cloneDomain(domain)
  isDomainModalOpen.value = true
}

function openEditDomain(domainId: number) {
  const domain = domains.value.find((item) => item.id === domainId)
  if (!domain) return
  isCreating.value = false
  isViewMode.value = false
  selectedDomainId.value = domainId
  domainForm.value = cloneDomain(domain)
  isDomainModalOpen.value = true
}

function closeDomainModal() {
  isDomainModalOpen.value = false
  isCreating.value = false
  isViewMode.value = false

  if (selectedDomain.value) {
    domainForm.value = cloneDomain(selectedDomain.value)
    return
  }

  domainForm.value = emptyDomain()
}

function handleSelectionChange(rows: DomainRecord[]) {
  selectedDomainIds.value = rows.map((row) => row.id)
}

async function syncDomains() {
  syncing.value = true
  try {
    const result = await adminAPI.syncDomains()
    await loadDomains()
    ElMessage.success(`同步完成：新增 ${result.created}，更新 ${result.updated}`)
  } catch (error) {
    console.error('同步域名失敗:', error)
    ElMessage.error('同步域名失敗: ' + (error as Error).message)
  } finally {
    syncing.value = false
  }
}

async function checkDomainDns() {
  checkingDns.value = true
  try {
    const result = await adminAPI.checkDomainDns()
    await loadDomains()
    ElMessage.success(`DNS 檢測完成：檢查 ${result.checked} 個，封殺 ${result.blocked} 個`)
  } catch (error) {
    console.error('DNS 檢測失敗:', error)
    ElMessage.error('DNS 檢測失敗: ' + (error as Error).message)
  } finally {
    checkingDns.value = false
  }
}

async function openDnsModal(domainId: number) {
  const domain = domains.value.find((item) => item.id === domainId)
  if (!domain) return
  selectedDomainId.value = domainId
  try {
    dnsRecords.value = await adminAPI.getDomainDnsRecords(domainId)
    isDnsModalOpen.value = true
  } catch (error) {
    console.error('讀取 DNS 記錄失敗:', error)
    ElMessage.error('讀取 DNS 記錄失敗: ' + (error as Error).message)
  }
}

async function saveDnsRecords() {
  if (!selectedDomainId.value) return
  try {
    await adminAPI.updateDomainDnsRecords(selectedDomainId.value, dnsRecords.value)
    isDnsModalOpen.value = false
    await loadDomains()
    ElMessage.success('DNS 記錄已成功保存')
  } catch (error) {
    console.error('保存 DNS 記錄失敗:', error)
    ElMessage.error('保存 DNS 記錄失敗: ' + (error as Error).message)
  }
}

function addDnsRecord() {
  dnsRecords.value.push({
    type: 'A',
    name: '@',
    data: '',
    ttl: 600,
  })
}

function removeDnsRecord(index: number) {
  dnsRecords.value.splice(index, 1)
}

async function saveDomain() {
  if (!domainForm.value.domainName.trim()) {
    ElMessage.warning('請先填寫域名')
    return
  }

  isSaving.value = true
  try {
    if (isCreating.value || domainForm.value.id === 0) {
      const created = await adminAPI.createDomain({
        domainName: domainForm.value.domainName.trim(),
        registrar: domainForm.value.registrar.trim() || 'manual',
        expireDate: domainForm.value.expireDate || null,
        dnsRecords: [],
        isBlocked: false,
        lastCheckIp: null,
        lastCheckedAt: null,
      })
      await loadDomains()
      isCreating.value = false
      isViewMode.value = false
      isDomainModalOpen.value = false
      selectedDomainId.value = created.id
      domainForm.value = cloneDomain(created)
      ElMessage.success('域名已成功建立')
      return
    }

    const updated = await adminAPI.updateDomain(domainForm.value.id, {
      domainName: domainForm.value.domainName.trim(),
      registrar: domainForm.value.registrar.trim() || 'manual',
      expireDate: domainForm.value.expireDate || null,
      dnsRecords: domainForm.value.dnsRecords ?? [],
      isBlocked: domainForm.value.isBlocked ?? false,
      lastCheckIp: domainForm.value.lastCheckIp ?? null,
      lastCheckedAt: domainForm.value.lastCheckedAt ?? null,
    })
    await loadDomains()
    selectedDomainId.value = updated.id
    domainForm.value = cloneDomain(updated)
    isViewMode.value = false
    isDomainModalOpen.value = false
    ElMessage.success('域名已成功儲存')
  } catch (error) {
    console.error('保存域名失敗:', error)
    ElMessage.error('保存域名失敗: ' + (error as Error).message)
  } finally {
    isSaving.value = false
  }
}

async function removeDomainById(domainId: number) {
  const domain = domains.value.find((item) => item.id === domainId)
  if (!domain) return
  try {
    await ElMessageBox.confirm(`確定要刪除域名「${domain.domainName}」嗎？`, '刪除域名', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    await adminAPI.deleteDomain(domain.id)
    await loadDomains()
    selectedDomainIds.value = selectedDomainIds.value.filter((id) => id !== domain.id)
    if (selectedDomainId.value === domain.id) {
      selectedDomainId.value = domains.value[0]?.id ?? 0
      domainForm.value = domains.value[0] ? cloneDomain(domains.value[0]) : emptyDomain()
      isDomainModalOpen.value = false
      isViewMode.value = false
      isCreating.value = false
    }
    ElMessage.success('域名已成功刪除')
  } catch (error) {
    console.error('刪除域名失敗:', error)
    ElMessage.error('刪除域名失敗: ' + (error as Error).message)
  }
}

async function removeSelectedDomains() {
  if (!selectedDomainIds.value.length) {
    ElMessage.warning('請先勾選至少一個域名')
    return
  }
  try {
    await ElMessageBox.confirm(`確定要刪除選中的 ${selectedDomainIds.value.length} 個域名嗎？`, '批量刪除域名', {
      type: 'warning',
      confirmButtonText: '刪除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  try {
    for (const domainId of [...selectedDomainIds.value]) {
      await adminAPI.deleteDomain(domainId)
    }
    await loadDomains()
    selectedDomainIds.value = []
    selectedDomainId.value = domains.value[0]?.id ?? 0
    domainForm.value = domains.value[0] ? cloneDomain(domains.value[0]) : emptyDomain()
    isDomainModalOpen.value = false
    isViewMode.value = false
    isCreating.value = false
    ElMessage.success('已成功刪除所選域名')
  } catch (error) {
    console.error('批量刪除域名失敗:', error)
    ElMessage.error('批量刪除域名失敗: ' + (error as Error).message)
  }
}

function formatDate(value?: string | null) {
  return value || '—'
}

onMounted(() => {
  loadDomains()
})
</script>

<template>
  <section class="domain-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">域名管理</span>
          <div class="actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索域名或注册商..."
              :prefix-icon="Search"
              clearable
              style="width: 260px"
              @clear="loadDomains"
              @keyup.enter="loadDomains"
            />
            <el-button :icon="Search" @click="loadDomains">搜索</el-button>
            <el-button :icon="Warning" :loading="checkingDns" @click="checkDomainDns">DNS 检测</el-button>
            <el-button :icon="Refresh" :loading="syncing" type="primary" @click="syncDomains">GoDaddy 同步</el-button>
            <el-button type="danger" :icon="Delete" :disabled="selectedDomainIds.length === 0" @click="removeSelectedDomains">
              批量刪除
            </el-button>
            <el-button type="success" :icon="Plus" @click="startCreateDomain">新增域名</el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="domains"
        v-loading="loading"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column label="域名" min-width="220">
          <template #default="{ row }">
            <div class="domain-name-cell">
              <strong>{{ row.domainName }}</strong>
              <small>ID {{ row.id }}</small>
              <el-tag v-if="row.isBlocked" type="danger" size="small">疑似封杀</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="registrar" label="注册商" width="120" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.isBlocked ? 'danger' : 'success'" effect="light">
              {{ row.isBlocked ? '疑似封杀' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="过期日期" width="140">
          <template #default="{ row }">
            {{ formatDate(row.expireDate) }}
          </template>
        </el-table-column>
        <el-table-column label="检测IP" min-width="160">
          <template #default="{ row }">
            <span v-if="row.lastCheckIp">{{ row.lastCheckIp }}</span>
            <span v-else class="text-muted">未检测</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button size="small" :icon="View" @click="openDnsModal(row.id)">DNS</el-button>
              <el-button size="small" :icon="View" @click="openViewDomain(row.id)">查看</el-button>
              <el-button size="small" type="primary" :icon="Edit" @click="openEditDomain(row.id)">编辑</el-button>
              <el-button size="small" type="danger" :icon="Delete" @click="removeDomainById(row.id)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="isDomainModalOpen"
      :title="isCreating ? '新增域名' : isViewMode ? '查看域名' : '编辑域名'"
      width="640px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="域名">
          <el-input v-model="domainForm.domainName" :disabled="isViewMode" placeholder="example.com" />
        </el-form-item>
        <el-form-item label="注册商">
          <el-input v-model="domainForm.registrar" :disabled="isViewMode" placeholder="godaddy / manual" />
        </el-form-item>
        <el-form-item label="过期日期">
          <el-date-picker
            v-model="domainForm.expireDate"
            :disabled="isViewMode"
            type="date"
            value-format="YYYY-MM-DD"
            format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDomainModal">取消</el-button>
        <el-button v-if="!isViewMode" type="primary" :loading="isSaving" @click="saveDomain">
          {{ isCreating ? '建立域名' : '儲存域名' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="isDnsModalOpen"
      :title="`DNS 记录 - ${selectedDomain?.domainName || ''}`"
      width="900px"
      :close-on-click-modal="false"
    >
      <div class="dns-dialog-header">
        <el-button type="primary" :icon="Plus" size="small" @click="addDnsRecord">添加记录</el-button>
        <el-button type="success" size="small" @click="saveDnsRecords">保存修改</el-button>
      </div>
      <el-table :data="dnsRecords" style="margin-top: 12px">
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-input v-model="row.type" />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="140">
          <template #default="{ row }">
            <el-input v-model="row.name" />
          </template>
        </el-table-column>
        <el-table-column prop="data" label="值" min-width="220">
          <template #default="{ row }">
            <el-input v-model="row.data" />
          </template>
        </el-table-column>
        <el-table-column prop="ttl" label="TTL" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.ttl" :min="60" :step="60" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ $index }">
            <el-button size="small" type="danger" :icon="Delete" @click="removeDnsRecord($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </section>
</template>

<style scoped>
.domain-list {
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

.domain-name-cell {
  display: grid;
  gap: 0.2rem;
}

.domain-name-cell small,
.text-muted {
  color: #909399;
  font-size: 12px;
}

.dns-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
