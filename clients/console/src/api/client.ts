import { useAuthStore } from '../stores/auth'

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

  if (pathname === 'auth/refresh' && method === 'POST') {
    const body = JSON.parse(options!.body as string)
    return mock.refresh(body) as T
  }

  if (pathname === 'auth/logout' && method === 'POST') {
    const body = options?.body ? JSON.parse(options.body as string) : {}
    return mock.logout(body) as T
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

  if (pathname === 'edges' && method === 'GET') {
    const offset = parseInt(params.get('offset') ?? '0')
    const limit = parseInt(params.get('limit') ?? '20')
    const search = params.get('search') ?? ''
    return mock.listEdges(offset, limit, search) as T
  }

  if (pathname === 'edges' && method === 'POST') {
    const body = JSON.parse(options!.body as string)
    return mock.createEdge(body) as T
  }

  const edgeMatch = pathname.match(/^edges\/(.+)\/(.+)$/)
  if (edgeMatch) {
    const source = edgeMatch[1]
    const target = edgeMatch[2]
    if (method === 'GET') return mock.getEdge(source, target) as T
    if (method === 'PUT') {
      const body = JSON.parse(options!.body as string)
      return mock.updateEdge(source, target, body) as T
    }
    if (method === 'DELETE') {
      mock.deleteEdge(source, target)
      return undefined as T
    }
  }

  throw new Error(`Unknown mock endpoint: ${path}`)
}

// One inflight refresh at a time. Multiple parallel 401s share the
// same refresh promise so the server sees a single /auth/refresh call
// per session-expiry event.
let refreshInFlight: Promise<boolean> | null = null

async function ensureFreshToken(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight
  const authStore = useAuthStore()
  refreshInFlight = authStore.refresh().finally(() => {
    refreshInFlight = null
  })
  return refreshInFlight
}

function redirectToLogin() {
  const authStore = useAuthStore()
  authStore.clearLocal()
  if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
}

async function readError(response: Response): Promise<Error> {
  let message = response.statusText || 'Request failed'
  try {
    const body = await response.json()
    if (body && typeof body.message === 'string') message = body.message
  } catch {
    // body wasn't JSON; keep statusText
  }
  return new Error(message)
}

export async function fetchAPI<T>(
  path: string,
  options: RequestInit & { _retried?: boolean } = {},
): Promise<T> {
  if (USE_MOCK) {
    return mockFetch<T>(path, options)
  }

  const { _retried, ...rest } = options
  const authStore = useAuthStore()
  const url = `${API_BASE}${path}`

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((rest.headers as Record<string, string>) || {}),
  }

  const currentToken = authStore.token || localStorage.getItem('token') || ''
  if (currentToken) {
    headers['Authorization'] = `Bearer ${currentToken}`
  }

  const response = await fetch(url, { ...rest, headers })

  if (response.ok) {
    if (response.status === 204) return undefined as T
    return response.json() as Promise<T>
  }

  if (response.status === 401 && !_retried) {
    // Try a single silent refresh, then retry the original request.
    // The refresh itself goes through authAPI.refresh (raw fetch) and
    // will not recurse back through this 401 handler.
    const refreshed = await ensureFreshToken()
    if (refreshed) {
      return fetchAPI<T>(path, { ...options, _retried: true })
    }
    // Refresh failed: refresh token is gone. Drop the session and
    // bounce the user to the login page.
    redirectToLogin()
  }

  throw await readError(response)
}
