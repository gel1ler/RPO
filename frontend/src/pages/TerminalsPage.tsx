import { Box, Button, Typography, Paper, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material'
import { Add } from '@mui/icons-material'
import { useCallback, useEffect, useRef, useState } from 'react'
import { CrudTable } from '../components/CrudTable'
import { EntityDialog, type FormValues } from '../components/EntityDialog'
import type { FormFieldDef } from '../components/EntityDialog'
import * as api from '../api/resources'
import type { Terminal, TerminalEvent } from '../types/api'

const createFields: FormFieldDef[] = [
  { name: 'serial_number', label: 'Серийный номер', type: 'text', required: true },
  { name: 'address', label: 'Адрес', type: 'text' },
  { name: 'name', label: 'Название', type: 'text' },
  { name: 'extra', label: 'Прочее', type: 'text' },
]

const editFields: FormFieldDef[] = createFields.filter((f) => f.name !== 'serial_number')

function fmtApprove(e: TerminalEvent): string {
  if (typeof e.approved !== 'boolean') return '—'
  return e.approved ? 'Да' : 'Нет'
}

export function TerminalsPage() {
  const [rows, setRows] = useState<Terminal[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editRow, setEditRow] = useState<Terminal | null>(null)
  const [events, setEvents] = useState<TerminalEvent[]>([])
  const sinceRef = useRef(0)

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

  useEffect(() => {
    const timer = window.setInterval(() => {
      void (async () => {
        try {
          const { items } = await api.listTerminalEvents(sinceRef.current)
          if (items.length === 0) return
          const highest = Math.max(...items.map((it) => it.id))
          sinceRef.current = Math.max(sinceRef.current, highest)
          setEvents((prev) => {
            const byId = new Map<number, TerminalEvent>()
            prev.forEach((e) => byId.set(e.id, e))
            items.forEach((e) => byId.set(e.id, e))
            return [...byId.values()].sort((a, b) => b.id - a.id).slice(0, 120)
          })
        } catch {
          /* сеть недоступна — тихий пропуск */
        }
      })()
    }, 2000)
    return () => window.clearInterval(timer)
  }, [])

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

      <Typography variant="h6" sx={{ mt: 4, mb: 1 }}>
        Журнал терминала (обновление ~2 сек)
      </Typography>
      <Paper variant="outlined" sx={{ maxHeight: 420, overflow: 'auto' }}>
        <Table size="small" stickyHeader>
          <TableHead>
            <TableRow>
              <TableCell>Причина / время</TableCell>
              <TableCell>Операция</TableCell>
              <TableCell>SN</TableCell>
              <TableCell>Карта</TableCell>
              <TableCell align="right">Сумма</TableCell>
              <TableCell>ОК</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {events.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  Пока нет событий (появятся после работы приложения-терминала)
                </TableCell>
              </TableRow>
            ) : (
              events.map((e) => (
                <TableRow key={e.id} hover>
                  <TableCell>
                    {e.created_at}
                    {e.reason != null ? ` — ${e.reason}` : ''}
                  </TableCell>
                  <TableCell>{e.operation}</TableCell>
                  <TableCell>{e.terminal_serial}</TableCell>
                  <TableCell>{e.card_number}</TableCell>
                  <TableCell align="right">{e.amount}</TableCell>
                  <TableCell>{fmtApprove(e)}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Paper>
    </Box>
  )
}
