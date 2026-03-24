import { http } from './request'
import { TOKEN_KEY, USER_KEY } from './config'
import type {
  LoginRequest,
  LoginResponse,
  LogoutResponse,
  RegisterRequest,
  RegisterResponse,
  User,
} from '@/types/auth'

export async function login(data: LoginRequest): Promise<LoginResponse> {
  const response = await http.post<LoginResponse>('/auth/login', data, { skipAuth: true })

  if (response.token) {
    localStorage.setItem(TOKEN_KEY, response.token)
  }
  if (response.user) {
    localStorage.setItem(USER_KEY, JSON.stringify(response.user))
  }

  return response
}

export async function register(data: RegisterRequest): Promise<RegisterResponse> {
  const response = await http.post<RegisterResponse>('/auth/register', data, {
    skipAuth: true,
  })

  if (response.token) {
    localStorage.setItem(TOKEN_KEY, response.token)
  }
  if (response.user) {
    localStorage.setItem(USER_KEY, JSON.stringify(response.user))
  }

  return response
}

export async function logout(): Promise<LogoutResponse> {
  try {
    const response = await http.post<LogoutResponse>('/auth/logout')
    return response
  } finally {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }
}

export function getCurrentUser(): User | null {
  const userStr = localStorage.getItem(USER_KEY)
  if (!userStr) return null

  try {
    return JSON.parse(userStr) as User
  } catch {
    return null
  }
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function isAuthenticated(): boolean {
  return !!getToken()
}
