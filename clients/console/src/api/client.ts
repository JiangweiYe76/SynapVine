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

function parsePath(raw: string): { pathname: string; params: URLSearchParams } {
  const [pathname, qs = ''] = raw.replace(/^\//, '').split('?')
  return { pathname, params: new URLSearchParams(qs) }
}

async function mockFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const mock = await ensureMockServer()
  const { pathname, params } = parsePath(path)
  const method = options?.method ?? 'GET'

  if (pathname === 'auth/login' && method === 'POST') {
    const body = JSON.parse(options!.body as string)
    return mock.login(body) as T
  }

  if (pathname === 'me' && method === 'GET') {
    return mock.me() as T
  }

  if (pathname === 'stats' && method === 'GET') {
    return mock.getStats() as T
  }

  if (pathname === 'nodes' && method === 'GET') {
    const offset = parseInt(params.get('offset') ?? '0')
    const limit = parseInt(params.get('limit') ?? '20')
    const search = params.get('search') ?? ''
    return mock.listNodes(offset, limit, search) as T
  }

  if (pathname === 'nodes' && method === 'POST') {
    const body = JSON.parse(options!.body as string)
    return mock.createNode(body) as T
  }

  const nodeMatch = pathname.match(/^nodes\/(.+)$/)
  if (nodeMatch) {
    const id = nodeMatch[1]
    if (method === 'GET') return mock.getNode(id) as T
    if (method === 'PUT') {
      const body = JSON.parse(options!.body as string)
      return mock.updateNode(id, body) as T
    }
    if (method === 'DELETE') {
      mock.deleteNode(id)
      return undefined as T
    }
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
