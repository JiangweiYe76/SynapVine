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

export interface LoginResponse {
  token: string
  user: User
}

export interface ErrorResponse {
  error: string
  message: string
}
