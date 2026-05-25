import { Box, Button, Paper, TextField, Typography, Alert } from '@mui/material'
import { useState } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { ApiError } from '../api/client'

export function LoginPage() {
  const { user, login, loading } = useAuth()
  const location = useLocation()
  const from = (location.state as { from?: { pathname: string } })?.from?.pathname ?? '/terminals'

  const [loginId, setLoginId] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  if (!loading && user) {
    return <Navigate to={from} replace />
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErr(null)
    setPending(true)
    try {
      await login(loginId, password)
    } catch (e2) {
      setErr(e2 instanceof ApiError ? e2.message : 'Ошибка входа')
    } finally {
      setPending(false)
    }
  }

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        bgcolor: 'grey.100',
      }}
    >
      <Paper sx={{ p: 4, width: 360, maxWidth: '90%' }} component="form" onSubmit={submit}>
        <Typography variant="h5" gutterBottom>
          Вход
        </Typography>
        {err && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {err}
          </Alert>
        )}
        <TextField
          label="Логин"
          fullWidth
          margin="normal"
          value={loginId}
          onChange={(e) => setLoginId(e.target.value)}
          autoComplete="username"
        />
        <TextField
          label="Пароль"
          type="password"
          fullWidth
          margin="normal"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
        />
        <Button type="submit" variant="contained" fullWidth sx={{ mt: 2 }} disabled={pending}>
          Войти
        </Button>
      </Paper>
    </Box>
  )
}
