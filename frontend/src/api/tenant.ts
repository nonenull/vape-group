import apiClient from './client'

export interface TenantResponse {
  id: number
  domain: string
  name: string
  is_active: boolean
  theme: string
  preview_image: string
  logo_image: string
  accent_color?: string
  surface_color?: string
  hero_title?: string
  tagline?: string
  announcement?: string
  support_text?: string
  seo_title: string
  seo_description: string
}

export const tenantAPI = {
  getCurrentTenant(): Promise<TenantResponse> {
    return apiClient.get('/api/tenant/current')
  },
}
