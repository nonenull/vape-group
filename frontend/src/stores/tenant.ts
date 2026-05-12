import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { tenantAPI } from '@/api/tenant'
import type { TenantInfo } from '@/data/mockProducts'

export const useTenantStore = defineStore('tenant', () => {
  const currentTenant = ref<TenantInfo | null>(null)
  const tenantId = ref<number | null>(null)

  const applyTheme = (tenant: TenantInfo) => {
    const root = document.documentElement
    root.style.setProperty('--tenant-accent', tenant.accentColor || '#2271b1')
    root.style.setProperty('--tenant-surface', tenant.surfaceColor || '#f6f7f7')
  }

  const initTenant = async () => {
    const response = await tenantAPI.getCurrentTenant()
    const tenant: TenantInfo = {
      id: response.id,
      domain: response.domain,
      name: response.name,
      isActive: response.is_active,
      theme: response.theme,
      previewImage: response.preview_image,
      logoImage: response.logo_image,
      accentColor: response.accent_color,
      surfaceColor: response.surface_color,
      heroTitle: response.hero_title,
      tagline: response.tagline,
      announcement: response.announcement,
      supportText: response.support_text,
      seoTitle: response.seo_title,
      seoDescription: response.seo_description,
    }

    currentTenant.value = tenant
    tenantId.value = tenant.id
    localStorage.setItem('tenant_id', tenant.id.toString())
    applyTheme(tenant)
  }

  const getTenantTheme = computed(() => currentTenant.value || null)
  const getTenantSEO = computed(() => ({
    title: currentTenant.value?.seoTitle ?? '',
    description: currentTenant.value?.seoDescription ?? '',
  }))

  return {
    currentTenant,
    tenantId,
    initTenant,
    getTenantTheme,
    getTenantSEO,
  }
})
