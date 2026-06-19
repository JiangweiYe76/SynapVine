export type Role = 'admin' | 'editor' | 'viewer'

export interface User {
  id: string
  username: string
  role: Role
  created_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RefreshRequest {
  refresh_token: string
}

export interface LogoutRequest {
  refresh_token: string
  all_devices?: boolean
}

// Backend /api/auth/login and /api/auth/refresh return the same shape.
// /refresh omits the user field in practice; consumers should fall back
// to the previously stored user.
export interface SessionResponse {
  token: string
  refresh_token: string
  expires_at: string
  user?: User
}

export type LoginResponse = SessionResponse

export interface ErrorResponse {
  error: string
  message: string
}
