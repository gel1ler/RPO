import { Box, Button, Typography } from '@mui/material'
import { Add } from '@mui/icons-material'
import { useCallback, useEffect, useState } from 'react'
import { CrudTable } from '../components/CrudTable'
import { EntityDialog, type FormValues } from '../components/EntityDialog'
import type { FormFieldDef } from '../components/EntityDialog'
import * as api from '../api/resources'
import type { Terminal } from '../types/api'

const createFields: FormFieldDef[] = [
  { name: 'serial_number', label: 'Серийный номер', type: 'text', required: true },
  { name: 'address', label: 'Адрес', type: 'text' },
  { name: 'name', label: 'Название', type: 'text' },
  { name: 'extra', label: 'Прочее', type: 'text' },
]

const editFields: FormFieldDef[] = createFields.filter((f) => f.name !== 'serial_number')

export function TerminalsPage() {
  const [rows, setRows] = useState<Terminal[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editRow, setEditRow] = useState<Terminal | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(await api.listTerminals())
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const openCreate = () => {
    setEditRow(null)
    setDialogOpen(true)
  }

  const openEdit = (row: Terminal) => {
    setEditRow(row)
    setDialogOpen(true)
  }

  const onSubmit = async (v: FormValues) => {
    if (editRow) {
      await api.updateTerminal(editRow.id, {
        address: String(v.address || ''),
        name: String(v.name || ''),
        extra: String(v.extra || ''),
      })
    } else {
      await api.createTerminal({
        serial_number: String(v.serial_number),
        address: v.address ? String(v.address) : undefined,
        name: v.name ? String(v.name) : undefined,
        extra: v.extra ? String(v.extra) : undefined,
      })
    }
    await load()
  }

  const onDelete = async (row: Terminal) => {
    if (!window.confirm(`Удалить терминал ${row.serial_number}?`)) {
      return
    }
    await api.deleteTerminal(row.id)
    await load()
  }

  const initial: FormValues | null = editRow
    ? {
        address: editRow.address ?? '',
        name: editRow.name ?? '',
        extra: editRow.extra ?? '',
      }
    : null

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">Терминалы</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={openCreate}>
          Добавить
        </Button>
      </Box>
      <CrudTable
        rows={rows}
        loading={loading}
        onEdit={openEdit}
        onDelete={onDelete}
        columns={[
          { id: 'id', label: 'ID' },
          { id: 'serial_number', label: 'SN' },
          { id: 'address', label: 'Адрес' },
          { id: 'name', label: 'Название' },
          { id: 'created_at', label: 'Создан' },
        ]}
      />
      <EntityDialog
        open={dialogOpen}
        title={editRow ? 'Изменить терминал' : 'Новый терминал'}
        fields={editRow ? editFields : createFields}
        initial={initial}
        onClose={() => setDialogOpen(false)}
        onSubmit={onSubmit}
      />
    </Box>
  )
}
