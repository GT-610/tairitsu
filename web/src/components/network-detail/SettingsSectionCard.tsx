import { Box, Paper, Typography, type PaperProps } from '@mui/material'
import type { ReactNode } from 'react'
import { useTranslation } from '../../i18n'

interface SettingsSectionCardProps extends PaperProps {
  title: string;
  unsaved?: boolean;
  children: ReactNode;
}

function SettingsSectionCard({ title, unsaved = false, children, ...paperProps }: SettingsSectionCardProps) {
  const { translateText } = useTranslation()

  return (
    <Paper elevation={3} sx={{ p: 3, mb: 4, border: unsaved ? '2px solid' : 'none', borderColor: 'warning.main', ...(paperProps.sx || {}) }} {...paperProps}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">
          {translateText(title)}
        </Typography>
        {unsaved && (
          <Typography variant="body2" color="warning.main">
            {translateText('未保存')}
          </Typography>
        )}
      </Box>
      {children}
    </Paper>
  )
}

export default SettingsSectionCard
