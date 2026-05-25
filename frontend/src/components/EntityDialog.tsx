import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Switch,
  TextField,
  MenuItem,
  Box,
} from '@mui/material'
import { useEffect, useState } from 'react'

export type FieldType = 'text' | 'number' | 'password' | 'checkbox' | 'select'

export interface FormFieldDef {
  name: string
  label: string
  type: FieldType
  required?: boolean
  options?: { value: number | string; label: string }[]
  disabled?: boolean
  min?: number
}

export type FormValues = Record<string, string | number | boolean | undefined>

function emptyFromFields(fields: FormFieldDef[]): FormValues {
  const o: FormValues = {}
  for (const f of fields) {
    if (f.type === 'checkbox') {
      o[f.name] = false
    } else if (f.type === 'number') {
      o[f.name] = f.min !== undefined ? f.min : 0
    } else if (f.type === 'select' && f.options?.length) {
      o[f.name] = f.options[0].value
    } else {
      o[f.name] = ''
    }
  }
  return o
}

export function EntityDialog({
  open,
  title,
  fields,
  initial,
  onClose,
  onSubmit,
}: {
  open: boolean
  title: string
  fields: FormFieldDef[]
  initial?: FormValues | null
  onClose: () => void
  onSubmit: (v: FormValues) => void | Promise<void>
}) {
  const [values, setValues] = useState<FormValues>(() => emptyFromFields(fields))
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({ ...emptyFromFields(fields), ...initial })
    }
  }, [open, fields, initial])

  const set = (name: string, v: string | number | boolean | undefined) => {
    setValues((prev) => ({ ...prev, [name]: v }))
  }

  const handleSave = async () => {
    setSubmitting(true)
    try {
      await onSubmit(values)
      onClose()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
          {fields.map((f) => {
            if (f.type === 'checkbox') {
              return (
                <FormControlLabel
                  key={f.name}
                  control={
                    <Switch
                      checked={Boolean(values[f.name])}
                      onChange={(e) => set(f.name, e.target.checked)}
                      disabled={f.disabled}
                    />
                  }
                  label={f.label}
                />
              )
            }
            if (f.type === 'select' && f.options) {
              const useNum = f.options.some((o) => typeof o.value === 'number')
              return (
                <TextField
                  key={f.name}
                  select
                  label={f.label}
                  required={f.required}
                  disabled={f.disabled}
                  value={String(values[f.name] ?? '')}
                  onChange={(e) => {
                    const raw = e.target.value
                    set(f.name, useNum && raw !== '' ? Number(raw) : raw)
                  }}
                >
                  {f.options.map((o) => (
                    <MenuItem key={String(o.value)} value={String(o.value)}>
                      {o.label}
                    </MenuItem>
                  ))}
                </TextField>
              )
            }
            return (
              <TextField
                key={f.name}
                label={f.label}
                required={f.required}
                disabled={f.disabled}
                type={f.type === 'number' ? 'number' : f.type === 'password' ? 'password' : 'text'}
                value={values[f.name] ?? ''}
                slotProps={{
                  htmlInput:
                    f.type === 'number' && f.min !== undefined ? { min: f.min } : {},
                }}
                onChange={(e) =>
                  set(
                    f.name,
                    f.type === 'number'
                      ? (e.target.value === '' ? 0 : Number(e.target.value))
                      : e.target.value,
                  )
                }
              />
            )
          })}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Отмена</Button>
        <Button variant="contained" onClick={() => void handleSave()} disabled={submitting}>
          Сохранить
        </Button>
      </DialogActions>
    </Dialog>
  )
}
