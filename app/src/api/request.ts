import { API_CONFIG, TOKEN_KEY, USER_KEY } from './config'

export interface RequestConfig extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>
  timeout?: number
  skipAuth?: boolean
}

export interface ApiResponse<T = unknown> {
  data?: T
  error?: string
  message?: string
}

export class ApiError extends Error {
  public status?: number
  public data?: unknown

  constructor(message: string, status?: number, data?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data

    Object.setPrototypeOf(this, ApiError.prototype)
  }
}

function getFullUrl(url: string, params?: Record<string, string | number | boolean | undefined>): string {
  const baseUrl = `${API_CONFIG.BASE_URL}${API_CONFIG.API_PREFIX}`
  const fullUrl = url.startsWith('http') ? url : `${baseUrl}${url}`

  if (params) {
    const searchParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      searchParams.append(key, String(value))
    })
    return `${fullUrl}?${searchParams.toString()}`
  }

  return fullUrl
}

function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

function getHeaders(config: RequestConfig, body?: BodyInit | null): HeadersInit {
  const headers: HeadersInit = {}

  if (!(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  Object.assign(headers, config.headers)

  if (!config.skipAuth) {
    const token = getToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
  }

  return headers
}

function pickErrorMessage(payload: unknown, status: number): string {
  if (typeof payload === 'object' && payload !== null) {
    if ('error' in payload && typeof payload.error === 'string') {
      return payload.error
    }
    if ('message' in payload && typeof payload.message === 'string') {
      return payload.message
    }
  }
  return `Request failed with status ${status}`
}

async function handleResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get('content-type')
  const isJson = contentType?.includes('application/json')

  let data: unknown
  if (isJson) {
    data = await response.json()
  } else {
    data = await response.text()
  }

  if (response.ok) {
    return data as T
  }

  const errorMessage = pickErrorMessage(data, response.status)

  if (response.status === 401) {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    window.location.href = '/sign-in'
  }

  throw new ApiError(errorMessage, response.status, data)
}

function createTimeoutPromise(timeout: number): Promise<never> {
  return new Promise((_, reject) => {
    setTimeout(() => {
      reject(new ApiError('Request timeout', 408))
    }, timeout)
  })
}

async function request<T = unknown>(
  url: string,
  config: RequestConfig = {}
): Promise<T> {
  const { params, timeout = API_CONFIG.TIMEOUT, skipAuth, body, ...fetchConfig } = config

  const fullUrl = getFullUrl(url, params)
  const headers = getHeaders({ ...config, skipAuth }, body)

  const fetchPromise = fetch(fullUrl, {
    ...fetchConfig,
    headers,
    body,
  }).then(handleResponse<T>)

  return Promise.race([fetchPromise, createTimeoutPromise(timeout)])
}

export const http = {
  get: <T = unknown>(url: string, config?: RequestConfig) =>
    request<T>(url, { ...config, method: 'GET' }),

  post: <T = unknown>(url: string, data?: unknown, config?: RequestConfig) =>
    request<T>(url, {
      ...config,
      method: 'POST',
      body: JSON.stringify(data),
    }),

  upload: <T = unknown>(url: string, formData: FormData, config?: RequestConfig) =>
    request<T>(url, {
      ...config,
      method: 'POST',
      body: formData,
    }),

  put: <T = unknown>(url: string, data?: unknown, config?: RequestConfig) =>
    request<T>(url, {
      ...config,
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: <T = unknown>(url: string, config?: RequestConfig) =>
    request<T>(url, { ...config, method: 'DELETE' }),

  patch: <T = unknown>(url: string, data?: unknown, config?: RequestConfig) =>
    request<T>(url, {
      ...config,
      method: 'PATCH',
      body: JSON.stringify(data),
    }),
}
