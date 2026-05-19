import { defineStore } from 'pinia'
import type { PlatformConfig, TenantInfo } from '~/types/store'
import { fetchTenantBundle } from '~/composables/useStoreApi'

export const useTenantStore = defineStore('tenant', {
  state: () => ({
    currentTenant: null as TenantInfo | null,
    tenantId: null as number | null,
    platformConfig: {
      id: 0,
      lineContactUrl: '',
      featuredCategoryIds: [],
      featuredBrandIds: [],
    } as PlatformConfig,
  }),
  getters: {
    getTenantTheme: (state) => state.currentTenant,
    getTenantSEO: (state) => ({
      title: state.currentTenant?.seoTitle ?? '',
      description: state.currentTenant?.seoDescription ?? '',
    }),
  },
  actions: {
    getThemeVariables(tenant: TenantInfo) {
      return {
        '--tenant-accent': tenant.accentColor || '#2271b1',
        '--tenant-accent-strong': tenant.accentStrongColor || '#135e96',
        '--tenant-surface': tenant.surfaceColor || '#f6f7f7',
        '--tenant-page-bg': tenant.pageBgColor || '#f3f6f9',
        '--tenant-card-bg': tenant.cardBgColor || '#ffffff',
        '--tenant-text': tenant.textColor || '#1d2327',
        '--tenant-muted': tenant.mutedTextColor || '#646970',
        '--tenant-border': tenant.borderColor || '#dcdcde',
        '--tenant-hero-bg': tenant.heroBgColor || '#f0f6fc',
        '--tenant-tag-bg': tenant.tagBgColor || '#f0f6fc',
      }
    },
    applyTheme(tenant: TenantInfo) {
      if (!import.meta.client) {
        return
      }

      const root = document.documentElement
      for (const [key, value] of Object.entries(this.getThemeVariables(tenant))) {
        root.style.setProperty(key, value)
      }
    },
    async initTenant(force = false) {
      if (this.currentTenant && !force) {
        return
      }

      const bundle = await fetchTenantBundle()
      this.currentTenant = bundle.tenant
      this.tenantId = bundle.tenant.id
      this.platformConfig = bundle.platformConfig
      this.applyTheme(bundle.tenant)
    },
  },
})
