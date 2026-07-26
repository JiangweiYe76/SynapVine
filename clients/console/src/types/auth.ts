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

// /api/auth/logout body. Only the all_devices flag is sent; the refresh
// token itself is read from the httpOnly cookie by the backend.
export interface LogoutRequest {
  all_devices?: boolean
}

// Backend /api/auth/login and /api/auth/refresh return the same shape.
// The refresh token is delivered via an HttpOnly Set-Cookie header and
// is intentionally absent from the JSON body, so client-side JS can
// never read it. /refresh omits the user field in practice; consumers
// should fall back to the previously stored user.
export interface SessionResponse {
  token: string
  expires_at: string
  user?: User
}

export type LoginResponse = SessionResponse

export interface ErrorResponse {
  error: string
  message: string
}
