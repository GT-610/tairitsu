import type { SxProps, Theme } from '@mui/material/styles'

export function getNavigationMessage(state: unknown): string {
  if (!state || typeof state !== 'object' || !('message' in state)) {
    return ''
  }

  const { message } = state as { message?: unknown }
  return typeof message === 'string' ? message : ''
}

export const summaryCardSx: SxProps<Theme> = {
  height: '100%',
  bgcolor: 'background.paper',
  border: 1,
  borderColor: 'divider',
  display: 'flex',
  flexDirection: 'column',
}
