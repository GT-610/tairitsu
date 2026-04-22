import { Box, Button, FormControlLabel, Switch, TextField, Typography } from '@mui/material'
import SettingsSectionCard from './SettingsSectionCard'
import type { MulticastSettingsDraft } from './types'

interface MulticastSettingsSectionProps {
  saving: boolean;
  initialValue: MulticastSettingsDraft;
  draftValue: MulticastSettingsDraft;
  onChange: (next: MulticastSettingsDraft) => void;
  onReset: () => void;
  onSave: () => void;
}

function MulticastSettingsSection({
  saving,
  initialValue,
  draftValue,
  onChange,
  onReset,
  onSave,
}: MulticastSettingsSectionProps) {
  const unsaved =
    draftValue.multicastLimit !== initialValue.multicastLimit ||
    draftValue.enableBroadcast !== initialValue.enableBroadcast

  return (
    <SettingsSectionCard title="多播设置" unsaved={unsaved}>
      <Box sx={{ display: 'grid', gap: 3, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Typography variant="body1">多播接收者限制</Typography>
          <TextField
            type="number"
            size="small"
            value={draftValue.multicastLimit}
            onChange={(e) => onChange({ ...draftValue, multicastLimit: parseInt(e.target.value, 10) || 0 })}
            sx={{ width: 120 }}
          />
        </Box>
        <FormControlLabel
          control={<Switch checked={draftValue.enableBroadcast} onChange={(e) => onChange({ ...draftValue, enableBroadcast: e.target.checked })} />}
          label="启用广播"
        />
      </Box>

      <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2, mt: 3 }}>
        <Button variant="outlined" onClick={onReset} disabled={saving || !unsaved}>重置更改</Button>
        <Button variant="contained" color="primary" onClick={onSave} disabled={saving || !unsaved}>保存</Button>
      </Box>
    </SettingsSectionCard>
  )
}

export default MulticastSettingsSection
