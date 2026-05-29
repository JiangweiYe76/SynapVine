import type { User } from '../types/auth'

export const mockUser: User = {
  id: '1',
  username: 'admin',
  role: 'admin',
  created_at: '2024-01-01T00:00:00Z',
}

export const mockCredentials = {
  username: 'admin',
  password: 'admin123',
}
