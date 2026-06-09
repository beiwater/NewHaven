import { DEFAULT_COMPANY_ID } from '@/constants'

const BASE = ''

const AUTH_KEY = 'atlas_auth_token'
const COMPANY_KEY = 'atlas_company_id'
const NEW_ACCOUNT_KEY = 'atlas_new_account'
export const AUTH_CHANGED_EVENT = 'atlas-auth-changed'

function getToken(): string | null {
  return localStorage.getItem(AUTH_KEY)
}

function getCompanyId(): string {
  return localStorage.getItem(COMPANY_KEY) ?? DEFAULT_COMPANY_ID
}

export function setAuth(token: string, companyId: string, newAccount = false): void {
  localStorage.setItem(AUTH_KEY, token)
  localStorage.setItem(COMPANY_KEY, companyId)
  if (newAccount) {
    sessionStorage.setItem(NEW_ACCOUNT_KEY, 'true')
  } else {
    sessionStorage.removeItem(NEW_ACCOUNT_KEY)
  }
  window.dispatchEvent(new CustomEvent(AUTH_CHANGED_EVENT, { detail: { authenticated: true } }))
}

export function clearAuth(): void {
  localStorage.removeItem(AUTH_KEY)
  localStorage.removeItem(COMPANY_KEY)
  sessionStorage.removeItem(NEW_ACCOUNT_KEY)
  window.dispatchEvent(new CustomEvent(AUTH_CHANGED_EVENT, { detail: { authenticated: false } }))
}

export function isAuthenticated(): boolean {
  return !!getToken()
}

export function isNewAccountSession(): boolean {
  return sessionStorage.getItem(NEW_ACCOUNT_KEY) === 'true'
}

export function finishNewAccountSession(): void {
  sessionStorage.removeItem(NEW_ACCOUNT_KEY)
}

export class ApiError extends Error {
  status: number = 0
  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

interface ApiEnvelope<T> {
  data: T
  error: {
    code?: string
    message?: string
    details?: unknown
  } | null
  meta?: unknown
}

function isApiEnvelope<T>(payload: unknown): payload is ApiEnvelope<T> {
  return typeof payload === 'object'
    && payload !== null
    && 'data' in payload
    && 'error' in payload
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

  const text = await res.text().catch(() => '')
  let payload: unknown
  try {
    payload = text ? JSON.parse(text) : null
  } catch {
    payload = text
  }

  if (!res.ok) {
    let message = text || res.statusText
    if (isApiEnvelope<unknown>(payload)) {
      message = payload.error?.message ?? message
    } else if (typeof payload === 'object' && payload !== null) {
      const errorPayload = payload as { error?: unknown; message?: unknown }
      if (typeof errorPayload.error === 'string') {
        message = errorPayload.error
      } else if (typeof errorPayload.message === 'string') {
        message = errorPayload.message
      }
    }
    if (res.status === 401) {
      clearAuth()
      message = message === 'unauthorized' || message === 'invalid token'
        ? 'Session expired. Please sign in again.'
        : message
    }
    throw new ApiError(res.status, message)
  }

  if (isApiEnvelope<T>(payload)) {
    if (payload.error) {
      throw new ApiError(res.status, payload.error.message ?? payload.error.code ?? 'API request failed')
    }
    return payload.data
  }
  return payload as T
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
}

export { getCompanyId }
