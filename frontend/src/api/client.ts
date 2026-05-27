import { useAuthStore } from '../store/auth'

const BASE = '/api'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = useAuthStore.getState().token
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string>),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, { ...init, headers })
  const body = await res.json()

  if (!res.ok) {
    throw new ApiError(res.status, body.error ?? 'Request failed')
  }
  return body as T
}

export const api = {
  get:  <T>(path: string)              => request<T>(path, { method: 'GET' }),
  post: <T>(path: string, data: unknown) => request<T>(path, { method: 'POST', body: JSON.stringify(data) }),
  put:  <T>(path: string, data: unknown) => request<T>(path, { method: 'PUT',  body: JSON.stringify(data) }),
}
