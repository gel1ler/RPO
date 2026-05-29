import axios from 'axios'

export const TOKEN_KEY = 'rpo_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(t: string): void {
  localStorage.setItem(TOKEN_KEY, t)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken()
  const headers = new Headers(init?.headers)
  const body = init?.body
  if (body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const axiosHeaders: Record<string, string> = {}
  headers.forEach((value, key) => {
    axiosHeaders[key] = value
  })

  try {
    const res = await axios.request<T>({
      url: `/api/v1${path}`,
      method: init?.method,
      headers: axiosHeaders,
      data: body,
      signal: init?.signal ?? undefined,
      withCredentials: init?.credentials === 'include',
    })

    if (res.status === 204) {
      return undefined as T
    }

    return res.data as T
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const status = error.response?.status ?? 0
      const rawData = error.response?.data as unknown
      const statusText = error.response?.statusText ?? error.message

      if (rawData && typeof rawData === 'object') {
        const errBody = rawData as { error?: string; message?: string }
        throw new ApiError(
          status,
          errBody.error ?? error.code ?? 'error',
          errBody.message ?? statusText,
        )
      }

      if (typeof rawData === 'string') {
        throw new ApiError(status, error.code ?? 'error', rawData || statusText)
      }

      throw new ApiError(status, error.code ?? 'error', statusText)
    }

    throw error
  }
}
