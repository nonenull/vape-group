export default defineEventHandler((event) => {
  const host = getHeader(event, 'host') || 'localhost:3000'
  const protocol = getHeader(event, 'x-forwarded-proto') || 'http'
  const baseURL = `${protocol}://${host}`

  setHeader(event, 'content-type', 'text/plain; charset=utf-8')
  return [
    'User-agent: *',
    'Allow: /',
    '',
    'Disallow: /cart',
    'Disallow: /checkout',
    'Disallow: /admin',
    '',
    `Sitemap: ${baseURL}/sitemap.xml`,
    '',
  ].join('\n')
})
