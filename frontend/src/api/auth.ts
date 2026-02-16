import type { PublicUser, UpdateProfileRequest } from '../types/user'

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

export interface ConfirmPasswordResetRequest {
  token: string
  new_password: string
}

interface MeResponse {
  user: PublicUser
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

// Mock storage for demo (only used when USE_MOCK is true)
const mockUsers: Record<string, PublicUser & { password: string }> = {}
let mockTokenStore = ''
const mockPasswordResetTokens: Record<string, { email: string; expiresAt: number }> = {}

function mockDelay(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, 400))
}

function toPublicMockUser(user: PublicUser & { password: string }): PublicUser {
  const { password, ...publicUser } = user
  void password
  return publicUser
}

async function request<TResponse>(path: string, options: RequestInit = {}): Promise<TResponse> {
  // Use mock API for frontend-only demo
  if (USE_MOCK) {
    return mockAPI<TResponse>(path, options)
  }

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

// Mock API implementation for frontend-only demos
async function mockAPI<TResponse>(path: string, options: RequestInit): Promise<TResponse> {
  await mockDelay()
  
  const method = options.method || 'GET'
  const body = options.body ? JSON.parse(options.body as string) : null

  // Register
  if (path === '/api/v1/auth/register' && method === 'POST') {
    const { email, password, full_name } = body as RegisterRequest
    
    if (mockUsers[email]) {
      throw new Error('Email already registered')
    }

    const user: PublicUser = {
      id: 'mock_' + Date.now(),
      email,
      full_name
    }

    mockUsers[email] = { ...user, password }
    mockTokenStore = 'mock_token_' + Date.now()

    return { token: mockTokenStore, user } as TResponse
  }

  // Login
  if (path === '/api/v1/auth/login' && method === 'POST') {
    const { email, password } = body as LoginRequest
    const storedUser = mockUsers[email]

    if (!storedUser || storedUser.password !== password) {
      throw new Error('Invalid email or password')
    }

    mockTokenStore = 'mock_token_' + Date.now()
    return { token: mockTokenStore, user: toPublicMockUser(storedUser) } as TResponse
  }

  // Get Me
  if (path === '/api/v1/auth/me' && method === 'GET') {
    const authHeader = (options.headers as Record<string, string>)?.Authorization
    const token = authHeader?.replace('Bearer ', '')

    if (!token || token !== mockTokenStore) {
      throw new Error('Invalid or expired token')
    }

    const firstUser = Object.values(mockUsers)[0]
    if (!firstUser) {
      throw new Error('User not found')
    }

    return { user: toPublicMockUser(firstUser) } as TResponse
  }

  // Logout
  if (path === '/api/v1/auth/logout' && method === 'POST') {
    const authHeader = (options.headers as Record<string, string>)?.Authorization
    const token = authHeader?.replace('Bearer ', '')

    if (!token || token !== mockTokenStore) {
      throw new Error('Invalid or expired token')
    }

    mockTokenStore = ''
    return { message: 'logged out' } as TResponse
  }

  // Update Profile
  if (path === '/api/v1/users/me' && method === 'PATCH') {
    const authHeader = (options.headers as Record<string, string>)?.Authorization
    const token = authHeader?.replace('Bearer ', '')

    if (!token || token !== mockTokenStore) {
      throw new Error('Invalid or expired token')
    }

    const updates = body as UpdateProfileRequest
    const firstEmail = Object.keys(mockUsers)[0]
    const storedUser = mockUsers[firstEmail]

    if (!storedUser) {
      throw new Error('User not found')
    }

    // Apply updates
    if (updates.full_name) storedUser.full_name = updates.full_name
    if (updates.city) storedUser.city = updates.city
    if (updates.bio) storedUser.bio = updates.bio
    if (updates.photo_url) storedUser.photo_url = updates.photo_url

    return { user: toPublicMockUser(storedUser) } as TResponse
  }

  // Password Reset Request
  if (path === '/api/v1/auth/password/reset/request' && method === 'POST') {
    const { email } = body as { email?: string }
    if (!email || !mockUsers[email]) {
      return { message: 'If an account exists, a password reset link has been sent' } as TResponse
    }

    const token = 'mock_reset_' + Date.now()
    mockPasswordResetTokens[token] = {
      email,
      expiresAt: Date.now() + 15 * 60 * 1000
    }
    return { message: `Password reset token (mock): ${token}` } as TResponse
  }

  // Password Reset Confirm
  if (path === '/api/v1/auth/password/reset/confirm' && method === 'POST') {
    const { token, new_password } = body as ConfirmPasswordResetRequest
    const resetEntry = mockPasswordResetTokens[token]
    if (!resetEntry || Date.now() > resetEntry.expiresAt) {
      throw new Error('Invalid or expired reset token')
    }
    if (!new_password || new_password.length < 8) {
      throw new Error('new_password must be at least 8 characters')
    }

    const user = mockUsers[resetEntry.email]
    if (!user) {
      throw new Error('Invalid or expired reset token')
    }

    user.password = new_password
    delete mockPasswordResetTokens[token]
    return { message: 'password reset successful' } as TResponse
  }

  throw new Error('Mock API endpoint not implemented: ' + method + ' ' + path)
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

export function logout(token: string): Promise<{ message: string }> {
  return request<{ message: string }>('/api/v1/auth/logout', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
}

export function updateProfile(token: string, updates: UpdateProfileRequest): Promise<MeResponse> {
  return request<MeResponse>('/api/v1/users/me', {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify(updates)
  })
}

export function requestPasswordReset(email: string): Promise<{ message: string }> {
  return request<{ message: string }>('/api/v1/auth/password/reset/request', {
    method: 'POST',
    body: JSON.stringify({ email })
  })
}

export function confirmPasswordReset(payload: ConfirmPasswordResetRequest): Promise<{ message: string }> {
  return request<{ message: string }>('/api/v1/auth/password/reset/confirm', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}
