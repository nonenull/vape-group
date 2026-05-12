<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAdminStore } from '@/stores/admin'
import type { TenantRecord } from '@/data/adminMock'

const store = useAdminStore()

const emptyTenant = (): TenantRecord => ({
  id: 0,
  domain: '',
  boundDomains: [],
  name: '',
  isActive: true,
  theme: '',
  previewImage: '/src/assets/logo.svg',
  logoImage: '/src/assets/logo.svg',
  seoTitle: '',
  seoDescription: '',
})

const selectedTenantId = ref(store.tenants[0]?.id ?? 0)
const tenantForm = ref<TenantRecord>(store.tenants[0] ? { ...store.tenants[0] } : emptyTenant())
const isCreating = ref(false)
const boundDomainsInput = ref('')
const isSaving = ref(false)

const selectedTenant = computed(() => store.tenants.find((item) => item.id === selectedTenantId.value) ?? null)

watch(
  selectedTenantId,
  (tenantId) => {
    if (isCreating.value) {
      return
    }
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
  selectedTenantId.value = 0
  tenantForm.value = emptyTenant()
  boundDomainsInput.value = ''
}

function selectTenant(tenantId: number) {
  isCreating.value = false
  selectedTenantId.value = tenantId
}

async function saveTenant() {
  if (!tenantForm.value.name.trim()) {
    alert('請先填寫租戶名稱')
    return
  }
  if (!tenantForm.value.domain.trim()) {
    alert('請先填寫主域名')
    return
  }

  tenantForm.value.boundDomains = boundDomainsInput.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)

  isSaving.value = true
  try {
    if (isCreating.value || tenantForm.value.id === 0) {
      const created = await store.createTenant({
        domain: tenantForm.value.domain,
        boundDomains: tenantForm.value.boundDomains,
        name: tenantForm.value.name,
        isActive: tenantForm.value.isActive,
        theme: tenantForm.value.theme,
        previewImage: tenantForm.value.previewImage,
        logoImage: tenantForm.value.logoImage,
        accentColor: tenantForm.value.accentColor,
        surfaceColor: tenantForm.value.surfaceColor,
        heroTitle: tenantForm.value.heroTitle,
        tagline: tenantForm.value.tagline,
        announcement: tenantForm.value.announcement,
        supportText: tenantForm.value.supportText,
        seoTitle: tenantForm.value.seoTitle,
        seoDescription: tenantForm.value.seoDescription,
      })
      isCreating.value = false
      selectedTenantId.value = created.id
      tenantForm.value = { ...created, boundDomains: [...created.boundDomains] }
      boundDomainsInput.value = created.boundDomains.join('\n')
      alert('租戶已成功建立')
      return
    }

    await store.updateTenant({ ...tenantForm.value, boundDomains: [...tenantForm.value.boundDomains] })
    alert('租戶設定已成功儲存')
  } catch (error) {
    console.error('保存租戶失敗:', error)
    alert('保存租戶失敗: ' + (error as Error).message)
  } finally {
    isSaving.value = false
  }
}

async function removeTenant() {
  if (!selectedTenant.value) {
    return
  }
  if (!confirm('確定要刪除此租戶嗎？')) {
    return
  }
  try {
    const deletingId = selectedTenant.value.id
    await store.deleteTenant(deletingId)
    isCreating.value = false
    const fallback = store.tenants[0]
    selectedTenantId.value = fallback?.id ?? 0
    tenantForm.value = fallback ? { ...fallback, boundDomains: [...fallback.boundDomains] } : emptyTenant()
    boundDomainsInput.value = fallback?.boundDomains.join('\n') ?? ''
    alert('租戶已成功刪除')
  } catch (error) {
    console.error('刪除租戶失敗:', error)
    alert('刪除租戶失敗: ' + (error as Error).message)
  }
}

onMounted(async () => {
  if (!store.tenants.length) {
    await store.fetchTenants()
  }
  const fallback = store.tenants[0]
  if (fallback && !selectedTenant.value) {
    selectedTenantId.value = fallback.id
    tenantForm.value = { ...fallback, boundDomains: [...fallback.boundDomains] }
    boundDomainsInput.value = fallback.boundDomains.join('\n')
  }
})
</script>

