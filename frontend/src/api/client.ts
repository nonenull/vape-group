import axios from 'axios'

const getApiBaseURL = () => {
  const envBaseURL = import.meta.env.VITE_API_URL?.trim()
  if (envBaseURL) {
    return envBaseURL.replace(/\/$/, '')
  }

  const { protocol, host, hostname, port } = window.location

  if (port === '8880') {
    return `${protocol}//${host}`
  }

  // In local Vite dev, keep the tenant hostname but switch requests to the backend port.
  if (port === '5173' || port === '5174') {
    return `http://${hostname}:8088`
  }

  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'http://localhost:8088'
  }

  return `${protocol}//${host}`
}

export const resolveAssetURL = (value: string) => {
  if (!value) {
    return value
  }
  if (/^(https?:)?\/\//.test(value) || value.startsWith('data:')) {
    return value
  }
  if (value.startsWith('/')) {
    return `${getApiBaseURL()}${value}`
  }
  return value
}

const apiClient = axios.create({
  baseURL: getApiBaseURL(),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use(
  (config: any) => {
    const tenantDomain = window.location.hostname
    if (config.headers) {
      config.headers['X-Tenant-Domain'] = tenantDomain
    }

    const token = localStorage.getItem('auth_token')
    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error: any) => Promise.reject(error),
)

apiClient.interceptors.response.use(
  (response: any) => response.data,
  (error: any) => {
    if (error?.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

export default apiClient
