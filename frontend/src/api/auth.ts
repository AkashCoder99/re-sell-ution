export interface PublicUser {
  id: string
  email: string
  full_name: string
}

export interface AuthPayload {
  token: string
  user: PublicUser
}

export interface RegisterRequest {
  full_name: string
  email: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

interface MeResponse {
  user: PublicUser
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

async function request<TResponse>(path: string, options: RequestInit = {}): Promise<TResponse> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    ...options
  })

  const data: unknown = await response.json().catch(() => ({}))

  if (!response.ok) {
    const message =
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as { error?: unknown }).error === 'string'
        ? (data as { error: string }).error
        : 'Request failed'

    throw new Error(message)
  }

  return data as TResponse
}

export function register(payload: RegisterRequest): Promise<AuthPayload> {
  return request<AuthPayload>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export function login(payload: LoginRequest): Promise<AuthPayload> {
  return request<AuthPayload>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export function getMe(token: string): Promise<MeResponse> {
  return request<MeResponse>('/api/v1/auth/me', {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
}
