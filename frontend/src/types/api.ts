/** API shapes aligned with backend JSON */

export interface ApiErrorBody {
  error: string
  message: string
}

export interface Me {
  id: number
  login: string
  display_name?: string
  is_admin: boolean
  created_at: string
}

export interface Terminal {
  id: number
  serial_number: string
  address?: string
  name?: string
  extra?: string
  created_at: string
}

export interface KeyEntity {
  id: number
  label?: string
  key_value: string
  created_at: string
}

export interface CardEntity {
  id: number
  card_number: string
  balance: number
  is_blocked: boolean
  owner_name?: string
  extra?: string
  key_id: number
  created_at: string
}

export interface TransactionEntity {
  id: number
  amount: number
  card_id: number
  terminal_id: number
  created_at: string
}

export interface TerminalEvent {
  id: number
  terminal_serial: string
  card_number: string
  operation: string
  amount: number
  trips_delta: number
  approved?: boolean
  reason?: string
  created_at: string
}

export interface UserEntity {
  id: number
  login: string
  display_name?: string
  is_admin: boolean
  created_at: string
}
