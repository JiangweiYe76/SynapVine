import { fetchAPI } from './client'
import type { LoginRequest, LoginResponse, User } from '../types/auth'

export const authAPI = {
  login: (data: LoginRequest) =>
    fetchAPI<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  me: () => fetchAPI<User>('/me'),
}
