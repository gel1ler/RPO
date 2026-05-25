import { Box, Button, Typography } from '@mui/material'
import { Add } from '@mui/icons-material'
import { useCallback, useEffect, useState } from 'react'
import { CrudTable } from '../components/CrudTable'
import type { FormFieldDef, FormValues } from '../components/EntityDialog'
import { EntityDialog } from '../components/EntityDialog'
import { useAuth } from '../context/AuthContext'
import * as api from '../api/resources'
import type { UserEntity } from '../types/api'

export function UsersPage() {
  const { user, refreshUser } = useAuth()
  const admin = !!user?.is_admin

  if (!admin) {
    return <ProfileSelf />
  }

  return <UsersAdmin currentId={user!.id} onSelfUpdate={refreshUser} />
}

function ProfileSelf() {
  const { user, refreshUser } = useAuth()
  const [open, setOpen] = useState(false)
  const [initial, setInitial] = useState<FormValues | null>(null)

  const fields: FormFieldDef[] = [
    { name: 'display_name', label: 'Отображаемое имя', type: 'text' },
    { name: 'password', label: 'Новый пароль (необяз.)', type: 'password' },
  ]

  const handleOpen = async () => {
    if (!user) return
    const u = await api.getUser(user.id)
    setInitial({ display_name: u.display_name ?? '', password: '' })
    setOpen(true)
  }

  const onSubmit = async (v: FormValues) => {
    if (!user) return
    const body: { display_name?: string; password?: string } = {}
    body.display_name = String(v.display_name ?? '')
    const p = v.password ? String(v.password).trim() : ''
    if (p) body.password = p
    await api.updateUser(user.id, body)
    await refreshUser()
  }

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Профиль
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 2 }}>
        Логин: {user?.login}
      </Typography>
      <Button variant="contained" onClick={() => void handleOpen()}>
        Редактировать
      </Button>
      <EntityDialog
        open={open}
        title="Мой профиль"
        fields={fields}
        initial={initial}
        onClose={() => setOpen(false)}
        onSubmit={onSubmit}
      />
    </Box>
  )
}

function UsersAdmin({
  currentId,
  onSelfUpdate,
}: {
  currentId: number
  onSelfUpdate: () => Promise<void>
}) {
  const [rows, setRows] = useState<UserEntity[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [editRow, setEditRow] = useState<UserEntity | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(await api.listUsers())
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const createFields: FormFieldDef[] = [
    { name: 'login', label: 'Логин', type: 'text', required: true },
    { name: 'password', label: 'Пароль', type: 'password', required: true },
    { name: 'display_name', label: 'Отображаемое имя', type: 'text' },
    { name: 'is_admin', label: 'Администратор', type: 'checkbox' },
  ]

  const editFields: FormFieldDef[] = [
    { name: 'display_name', label: 'Отображаемое имя', type: 'text' },
    { name: 'password', label: 'Новый пароль (необяз.)', type: 'password' },
    { name: 'is_admin', label: 'Администратор', type: 'checkbox' },
  ]

  const onSubmit = async (v: FormValues) => {
    if (editRow) {
      const body: {
        display_name?: string
        password?: string
        is_admin?: boolean
      } = {
        display_name: String(v.display_name ?? ''),
        is_admin: typeof v.is_admin === 'boolean' ? v.is_admin : undefined,
      }
      const p = v.password ? String(v.password).trim() : ''
      if (p) body.password = p
      await api.updateUser(editRow.id, body)
      if (editRow.id === currentId) await onSelfUpdate()
    } else {
      await api.createUser({
        login: String(v.login),
        password: String(v.password),
        display_name: v.display_name ? String(v.display_name) : undefined,
        is_admin: Boolean(v.is_admin),
      })
    }
    await load()
  }

  const onDelete = async (row: UserEntity) => {
    if (row.id === currentId) {
      window.alert('Нельзя удалить текущего пользователя')
      return
    }
    if (!window.confirm(`Удалить ${row.login}?`)) return
    await api.deleteUser(row.id)
    await load()
  }

  const initialEdit: FormValues | null = editRow
    ? {
        display_name: editRow.display_name ?? '',
        password: '',
        is_admin: editRow.is_admin,
      }
    : null

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5">Пользователи</Typography>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={() => {
            setEditRow(null)
            setOpen(true)
          }}
        >
          Добавить
        </Button>
      </Box>
      <CrudTable
        rows={rows}
        loading={loading}
        columns={[
          { id: 'id', label: 'ID' },
          { id: 'login', label: 'Логин' },
          { id: 'display_name', label: 'Имя' },
          {
            id: 'is_admin',
            label: 'Admin',
            format: (u) => (u.is_admin ? 'Да' : ''),
          },
        ]}
        onEdit={(r) => {
          setEditRow(r)
          setOpen(true)
        }}
        onDelete={onDelete}
      />
      <EntityDialog
        open={open}
        title={editRow ? editRow.login : 'Новый пользователь'}
        fields={editRow ? editFields : createFields}
        initial={initialEdit}
        onClose={() => setOpen(false)}
        onSubmit={onSubmit}
      />
    </Box>
  )
}
