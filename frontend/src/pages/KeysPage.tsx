import { Navigate } from 'react-router-dom'
import { Box, Button, Typography } from '@mui/material'
import { Add } from '@mui/icons-material'
import { useAuth } from '../context/AuthContext'
import { useCallback, useEffect, useState } from 'react'
import { CrudTable } from '../components/CrudTable'
import { EntityDialog, type FormFieldDef, type FormValues } from '../components/EntityDialog'
import type { KeyEntity } from '../types/api'
import * as api from '../api/resources'

const fields: FormFieldDef[] = [
  { name: 'label', label: 'Метка', type: 'text' },
  { name: 'key_value', label: 'Значение ключа', type: 'text', required: true },
]

export function KeysPage() {
  const { user } = useAuth()
  const [rows, setRows] = useState<KeyEntity[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [editRow, setEditRow] = useState<KeyEntity | null>(null)

  if (!user?.is_admin) {
    return <Navigate to="/" replace />
  }

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(await api.listKeys())
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const onSubmit = async (v: FormValues) => {
    if (editRow) {
      await api.updateKey(editRow.id, {
        label: v.label ? String(v.label) : '',
        key_value: v.key_value ? String(v.key_value) : undefined,
      })
    } else {
      await api.createKey({
        label: v.label ? String(v.label) : undefined,
        key_value: String(v.key_value),
      })
    }
    await load()
  }

  const onDelete = async (row: KeyEntity) => {
    if (!window.confirm('Удалить ключ?')) return
    await api.deleteKey(row.id)
    await load()
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5">Ключи</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={() => { setEditRow(null); setOpen(true) }}>
          Добавить
        </Button>
      </Box>
      <CrudTable
        rows={rows}
        loading={loading}
        columns={[
          { id: 'id', label: 'ID' },
          { id: 'label', label: 'Метка' },
          { id: 'key_value', label: 'Значение' },
          { id: 'created_at', label: 'Создан' },
        ]}
        onEdit={(row) => { setEditRow(row); setOpen(true) }}
        onDelete={onDelete}
      />
      <EntityDialog
        open={open}
        title={editRow ? 'Ключ' : 'Новый ключ'}
        fields={fields}
        initial={
          editRow
            ? { label: editRow.label ?? '', key_value: editRow.key_value }
            : null
        }
        onClose={() => setOpen(false)}
        onSubmit={onSubmit}
      />
    </Box>
  )
}
