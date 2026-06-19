import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '../api/auth'
import type { User, LoginRequest, LogoutRequest } from '../types/auth'

// Storage keys. Centralised so callers never have to know the wire
// format. Stored in localStorage so the user stays signed in across
// page reloads. The XSS-exposure of localStorage is a known trade-off
// for dev; the production fix is to move the refresh token into an
// httpOnly cookie (tracked separately).
const STORAGE_TOKEN = 'token'
const STORAGE_REFRESH = 'refresh_token'
const STORAGE_USER = 'user'
const STORAGE_EXPIRES = 'token_expires_at'

// Pre-emptively refresh this many milliseconds before the access token
// expires. Keeps a 60s safety margin for clock skew and network latency.
const REFRESH_LEAD_MS = 60_000

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string>(localStorage.getItem(STORAGE_TOKEN) || '')
  const refreshToken = ref<string>(localStorage.getItem(STORAGE_REFRESH) || '')
  const expiresAt = ref<number>(parseExpiresAt(localStorage.getItem(STORAGE_EXPIRES)))
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
    if (token.value) localStorage.setItem(STORAGE_TOKEN, token.value)
    else localStorage.removeItem(STORAGE_TOKEN)
    if (refreshToken.value) localStorage.setItem(STORAGE_REFRESH, refreshToken.value)
    else localStorage.removeItem(STORAGE_REFRESH)
    if (expiresAt.value > 0) {
      localStorage.setItem(STORAGE_EXPIRES, new Date(expiresAt.value).toISOString())
    } else {
      localStorage.removeItem(STORAGE_EXPIRES)
    }
    if (user.value) localStorage.setItem(STORAGE_USER, JSON.stringify(user.value))
    else localStorage.removeItem(STORAGE_USER)
  }

  function applySession(payload: {
    token: string
    refresh_token: string
    expires_at: string
    user?: User
  }) {
    token.value = payload.token
    refreshToken.value = payload.refresh_token
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

  async function doRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    try {
      const response = await authAPI.refresh({ refresh_token: refreshToken.value })
      applySession(response)
      return true
    } catch {
      // Refresh failed (revoked / expired). Drop the session and let
      // the next API call (or the router) redirect to /login.
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
    const body: LogoutRequest = { refresh_token: refreshToken.value }
    if (allDevices) body.all_devices = true
    if (refreshToken.value) {
      // Best-effort: the server might already consider this token
      // invalid. Swallow network / auth errors so the local clear
      // always happens.
      try {
        await authAPI.logout(body)
      } catch {
        // ignored
      }
    }
    clearLocal()
  }

  function clearLocal() {
    user.value = null
    token.value = ''
    refreshToken.value = ''
    expiresAt.value = 0
    clearRefreshTimer()
    persist()
  }

  function init() {
    const savedUser = localStorage.getItem(STORAGE_USER)
    if (savedUser) {
      try {
        user.value = JSON.parse(savedUser) as User
      } catch {
        localStorage.removeItem(STORAGE_USER)
      }
    }
    // If we have a token + refresh token, arm the pre-emptive refresh
    // timer so the user doesn't get bounced on the next expired call.
    if (token.value && refreshToken.value && expiresAt.value > 0) {
      scheduleRefresh(expiresAt.value)
    }
  }

  return {
    user,
    token,
    refreshToken,
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
    init,
  }
})
