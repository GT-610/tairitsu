import { Box, Button, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material'
import { useTranslation } from '../../i18n'
import SettingsSectionCard from './SettingsSectionCard'
import type { DNSSettingsDraft } from './types'

interface DNSSettingsSectionProps {
  saving: boolean;
  initialValue: DNSSettingsDraft;
  draftValue: DNSSettingsDraft;
  onChange: (next: DNSSettingsDraft) => void;
  onReset: () => void;
  onSave: () => void;
  onAddServer: () => void;
  onRemoveServer: (index: number) => void;
}

function DNSSettingsSection({
  saving,
  initialValue,
  draftValue,
  onChange,
  onReset,
  onSave,
  onAddServer,
  onRemoveServer,
}: DNSSettingsSectionProps) {
  const { translateText } = useTranslation()
  const unsaved =
    draftValue.domain !== initialValue.domain ||
    JSON.stringify(draftValue.servers) !== JSON.stringify(initialValue.servers)

  return (
    <SettingsSectionCard title="DNS 设置" unsaved={unsaved}>
      <Typography variant="body1" sx={{ mb: 3 }}>
        {translateText('为网络内的自定义域名解析配置 DNS。每个网络只允许一个搜索域，但可以配置多个 DNS 服务器。')}
      </Typography>

      <TextField
        fullWidth
        label={translateText('搜索域')}
        placeholder={translateText('例如 home.arpa')}
        value={draftValue.domain}
        onChange={(e) => onChange({ ...draftValue, domain: e.target.value })}
        sx={{ mb: 3 }}
      />

      <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: '1fr 160px' }, mb: 2 }}>
        <TextField
          fullWidth
          label={translateText('DNS 服务器')}
          placeholder={translateText('例如 1.1.1.1 或 fd00::53')}
          value={draftValue.serverDraft}
          onChange={(e) => onChange({ ...draftValue, serverDraft: e.target.value })}
        />
        <Button fullWidth variant="outlined" onClick={onAddServer} sx={{ height: '100%' }}>
          {translateText('添加 DNS')}
        </Button>
      </Box>

      {draftValue.servers.length > 0 ? (
        <TableContainer component={Paper} variant="outlined" sx={{ mb: 3 }}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>{translateText('服务器')}</TableCell>
                <TableCell align="right">{translateText('操作')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {draftValue.servers.map((server, index) => (
                <TableRow key={server}>
                  <TableCell>{server}</TableCell>
                  <TableCell align="right">
                    <Button variant="outlined" color="error" size="small" onClick={() => onRemoveServer(index)}>{translateText('删除')}</Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>{translateText('尚未配置 DNS 服务器。')}</Typography>
      )}

      <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
        <Button variant="outlined" onClick={onReset} disabled={saving || !unsaved}>{translateText('重置更改')}</Button>
        <Button variant="contained" color="primary" onClick={onSave} disabled={saving || !unsaved}>{translateText('保存')}</Button>
      </Box>
    </SettingsSectionCard>
  )
}

export default DNSSettingsSection
