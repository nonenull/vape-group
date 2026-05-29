export default defineEventHandler(async (event) => {
  const path = event.path || event.node.req.url || '/'
  if (
    path.startsWith('/_nuxt/') ||
    path.startsWith('/__nuxt_error') ||
    path.startsWith('/api/') ||
    path.startsWith('/uploads/') ||
    path.startsWith('/wp-content/uploads/') ||
    path === '/favicon.ico'
  ) {
    return
  }

  const host = getHeader(event, 'host')
  if (!host) {
    return
  }

  const runtimeConfig = useRuntimeConfig(event)
  const requestHost = host.split(':')[0]?.trim().toLowerCase()
  if (!requestHost) {
    return
  }

  const protocol = getHeader(event, 'x-forwarded-proto') || 'http'
  const apiBaseURL = runtimeConfig.serverApiBase?.trim() || runtimeConfig.public.apiBase?.trim() || 'http://backend:8088'

  try {
    const response = await $fetch.raw('/__tenant_host_check', {
      baseURL: apiBaseURL.replace(/\/$/, ''),
      method: 'GET',
      headers: {
        host,
        'x-forwarded-proto': protocol,
        'x-tenant-domain': requestHost,
      },
      retry: 0,
    })

    const primaryDomain = response.headers.get('x-primary-domain')
    if (!primaryDomain) {
      return
    }

    const redirectTarget = new URL(primaryDomain)
    if (redirectTarget.hostname === requestHost) {
      return
    }

    redirectTarget.pathname = getRequestURL(event).pathname
    redirectTarget.search = getRequestURL(event).search
    return sendRedirect(event, redirectTarget.toString(), 301)
  } catch (error) {
    console.error('[tenant-domain-redirect] failed to resolve host redirect', error)
  }
})