<template>
  <section class="tenants-page">
    <div class="page-heading">
      <div>
        <p class="label">Tenant Center</p>
        <h2>租戶網站 CRUD 與預覽圖</h2>
        <p class="subcopy">現在可直接新增、修改、刪除租戶網站，並維護站點預覽圖與 Logo，方便做多租戶店鋪管理。</p>
      </div>
      <button class="primary" type="button" @click="startCreateTenant">新增租戶網站</button>
    </div>

    <div class="tenant-layout">
      <article class="tenant-list-card">
        <h3>租戶清單</h3>
        <div class="tenant-list">
          <button
            v-for="tenant in store.tenants"
            :key="tenant.id"
            type="button"
            class="tenant-item"
            :class="{ selected: tenant.id === selectedTenantId && !isCreating }"
            @click="selectTenant(tenant.id)"
          >
            <img :src="tenant.previewImage" :alt="tenant.name" class="tenant-thumb" />
            <div class="tenant-copy">
              <strong>{{ tenant.name }}</strong>
              <p>{{ tenant.domain }}</p>
            </div>
            <span :class="['pill', tenant.isActive ? 'active' : 'inactive']">
              {{ tenant.isActive ? '啟用中' : '停用' }}
            </span>
          </button>
        </div>
      </article>

      <article class="editor-card">
        <div class="card-heading">
          <h3>{{ isCreating ? '新增租戶網站' : '租戶編輯' }}</h3>
          <small v-if="!isCreating">ID {{ tenantForm.id }}</small>
        </div>

        <div class="preview-panel">
          <div class="preview-frame">
            <img :src="tenantForm.previewImage" :alt="tenantForm.name || 'tenant preview'" />
          </div>
          <div class="logo-chip">
            <img :src="tenantForm.logoImage" :alt="tenantForm.name || 'tenant logo'" />
          </div>
        </div>

        <div class="form-grid">
          <label>
            <span>租戶名稱</span>
            <input v-model="tenantForm.name" />
          </label>
          <label>
            <span>主域名</span>
            <input v-model="tenantForm.domain" placeholder="主域名，例如 tenant1.localhost" />
          </label>
          <label class="full">
            <span>綁定網域（每行一個）</span>
            <textarea
              v-model="boundDomainsInput"
              rows="4"
              placeholder="例如&#10;www.tenant1.localhost&#10;shop.brand.com"
            ></textarea>
          </label>
          <label class="full">
            <span>主題風格</span>
            <input v-model="tenantForm.theme" />
          </label>
          <label class="full">
            <span>站點預覽圖 URL</span>
            <input v-model="tenantForm.previewImage" />
          </label>
          <label class="full">
            <span>Logo URL</span>
            <input v-model="tenantForm.logoImage" />
          </label>
          <label class="full">
            <span>SEO 標題</span>
            <input v-model="tenantForm.seoTitle" />
          </label>
          <label class="full">
            <span>SEO 描述</span>
            <textarea v-model="tenantForm.seoDescription" rows="4"></textarea>
          </label>
        </div>

        <label class="toggle">
          <input v-model="tenantForm.isActive" type="checkbox" />
          <span>租戶站點啟用</span>
        </label>

        <div class="actions">
          <button v-if="!isCreating" class="danger" type="button" @click="removeTenant">刪除租戶</button>
          <button class="primary" type="button" :disabled="isSaving" @click="saveTenant">
            {{ isSaving ? '儲存中...' : isCreating ? '建立租戶' : '儲存租戶設定' }}
          </button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.tenants-page {
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

.subcopy {
  color: var(--wp-text-muted);
  max-width: 70ch;
}

.tenant-layout {
  display: grid;
  gap: 1rem;
  grid-template-columns: 0.9fr 1.1fr;
}

.tenant-list-card,
.editor-card {
  background: #fff;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  box-shadow: var(--wp-shadow);
  padding: 1rem 1.25rem;
}

.tenant-list {
  display: grid;
  gap: 0.75rem;
  margin-top: 1rem;
}

.tenant-item {
  display: grid;
  grid-template-columns: 68px 1fr auto;
  gap: 1rem;
  align-items: center;
  padding: 0.9rem 1rem;
  border: 1px solid var(--wp-border);
  border-radius: 0.5rem;
  background: var(--wp-surface-soft);
  text-align: left;
}

.tenant-item.selected {
  border-color: var(--wp-blue);
  background: #f0f6fc;
}

.tenant-thumb {
  width: 68px;
  height: 52px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: #fff;
}

.tenant-copy p {
  color: var(--wp-text-muted);
}

.pill {
  display: inline-flex;
  align-items: center;
  min-height: 1.9rem;
  padding: 0 0.7rem;
  border-radius: 999px;
  font-weight: 700;
  font-size: 0.8125rem;
}

.pill.active {
  background: rgba(0, 163, 42, 0.12);
  color: var(--wp-green);
}

.pill.inactive {
  background: rgba(214, 54, 56, 0.1);
  color: var(--wp-red);
}

.card-heading {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.card-heading small {
  color: var(--wp-text-muted);
}

.preview-panel {
  display: grid;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.preview-frame {
  overflow: hidden;
  border-radius: 0.5rem;
  border: 1px solid var(--wp-border);
  background: var(--wp-surface-soft);
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
  border-radius: 1rem;
  border: 1px solid var(--wp-border);
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

input,
textarea {
  width: 100%;
  min-height: 2.5rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--wp-border-strong);
  border-radius: 0.375rem;
}

textarea {
  min-height: auto;
  resize: vertical;
}

.toggle {
  display: flex;
  gap: 0.6rem;
  align-items: center;
  margin-top: 1rem;
}

.toggle input {
  width: auto;
  min-height: auto;
}

.actions {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 1rem;
}

.primary,
.danger {
  min-height: 2.5rem;
  padding: 0.65rem 1rem;
  border-radius: 0.375rem;
  font-weight: 600;
  border: 1px solid transparent;
}

.primary {
  background: var(--wp-blue);
  color: #fff;
  border-color: var(--wp-blue);
}

.danger {
  background: #fff;
  color: var(--wp-red);
  border-color: rgba(214, 54, 56, 0.3);
}

@media (max-width: 960px) {
  .page-heading,
  .tenant-layout,
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
