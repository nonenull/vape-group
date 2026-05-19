import { fetchProducts } from '~/composables/useStoreApi'
import { buildProductPath } from '~/composables/useProductSlug'

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
  const { products } = await fetchProducts(1, 500)

  const urls = [
    `${baseURL}/`,
    `${baseURL}/products`,
    `${baseURL}/about`,
    ...products.map((product) => `${baseURL}${buildProductPath(product)}`),
  ]

  const xml = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...urls.map((url) => `  <url><loc>${escapeXml(url)}</loc></url>`),
    '</urlset>',
  ].join('\n')

  setHeader(event, 'content-type', 'application/xml; charset=utf-8')
  return xml
})
