import { Box, Button, Typography } from '@mui/material'
import { Add } from '@mui/icons-material'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { CrudTable } from '../components/CrudTable'
import {
  EntityDialog,
  type FormFieldDef,
  type FormValues,
} from '../components/EntityDialog'
import * as api from '../api/resources'
import { useAuth } from '../context/AuthContext'
import type { CardEntity, KeyEntity, UserEntity } from '../types/api'

/** Значение для поля owner_name в БД: отображаемое имя или логин. */
function ownerValueFromUser(u: UserEntity): string {
  const d = u.display_name?.trim()
  return d && d.length > 0 ? d : u.login
}

function ownerSelectLabel(u: UserEntity): string {
  const d = u.display_name?.trim()
  return d && d.length > 0 ? `${u.login} — ${d}` : u.login
}

function buildOwnerSelectOptions(users: UserEntity[], currentOwner?: string | null): { value: string; label: string }[] {
  const opts = users.map((u) => ({
    value: ownerValueFromUser(u),
    label: ownerSelectLabel(u),
  }))
  if (currentOwner && currentOwner.trim() !== '' && !opts.some((o) => o.value === currentOwner)) {
    opts.unshift({
      value: currentOwner,
      label: `${currentOwner} (вне списка)`,
    })
  }
  return [{ value: '', label: '— не выбран —' }, ...opts]
}

export function CardsPage() {
  const { user } = useAuth()
  const admin = !!user?.is_admin
  const [rows, setRows] = useState<CardEntity[]>([])
  const [keys, setKeys] = useState<KeyEntity[]>([])
  const [users, setUsers] = useState<UserEntity[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [editRow, setEditRow] = useState<CardEntity | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [cards, ks, us] = await Promise.all([
        api.listCards(),
        admin ? api.listKeys().catch(() => [] as KeyEntity[]) : Promise.resolve([] as KeyEntity[]),
        admin ? api.listUsers().catch(() => [] as UserEntity[]) : Promise.resolve([] as UserEntity[]),
      ])
      setRows(cards)
      setKeys(ks)
      setUsers(us)
    } finally {
      setLoading(false)
    }
  }, [admin])

  useEffect(() => {
    void load()
  }, [load])

  const ownerFieldCreate = useMemo((): FormFieldDef => {
    if (admin && users.length > 0) {
      return {
        name: 'owner_name',
        label: 'Владелец',
        type: 'select',
        options: buildOwnerSelectOptions(users),
      }
    }
    return { name: 'owner_name', label: 'Владелец (имя)', type: 'text' }
  }, [admin, users])

  const ownerFieldEdit = useMemo((): FormFieldDef => {
    if (admin && users.length > 0) {
      return {
        name: 'owner_name',
        label: 'Владелец',
        type: 'select',
        options: buildOwnerSelectOptions(users, editRow?.owner_name),
      }
    }
    return { name: 'owner_name', label: 'Владелец (имя)', type: 'text' }
  }, [admin, users, editRow?.owner_name])

  const createFields = useMemo((): FormFieldDef[] => {
    const base: FormFieldDef[] = [
      { name: 'card_number', label: 'Номер карты', type: 'text', required: true },
      { name: 'balance', label: 'Баланс', type: 'number', required: true, min: 0 },
      { name: 'is_blocked', label: 'Заблокирована', type: 'checkbox' },
      ownerFieldCreate,
      { name: 'extra', label: 'Прочее', type: 'text' },
    ]
    if (admin && keys.length > 0) {
      return [
        ...base,
        {
          name: 'key_id',
          label: 'Ключ',
          type: 'select',
          required: true,
          options: keys.map((k) => ({
            value: k.id,
            label: `#${k.id} ${k.label || k.key_value.slice(0, 12)}`,
          })),
        },
      ]
    }
    return [...base, { name: 'key_id', label: 'ID ключа', type: 'number', required: true, min: 1 }]
  }, [admin, keys, ownerFieldCreate])

  const editFields = useMemo((): FormFieldDef[] => {
    const fld: FormFieldDef[] = [
      { name: 'balance', label: 'Баланс', type: 'number', min: 0 },
      { name: 'is_blocked', label: 'Заблокирована', type: 'checkbox' },
      ownerFieldEdit,
      { name: 'extra', label: 'Прочее', type: 'text' },
    ]
    if (admin && keys.length > 0) {
      fld.push({
        name: 'key_id',
        label: 'Ключ',
        type: 'select',
        options: keys.map((k) => ({
          value: k.id,
          label: `#${k.id} ${k.label || k.key_value.slice(0, 12)}`,
        })),
      })
    }
    return fld
  }, [admin, keys, ownerFieldEdit])

  const onSubmit = async (v: FormValues) => {
    if (editRow) {
      await api.updateCard(editRow.id, {
        balance: v.balance !== undefined && v.balance !== '' ? Number(v.balance) : undefined,
        is_blocked:
          typeof v.is_blocked === 'boolean' ? v.is_blocked : undefined,
        owner_name:
          v.owner_name !== undefined && String(v.owner_name).trim() !== ''
            ? String(v.owner_name)
            : undefined,
        extra: v.extra !== undefined ? String(v.extra) : undefined,
        key_id: v.key_id !== undefined && v.key_id !== '' ? Number(v.key_id) : undefined,
      })
    } else {
      await api.createCard({
        card_number: String(v.card_number),
        balance: Number(v.balance ?? 0),
        is_blocked: Boolean(v.is_blocked),
        owner_name:
          v.owner_name && String(v.owner_name).trim() !== ''
            ? String(v.owner_name)
            : undefined,
        extra: v.extra ? String(v.extra) : undefined,
        key_id: Number(v.key_id),
      })
    }
    await load()
  }

  const onDelete = async (row: CardEntity) => {
    if (!window.confirm(`Удалить карту ${row.card_number}?`)) return
    await api.deleteCard(row.id)
    await load()
  }

  const initialEdit: FormValues | null = editRow
    ? {
        balance: editRow.balance,
        is_blocked: editRow.is_blocked,
        owner_name: editRow.owner_name ?? '',
        extra: editRow.extra ?? '',
        ...(admin && keys.length > 0 ? { key_id: editRow.key_id } : {}),
      }
    : null

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5">Транспортные карты</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={() => { setEditRow(null); setOpen(true) }}>
          Добавить
        </Button>
      </Box>
      <CrudTable
        rows={rows}
        loading={loading}
        columns={[
          { id: 'id', label: 'ID' },
          { id: 'card_number', label: 'Номер' },
          { id: 'balance', label: 'Баланс' },
          {
            id: 'is_blocked',
            label: 'Блок.',
            format: (r) => (r.is_blocked ? 'Да' : ''),
          },
          { id: 'key_id', label: 'Ключ' },
          { id: 'owner_name', label: 'Владелец' },
        ]}
        onEdit={(r) => { setEditRow(r); setOpen(true) }}
        onDelete={onDelete}
      />
      <EntityDialog
        open={open}
        title={editRow ? 'Карта' : 'Новая карта'}
        fields={editRow ? editFields : createFields}
        initial={initialEdit}
        onClose={() => setOpen(false)}
        onSubmit={onSubmit}
      />
    </Box>
  )
}
