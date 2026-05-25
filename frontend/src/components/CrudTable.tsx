import {
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Typography,
} from '@mui/material'
import { Delete, Edit } from '@mui/icons-material'

export interface Column<T> {
  id: keyof T | string
  label: string
  format?: (row: T) => React.ReactNode
}

interface CrudTableProps<T extends { id: number }> {
  rows: T[]
  columns: Column<T>[]
  onEdit?: (row: T) => void
  onDelete?: (row: T) => void
  loading?: boolean
}

export function CrudTable<T extends { id: number }>({
  rows,
  columns,
  onEdit,
  onDelete,
  loading,
}: CrudTableProps<T>) {
  if (!loading && rows.length === 0) {
    return <Typography color="text.secondary">Нет данных</Typography>
  }

  return (
    <TableContainer component={Paper}>
      <Table size="small">
        <TableHead>
          <TableRow>
            {columns.map((c) => (
              <TableCell key={String(c.id)}>{c.label}</TableCell>
            ))}
            {(onEdit || onDelete) && <TableCell align="right">Действия</TableCell>}
          </TableRow>
        </TableHead>
        <TableBody>
          {loading ? (
            <TableRow>
              <TableCell colSpan={columns.length + (onEdit || onDelete ? 1 : 0)}>
                Загрузка…
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow key={row.id}>
                {columns.map((c) => (
                  <TableCell key={String(c.id)}>
                    {c.format
                      ? c.format(row)
                      : String((row as Record<string, unknown>)[String(c.id)] ?? '')}
                  </TableCell>
                ))}
                {(onEdit || onDelete) && (
                  <TableCell align="right">
                    {onEdit && (
                      <Tooltip title="Изменить">
                        <IconButton size="small" onClick={() => onEdit(row)}>
                          <Edit fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}
                    {onDelete && (
                      <Tooltip title="Удалить">
                        <IconButton size="small" color="error" onClick={() => onDelete(row)}>
                          <Delete fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}
                  </TableCell>
                )}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </TableContainer>
  )
}
