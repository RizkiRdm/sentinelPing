const BASE_URL = ''

interface SignupData {
  email: string
  password: string
}

interface AuthResponse {
  id: number
  email: string
}

interface ApiError {
  error: string
  message: string
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })

  if (!res.ok) {
    const body: ApiError = await res.json().catch(() => ({
      error: 'unknown',
      message: 'unknown error',
    }))
    throw body
  }

  return res.json()
}

export async function signup(data: SignupData): Promise<AuthResponse> {
  return request<AuthResponse>('/api/auth/signup', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function login(data: SignupData): Promise<AuthResponse> {
  return request<AuthResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function logout(): Promise<void> {
  await request('/api/auth/logout', { method: 'POST' })
}

export async function me(): Promise<AuthResponse> {
  return request<AuthResponse>('/api/auth/me')
}

export type { AuthResponse, ApiError, SignupData }
