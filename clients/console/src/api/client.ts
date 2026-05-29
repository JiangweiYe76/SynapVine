const API_BASE = '/api'
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

let _mockServer: ReturnType<typeof import('../mock/server').createMockServer> | null = null

async function ensureMockServer() {
  if (!_mockServer) {
    const { createMockServer } = await import('../mock/server')
    _mockServer = createMockServer()
  }
  return _mockServer
}

async function mockFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const mock = await ensureMockServer()
  const cleanPath = path.replace(/^\//, '')

  if (cleanPath === 'auth/login' && options?.method === 'POST') {
    const body = JSON.parse(options.body as string)
    return mock.login(body) as T
  }

  if (cleanPath === 'me') {
    return mock.me() as T
  }

  throw new Error(`Unknown mock endpoint: ${path}`)
}

export async function fetchAPI<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  if (USE_MOCK) {
    return mockFetch<T>(path, options)
  }

  const url = `${API_BASE}${path}`
  const token = localStorage.getItem('token')

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })

  if (!response.ok) {
    if (response.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    const error = await response.json()
    throw new Error(error.message || 'Request failed')
  }

  return response.json()
}
