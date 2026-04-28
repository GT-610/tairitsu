import { Alert, Box, Button, Paper, Switch, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material'
import { useTranslation } from '../../i18n'
import type { IpAssignmentPool } from '../../services/api'
import SettingsSectionCard from './SettingsSectionCard'
import type { IPv4SettingsDraft } from './types'

interface IPv4AssignmentSectionProps {
  saving: boolean;
  initialValue: IPv4SettingsDraft;
  draftValue: IPv4SettingsDraft;
  draftRangeIssue: string | null;
  configurationIssues: string[];
  onChange: (next: IPv4SettingsDraft) => void;
  onReset: () => void;
  onSave: () => void;
  onGenerateSubnetAndRange: () => void;
  onSetRange: () => void;
  onRemoveRange: (index: number) => void;
}

function IPv4AssignmentSection({
  saving,
  initialValue,
  draftValue,
  draftRangeIssue,
  configurationIssues,
  onChange,
  onReset,
  onSave,
  onGenerateSubnetAndRange,
  onSetRange,
  onRemoveRange,
}: IPv4AssignmentSectionProps) {
  const { translateText } = useTranslation()
  const unsaved =
    draftValue.subnet !== initialValue.subnet ||
    draftValue.autoAssign !== initialValue.autoAssign ||
    JSON.stringify(draftValue.pools) !== JSON.stringify(initialValue.pools)

  return (
    <SettingsSectionCard title="IPv4分配" unsaved={unsaved}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="body1" sx={{ mb: 2 }}>
          {translateText('默认会以一个 IPv4 子网作为网络边界。自动分配地址池必须完全落在该子网内。')}
        </Typography>
        <TextField
          fullWidth
          label={translateText('IPv4 子网')}
          placeholder={translateText('例如 10.22.2.0/24')}
          value={draftValue.subnet}
          onChange={(e) => onChange({ ...draftValue, subnet: e.target.value })}
          sx={{ mb: 2 }}
        />
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center', flexWrap: 'wrap' }}>
          <Button variant="outlined" onClick={onGenerateSubnetAndRange}>{translateText('生成新的子网与范围')}</Button>
          <Typography variant="body2" color="text.secondary">
            {translateText('保存时会自动将该子网同步为网络的主托管路由。')}
          </Typography>
        </Box>
      </Box>

      <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, mb: draftValue.autoAssign ? 2 : 0, flexWrap: 'wrap' }}>
          <Box>
            <Typography variant="h6">{translateText('自动分配 IPv4 地址给新成员设备')}</Typography>
            <Typography variant="body2" color="text.secondary">
              {translateText('关闭后会收起自动分配范围，并在保存时清空 `ipAssignmentPools`。成员手动指定 IP 不受影响。')}
            </Typography>
          </Box>
          <Switch checked={draftValue.autoAssign} onChange={(e) => onChange({
            ...draftValue,
            autoAssign: e.target.checked,
            ...(e.target.checked ? {} : { pools: [], poolStartDraft: '', poolEndDraft: '' }),
          })} />
        </Box>

        {draftValue.autoAssign && (
          <>
            <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
              <TextField
                fullWidth
                label={translateText('起始 IPv4')}
                placeholder={translateText('例如 10.22.2.1')}
                value={draftValue.poolStartDraft}
                onChange={(e) => onChange({ ...draftValue, poolStartDraft: e.target.value })}
                error={Boolean(draftValue.poolStartDraft || draftValue.poolEndDraft) && Boolean(draftRangeIssue)}
                helperText={draftValue.subnet ? `${translateText('必须落在 ')}${draftValue.subnet}${translateText(' 内')}` : translateText('请先填写 IPv4 子网')}
              />
              <TextField
                fullWidth
                label={translateText('结束 IPv4')}
                placeholder={translateText('例如 10.22.2.254')}
                value={draftValue.poolEndDraft}
                onChange={(e) => onChange({ ...draftValue, poolEndDraft: e.target.value })}
                error={Boolean(draftValue.poolStartDraft || draftValue.poolEndDraft) && Boolean(draftRangeIssue)}
                helperText={draftValue.subnet ? `${translateText('必须落在 ')}${draftValue.subnet}${translateText(' 内')}` : translateText('请先填写 IPv4 子网')}
              />
            </Box>

            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 2, flexWrap: 'wrap', mt: 2 }}>
              <Box>
                <Typography variant="subtitle1">{translateText('当前 IPv4 自动分配范围')}</Typography>
                <Typography variant="body2" color="text.secondary">{translateText('使用 `Set Range` 将当前输入加入自动分配地址池列表。')}</Typography>
              </Box>
              <Button variant="outlined" onClick={onSetRange} disabled={Boolean(draftRangeIssue)}>{translateText('设置范围')}</Button>
            </Box>

            {Boolean(draftRangeIssue) && Boolean(draftValue.poolStartDraft || draftValue.poolEndDraft) && (
              <Alert severity="warning" sx={{ mt: 2 }}>{draftRangeIssue ? translateText(draftRangeIssue) : ''}</Alert>
            )}

            {draftValue.pools.length > 0 ? (
              <TableContainer component={Paper} variant="outlined" sx={{ mt: 2 }}>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>{translateText('起始')}</TableCell>
                      <TableCell>{translateText('结束')}</TableCell>
                      <TableCell align="right">{translateText('操作')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {draftValue.pools.map((pool: IpAssignmentPool, index: number) => (
                      <TableRow key={`${pool.ipRangeStart}-${pool.ipRangeEnd}`}>
                        <TableCell>{pool.ipRangeStart}</TableCell>
                        <TableCell>{pool.ipRangeEnd}</TableCell>
                        <TableCell align="right">
                          <Button variant="outlined" color="error" size="small" onClick={() => onRemoveRange(index)}>{translateText('删除')}</Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>{translateText('尚未配置自动分配地址池。')}</Typography>
            )}
          </>
        )}
      </Paper>

      {configurationIssues.length > 0 && <Alert severity="warning" sx={{ mb: 3 }}>{configurationIssues.map(translateText).join(' ')}</Alert>}

      <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
        <Button variant="outlined" onClick={onReset} disabled={saving || !unsaved}>{translateText('重置更改')}</Button>
        <Button variant="contained" color="primary" onClick={onSave} disabled={saving || !unsaved}>{translateText('保存')}</Button>
      </Box>
    </SettingsSectionCard>
  )
}

export default IPv4AssignmentSection
