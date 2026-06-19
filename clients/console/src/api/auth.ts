import { fetchAPI } from './client'
import type {
  LoginRequest,
  LoginResponse,
  LogoutRequest,
  RefreshRequest,
  SessionResponse,
  User,
} from '../types/auth'

// Auth endpoints that should NOT be subject to the 401-refresh-retry
// logic in fetchAPI:
//   - login: a 401 here means "wrong password", no refresh helps.
//   - refresh: infinite recursion; the server's 401 here is terminal
//     and means the refresh token is gone.
//   - logout: best-effort; must always succeed locally regardless of
//     token state.
// me() is intentionally routed through fetchAPI so a stale access
// token is auto-refreshed before the call goes out.

const API_BASE = '/api'

async function rawPost<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body == null ? undefined : JSON.stringify(body),
  })
  if (!response.ok) {
    let message = response.statusText
    try {
      const err = await response.json()
      if (err && typeof err.message === 'string') message = err.message
    } catch {
      // body wasn't JSON; fall back to statusText
    }
    throw new Error(message)
  }
  if (response.status === 204) {
    return undefined as T
  }
  return response.json() as Promise<T>
}

export const authAPI = {
  login: (data: LoginRequest) => rawPost<LoginResponse>('/auth/login', data),
  refresh: (data: RefreshRequest) => rawPost<SessionResponse>('/auth/refresh', data),
  logout: (data: LogoutRequest) => rawPost<void>('/auth/logout', data),
  me: () => fetchAPI<User>('/me'),
}
