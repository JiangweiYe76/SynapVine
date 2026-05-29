import { mockUser, mockCredentials } from './data'
import type { LoginRequest, LoginResponse, User } from '../types/auth'

export function createMockServer() {
  let token: string | null = null

  return {
    login(credentials: LoginRequest): LoginResponse {
      if (
        credentials.username === mockCredentials.username &&
        credentials.password === mockCredentials.password
      ) {
        token = 'mock_token_' + Date.now()
        return { token, user: mockUser }
      }
      throw new Error('Invalid username or password')
    },

    me(_authHeader?: string): User {
      return mockUser
    },
  }
}
