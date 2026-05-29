import { buildCategoryPath, buildProductPath } from '~/composables/useProductSlug'

interface SitemapCategory {
  id: number
  name: string
}

interface SitemapProduct {
  id: number
  name?: string
  base_name?: string
  custom_name?: string
  slug?: string
  is_visible?: boolean
}

interface SitemapProductsResponse {
  data?: SitemapProduct[]
  total?: number
}

function escapeXml(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&apos;')
}

export default defineEventHandler(async (event) => {
  const host = getHeader(event, 'host') || 'localhost:3000'
  const protocol = getHeader(event, 'x-forwarded-proto') || 'http'
  const baseURL = `${protocol}://${host}`
  const runtimeConfig = useRuntimeConfig(event)
  const apiBaseURL = (runtimeConfig.serverApiBase || runtimeConfig.public.apiBase || 'http://backend:8088').replace(/\/$/, '')
  const pageSize = 500
  const requestHeaders = {
    host,
    'x-forwarded-proto': protocol,
    'x-tenant-domain': getHeader(event, 'x-tenant-domain') || host,
  }

  const [firstPageResponse, categories] = await Promise.all([
    $fetch<SitemapProductsResponse>('/api/products', {
      baseURL: apiBaseURL,
      headers: requestHeaders,
      query: {
        page: 1,
        limit: pageSize,
      },
    }),
    $fetch<SitemapCategory[]>('/api/categories', {
      baseURL: apiBaseURL,
      headers: requestHeaders,
    }),
  ])

  const products = (firstPageResponse.data || [])
    .filter((product) => product.is_visible !== false)
    .map((product) => ({
      id: Number(product.id),
      name: product.custom_name || product.base_name || product.name || 'product',
      slug: product.slug?.trim() || undefined,
    }))
  const totalPages = Math.max(1, Math.ceil(Number(firstPageResponse.total || products.length) / pageSize))

  for (let page = 2; page <= totalPages; page += 1) {
    const response = await $fetch<SitemapProductsResponse>('/api/products', {
      baseURL: apiBaseURL,
      headers: requestHeaders,
      query: {
        page,
        limit: pageSize,
      },
    })

    products.push(...(response.data || [])
      .filter((product) => product.is_visible !== false)
      .map((product) => ({
        id: Number(product.id),
        name: product.custom_name || product.base_name || product.name || 'product',
        slug: product.slug?.trim() || undefined,
      })))
  }

  const urls = new Set([
    `${baseURL}/`,
    `${baseURL}/products`,
    `${baseURL}/about`,
    ...categories.map((category) => `${baseURL}${buildCategoryPath({
      id: category.id,
      name: category.name,
    })}`),
    ...products.map((product) => `${baseURL}${buildProductPath(product)}`),
  ])

  const xml = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...Array.from(urls).map((url) => `  <url><loc>${escapeXml(url)}</loc></url>`),
    '</urlset>',
  ].join('\n')

  setHeader(event, 'content-type', 'application/xml; charset=utf-8')
  setHeader(event, 'cache-control', 'public, max-age=3600, s-maxage=3600')
  return xml
})
