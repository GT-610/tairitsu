import { Box } from '@mui/material'
import { getUserRoleLabel } from '../utils/userPresentation'
import type { User } from '../services/api'

interface UserRoleBadgeProps {
  role?: User['role'];
}

function UserRoleBadge({ role }: UserRoleBadgeProps) {
  const isAdmin = role === 'admin'

  return (
    <Box
      sx={{
        display: 'inline-block',
        px: 1,
        py: 0.25,
        borderRadius: 1,
        backgroundColor: isAdmin ? '#e3f2fd' : '#f1f8e9',
        color: isAdmin ? '#1565c0' : '#388e3c',
        fontWeight: 'bold',
      }}
    >
      {getUserRoleLabel(role)}
    </Box>
  )
}

export default UserRoleBadge
