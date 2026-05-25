import { api } from './client'
import type { Me } from '../types/api'

export interface LoginResult {
  token: string
}

export async function login(login: string, password: string): Promise<LoginResult> {
  return api<LoginResult>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ login, password }),
  })
}

export async function fetchMe(): Promise<Me> {
  return api<Me>('/me')
}
