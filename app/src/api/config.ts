function normalizeBaseUrl(url: string): string {
  return url.replace(/\/+$/, '')
}

function resolveApiBaseUrl(): string {
  const fromEnv = import.meta.env.VITE_API_BASE_URL?.trim()
  if (fromEnv) {
    return normalizeBaseUrl(fromEnv)
  }

  if (typeof window !== 'undefined' && window.location?.origin) {
    return normalizeBaseUrl(window.location.origin)
  }

  return ''
}

export const API_CONFIG = {
  BASE_URL: resolveApiBaseUrl(),
  API_PREFIX: import.meta.env.VITE_API_PREFIX || '/api/user',
  TIMEOUT: 30000,
} as const

export const TOKEN_KEY = 'auth_token'
export const USER_KEY = 'user_info'
