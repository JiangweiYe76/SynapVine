import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '../api/auth'
import type { User, LoginRequest, LogoutRequest, SessionResponse } from '../types/auth'

// Storage keys. Only the non-sensitive user profile is persisted, so a
// reload avoids a flash of unauthenticated UI. The access token lives
// in memory only (Pinia state, never localStorage) and the refresh
// token lives in an httpOnly cookie the browser manages — JS cannot
// read either token from storage, eliminating the localStorage XSS
// surface that previously exposed a 7-day refresh token.
const STORAGE_USER = 'user'

// Pre-emptively refresh this many milliseconds before the access token
// expires. Keeps a 60s safety margin for clock skew and network latency.
const REFRESH_LEAD_MS = 60_000

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  // Access token lives in memory only. On reload it is repopulated by a
  // silent /api/auth/refresh call driven by the router guard; the
  // refresh token in the httpOnly cookie is sent automatically.
  const token = ref<string>('')
  const expiresAt = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isEditor = computed(
    () => user.value?.role === 'admin' || user.value?.role === 'editor',
  )
  const isViewer = computed(() => user.value?.role === 'viewer')

  let refreshTimer: ReturnType<typeof setTimeout> | null = null

  function parseExpiresAt(raw: string | null): number {
    if (!raw) return 0
    const n = Date.parse(raw)
    return Number.isFinite(n) ? n : 0
  }

  function scheduleRefresh(at: number) {
    if (refreshTimer) {
      clearTimeout(refreshTimer)
      refreshTimer = null
    }
    if (at <= 0) return
    const delay = Math.max(at - Date.now() - REFRESH_LEAD_MS, 1_000)
    refreshTimer = setTimeout(() => {
      void doRefresh()
    }, delay)
  }

  function clearRefreshTimer() {
    if (refreshTimer) {
      clearTimeout(refreshTimer)
      refreshTimer = null
    }
  }

  function persist() {
    // Only the user profile is persisted. Tokens are never written to
    // localStorage.
    if (user.value) localStorage.setItem(STORAGE_USER, JSON.stringify(user.value))
    else localStorage.removeItem(STORAGE_USER)
  }

  function applySession(payload: SessionResponse) {
    token.value = payload.token
    expiresAt.value = parseExpiresAt(payload.expires_at)
    if (payload.user) user.value = payload.user
    persist()
    scheduleRefresh(expiresAt.value)
  }

  async function login(credentials: LoginRequest) {
    loading.value = true
    error.value = null
    try {
      const response = await authAPI.login(credentials)
      applySession(response)
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  // doRefresh calls /api/auth/refresh with no body. The refresh token is
  // sent automatically by the browser via the httpOnly cookie. On
  // success a new access token + user are applied in memory; on failure
  // the in-memory session is cleared so the next navigation redirects
  // to /login.
  async function doRefresh(): Promise<boolean> {
    try {
      const response = await authAPI.refresh()
      applySession(response)
      return true
    } catch {
      clearLocal()
      return false
    }
  }

  async function fetchUser() {
    if (!token.value) return false
    try {
      const data = await authAPI.me()
      user.value = data
      persist()
      return true
    } catch {
      clearLocal()
      return false
    }
  }

  async function logout(allDevices = false) {
    const body: LogoutRequest = { all_devices: allDevices }
    // Best-effort: the server might already consider this session
    // invalid. Swallow network / auth errors so the local clear always
    // happens. The backend clears the refresh cookie on its end.
    try {
      await authAPI.logout(body)
    } catch {
      // ignored
    }
    clearLocal()
  }

  function clearLocal() {
    user.value = null
    token.value = ''
    expiresAt.value = 0
    clearRefreshTimer()
    persist()
  }

  return {
    user,
    token,
    expiresAt,
    loading,
    error,
    isAuthenticated,
    isAdmin,
    isEditor,
    isViewer,
    login,
    fetchUser,
    logout,
    refresh: doRefresh,
    clearLocal,
  }
})
