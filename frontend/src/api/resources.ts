import { api } from './client'
import type {
  Terminal,
  TerminalEvent,
  KeyEntity,
  CardEntity,
  TransactionEntity,
  UserEntity,
} from '../types/api'

function q(limit = 500, offset = 0): string {
  return `?limit=${limit}&offset=${offset}`
}

/** Terminals CRUD */

export async function listTerminals(): Promise<Terminal[]> {
  return api<Terminal[]>(`/terminals${q()}`)
}

export async function createTerminal(body: {
  serial_number: string
  address?: string
  name?: string
  extra?: string
}): Promise<Terminal> {
  return api<Terminal>('/terminals', { method: 'POST', body: JSON.stringify(body) })
}

export async function updateTerminal(
  id: number,
  body: { address?: string; name?: string; extra?: string },
): Promise<Terminal> {
  return api<Terminal>(`/terminals/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export async function deleteTerminal(id: number): Promise<void> {
  return api(`/terminals/${id}`, { method: 'DELETE' })
}

export async function listTerminalEvents(
  since = 0,
  limit = 100,
): Promise<{ items: TerminalEvent[] }> {
  return api<{ items: TerminalEvent[] }>(
    `/terminal/events?since=${encodeURIComponent(String(since))}&limit=${encodeURIComponent(String(limit))}`,
  )
}

/** Keys (admin) */

export async function listKeys(): Promise<KeyEntity[]> {
  return api<KeyEntity[]>(`/keys${q()}`)
}

export async function createKey(body: {
  label?: string
  key_value: string
}): Promise<KeyEntity> {
  return api<KeyEntity>('/keys', { method: 'POST', body: JSON.stringify(body) })
}

export async function updateKey(
  id: number,
  body: { label?: string; key_value?: string },
): Promise<KeyEntity> {
  return api<KeyEntity>(`/keys/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export async function deleteKey(id: number): Promise<void> {
  return api(`/keys/${id}`, { method: 'DELETE' })
}

/** Cards */

export async function listCards(): Promise<CardEntity[]> {
  return api<CardEntity[]>(`/cards${q()}`)
}

export async function createCard(body: {
  card_number: string
  balance: number
  is_blocked: boolean
  owner_name?: string
  extra?: string
  key_id: number
}): Promise<CardEntity> {
  return api<CardEntity>('/cards', { method: 'POST', body: JSON.stringify(body) })
}

export async function updateCard(
  id: number,
  body: {
    balance?: number
    is_blocked?: boolean
    owner_name?: string
    extra?: string
    key_id?: number
  },
): Promise<CardEntity> {
  return api<CardEntity>(`/cards/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export async function deleteCard(id: number): Promise<void> {
  return api(`/cards/${id}`, { method: 'DELETE' })
}

/** Transactions */

export async function listTransactions(): Promise<TransactionEntity[]> {
  return api<TransactionEntity[]>(`/transactions${q()}`)
}

export async function createTransaction(body: {
  amount: number
  card_id: number
  terminal_id: number
}): Promise<TransactionEntity> {
  return api<TransactionEntity>('/transactions', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function updateTransaction(
  id: number,
  body: { amount?: number; card_id?: number; terminal_id?: number },
): Promise<TransactionEntity> {
  return api<TransactionEntity>(`/transactions/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export async function deleteTransaction(id: number): Promise<void> {
  return api(`/transactions/${id}`, { method: 'DELETE' })
}

/** Users */

export async function listUsers(): Promise<UserEntity[]> {
  return api<UserEntity[]>(`/users${q()}`)
}

export async function createUser(body: {
  login: string
  password: string
  display_name?: string
  is_admin: boolean
}): Promise<UserEntity> {
  return api<UserEntity>('/users', { method: 'POST', body: JSON.stringify(body) })
}

export async function getUser(id: number): Promise<UserEntity> {
  return api<UserEntity>(`/users/${id}`)
}

export async function updateUser(
  id: number,
  body: { display_name?: string; password?: string; is_admin?: boolean },
): Promise<UserEntity> {
  return api<UserEntity>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export async function deleteUser(id: number): Promise<void> {
  return api(`/users/${id}`, { method: 'DELETE' })
}
