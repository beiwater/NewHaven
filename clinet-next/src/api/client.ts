import { DEFAULT_COMPANY_ID } from '@/constants'
import { ApiError } from './errors'

const BASE = ''

const AUTH_KEY = 'atlas_auth_token'
const COMPANY_KEY = 'atlas_company_id'
export const AUTH_CHANGED_EVENT = 'atlas-auth-changed'

function getToken(): string | null {
  return localStorage.getItem(AUTH_KEY)
}

function getCompanyId(): string {
  return localStorage.getItem(COMPANY_KEY) ?? DEFAULT_COMPANY_ID
}

export function setAuth(token: string, companyId: string): void {
  localStorage.setItem(AUTH_KEY, token)
  localStorage.setItem(COMPANY_KEY, companyId)
  window.dispatchEvent(new CustomEvent(AUTH_CHANGED_EVENT, { detail: { authenticated: true } }))
}

export function clearAuth(): void {
  localStorage.removeItem(AUTH_KEY)
  localStorage.removeItem(COMPANY_KEY)
  window.dispatchEvent(new CustomEvent(AUTH_CHANGED_EVENT, { detail: { authenticated: false } }))
}

export function isAuthenticated(): boolean {
  return !!getToken()
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    const text = await res.text().catch(() => '')
    let message = text || res.statusText
    try {
      const payload = JSON.parse(text) as { error?: string; message?: string }
      message = payload.error ?? payload.message ?? message
    } catch {
      // Keep the raw response text when the backend does not return JSON.
    }
    if (res.status === 401) {
      clearAuth()
      message = message === 'unauthorized' || message === 'invalid token'
        ? 'Session expired. Please sign in again.'
        : message
    }
    throw new ApiError(res.status, message)
  }

  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
}

export { getCompanyId }
