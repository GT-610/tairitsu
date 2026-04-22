import { Box, Button, Grid, TextField } from '@mui/material'
import SettingsSectionCard from './SettingsSectionCard'
import type { BasicSettingsDraft } from './types'

interface NetworkBasicSettingsSectionProps {
  saving: boolean;
  initialValue: BasicSettingsDraft;
  draftValue: BasicSettingsDraft;
  onChange: (next: BasicSettingsDraft) => void;
  onReset: () => void;
  onSave: () => void;
}

function NetworkBasicSettingsSection({ saving, initialValue, draftValue, onChange, onReset, onSave }: NetworkBasicSettingsSectionProps) {
  const unsaved = draftValue.name !== initialValue.name || draftValue.description !== initialValue.description

  return (
    <SettingsSectionCard title="网络基本信息" unsaved={unsaved}>
      <Grid container spacing={3}>
        <Grid size={{ xs: 12 }}>
          <TextField
            fullWidth
            label="网络名称"
            variant="outlined"
            value={draftValue.name}
            onChange={(e) => onChange({ ...draftValue, name: e.target.value })}
            sx={{ mb: 2 }}
          />
        </Grid>
        <Grid size={{ xs: 12 }}>
          <TextField
            fullWidth
            label="网络描述"
            variant="outlined"
            multiline
            rows={3}
            value={draftValue.description}
            onChange={(e) => onChange({ ...draftValue, description: e.target.value })}
            sx={{ mb: 2 }}
          />
        </Grid>
        <Grid size={{ xs: 12 }}>
          <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
            <Button variant="outlined" onClick={onReset} disabled={saving || !unsaved}>重置更改</Button>
            <Button variant="contained" color="primary" onClick={onSave} disabled={saving || !unsaved}>保存</Button>
          </Box>
        </Grid>
      </Grid>
    </SettingsSectionCard>
  )
}

export default NetworkBasicSettingsSection
