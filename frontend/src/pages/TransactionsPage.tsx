import { Box, Button, Typography } from '@mui/material'
import { Add } from '@mui/icons-material'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { CrudTable } from '../components/CrudTable'
import type { FormFieldDef, FormValues } from '../components/EntityDialog'
import { EntityDialog } from '../components/EntityDialog'
import * as api from '../api/resources'
import type { Terminal, CardEntity, TransactionEntity } from '../types/api'

const POLL_MS = 2000

export function TransactionsPage() {
  const [rows, setRows] = useState<TransactionEntity[]>([])
  const [cards, setCards] = useState<CardEntity[]>([])
  const [terminals, setTerminals] = useState<Terminal[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [editRow, setEditRow] = useState<TransactionEntity | null>(null)

  const load = useCallback(async (silent = false) => {
    if (!silent) setLoading(true)
    try {
      const [tRows, cs, tm] = await Promise.all([
        api.listTransactions(),
        api.listCards(),
        api.listTerminals(),
      ])
      setRows(tRows)
      setCards(cs)
      setTerminals(tm)
    } catch {
      /* сеть недоступна — тихий пропуск при фоновом опросе */
    } finally {
      if (!silent) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    const timer = window.setInterval(() => {
      void load(true)
    }, POLL_MS)
    return () => window.clearInterval(timer)
  }, [load])

  const createFields: FormFieldDef[] = useMemo(
    () => [
      {
        name: 'amount',
        label: 'Сумма',
        type: 'number',
        required: true,
        min: 1,
      },
      {
        name: 'card_id',
        label: 'Карта',
        type: 'select',
        required: true,
        options: cards.map((c) => ({
          value: c.id,
          label: `#${c.id} ${c.card_number}`,
        })),
      },
      {
        name: 'terminal_id',
        label: 'Терминал',
        type: 'select',
        required: true,
        options: terminals.map((t) => ({
          value: t.id,
          label: `#${t.id} ${t.serial_number}`,
        })),
      },
    ],
    [cards, terminals],
  )

  const editFields: FormFieldDef[] = useMemo(
    () => [
      { name: 'amount', label: 'Сумма', type: 'number', min: 1 },
      {
        name: 'card_id',
        label: 'Карта',
        type: 'select',
        options: cards.map((c) => ({
          value: c.id,
          label: `#${c.id} ${c.card_number}`,
        })),
      },
      {
        name: 'terminal_id',
        label: 'Терминал',
        type: 'select',
        options: terminals.map((t) => ({
          value: t.id,
          label: `#${t.id} ${t.serial_number}`,
        })),
      },
    ],
    [cards, terminals],
  )

  const onSubmit = async (v: FormValues) => {
    if (editRow) {
      await api.updateTransaction(editRow.id, {
        amount: v.amount !== '' && v.amount !== undefined ? Number(v.amount) : undefined,
        card_id:
          v.card_id !== '' && v.card_id !== undefined ? Number(v.card_id) : undefined,
        terminal_id:
          v.terminal_id !== '' && v.terminal_id !== undefined
            ? Number(v.terminal_id)
            : undefined,
      })
    } else {
      await api.createTransaction({
        amount: Number(v.amount),
        card_id: Number(v.card_id),
        terminal_id: Number(v.terminal_id),
      })
    }
    await load()
  }

  const onDelete = async (row: TransactionEntity) => {
    if (!window.confirm('Удалить транзакцию?')) return
    await api.deleteTransaction(row.id)
    await load()
  }

  const initialEdit: FormValues | null = editRow
    ? {
        amount: editRow.amount,
        card_id: editRow.card_id,
        terminal_id: editRow.terminal_id,
      }
    : null

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5">Транзакции</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={() => { setEditRow(null); setOpen(true) }}>
          Добавить
        </Button>
      </Box>
      <CrudTable
        rows={rows}
        loading={loading}
        columns={[
          { id: 'id', label: 'ID' },
          { id: 'amount', label: 'Сумма' },
          { id: 'card_id', label: 'Карта' },
          { id: 'terminal_id', label: 'Терминал' },
          { id: 'created_at', label: 'Время' },
        ]}
        onEdit={(r) => { setEditRow(r); setOpen(true) }}
        onDelete={onDelete}
      />
      <EntityDialog
        open={open}
        title={editRow ? 'Транзакция' : 'Новая транзакция'}
        fields={editRow ? editFields : createFields}
        initial={initialEdit}
        onClose={() => setOpen(false)}
        onSubmit={onSubmit}
      />
    </Box>
  )
}
