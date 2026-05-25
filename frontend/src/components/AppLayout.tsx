import type { ReactNode } from 'react'
import {
  Drawer,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  AppBar,
  Box,
  Typography,
  CssBaseline,
  Button,
  IconButton,
} from '@mui/material'
import {
  DirectionsBus,
  CreditCard,
  SwapHoriz,
  VpnKey,
  People,
  Menu as MenuIcon,
} from '@mui/icons-material'
import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const drawerWidth = 220

interface NavEntry {
  to: string
  label: string
  icon: ReactNode
  adminOnly?: boolean
}

const navEntries: NavEntry[] = [
  { to: '/terminals', label: 'Терминалы', icon: <DirectionsBus /> },
  { to: '/cards', label: 'Карты', icon: <CreditCard /> },
  { to: '/transactions', label: 'Транзакции', icon: <SwapHoriz /> },
  { to: '/keys', label: 'Ключи', icon: <VpnKey />, adminOnly: true },
  { to: '/users', label: 'Пользователи', icon: <People /> },
]

export function AppLayout() {
  const { user, logout } = useAuth()
  const [mobileOpen, setMobileOpen] = useState(false)

  const drawer = (
    <div>
      <Toolbar>RPO Admin</Toolbar>
      <List>
        {navEntries.map((entry) => {
          if (entry.adminOnly && !user?.is_admin) {
            return null
          }
          return (
            <ListItemButton
              key={entry.to}
              component={NavLink}
              to={entry.to}
              onClick={() => setMobileOpen(false)}
              sx={{ '&.active': { bgcolor: 'action.selected' } }}
            >
              <ListItemIcon>{entry.icon}</ListItemIcon>
              <ListItemText primary={entry.label} />
            </ListItemButton>
          )
        })}
      </List>
    </div>
  )

  return (
    <Box sx={{ display: 'flex' }}>
      <CssBaseline />
      <AppBar position="fixed" sx={{ zIndex: (t) => t.zIndex.drawer + 1 }}>
        <Toolbar sx={{ gap: 1 }}>
          <IconButton
            color="inherit"
            edge="start"
            sx={{ display: { sm: 'none' } }}
            onClick={() => setMobileOpen((v) => !v)}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            Транспортные карты
          </Typography>
          <Typography variant="body2" sx={{ display: { xs: 'none', sm: 'block' } }}>
            {user?.login}
            {user?.is_admin ? ' (admin)' : ''}
          </Typography>
          <Button color="inherit" onClick={logout}>
            Выход
          </Button>
        </Toolbar>
      </AppBar>
      <Box component="nav" sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}>
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          ModalProps={{ keepMounted: true }}
          sx={{ display: { xs: 'block', sm: 'none' } }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          mt: 8,
        }}
      >
        <Outlet />
      </Box>
    </Box>
  )
}
