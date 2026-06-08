import { proxyRequest } from 'h3'

export default defineEventHandler((event) => {
  const config = useRuntimeConfig()
  const backendBase = (config.serverApiBase || config.public.apiBase || 'http://backend:8088').replace(/\/$/, '')
  const path = event.path || '/uploads'
  return proxyRequest(event, `${backendBase}${path}`)
})
