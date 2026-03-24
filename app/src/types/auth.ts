export type UserRole = 'user'

export interface User {
  id: number
  email: string
  phone: string
  username: string
  avatar_url: string
  last_active_at: string
  created_at: string
  updated_at: string
  role: UserRole
}

export interface LoginRequest {
  account: string
  password: string
}

export interface RegisterRequest {
  email?: string
  phone?: string
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RegisterResponse {
  token: string
  user: User
}

export interface LogoutResponse {
  message: string
}
