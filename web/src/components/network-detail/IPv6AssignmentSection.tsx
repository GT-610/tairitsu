import { Alert, Box, Button, Paper, Switch, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material'
import { useTranslation } from '../../i18n'
import type { IpAssignmentPool } from '../../services/api'
import SettingsSectionCard from './SettingsSectionCard'
import type { IPv6SettingsDraft } from './types'

interface IPv6AssignmentSectionProps {
  saving: boolean;
  initialValue: IPv6SettingsDraft;
  draftValue: IPv6SettingsDraft;
  draftRangeIssue: string | null;
  configurationIssues: string[];
  subnetValid: boolean;
  onChange: (next: IPv6SettingsDraft) => void;
  onReset: () => void;
  onSave: () => void;
  onSetRange: () => void;
  onRemoveRange: (index: number) => void;
}

function IPv6AssignmentSection({
  saving,
  initialValue,
  draftValue,
  draftRangeIssue,
  configurationIssues,
  subnetValid,
  onChange,
  onReset,
  onSave,
  onSetRange,
  onRemoveRange,
}: IPv6AssignmentSectionProps) {
  const { translateText } = useTranslation()
  const unsaved =
    draftValue.subnet !== initialValue.subnet ||
    draftValue.customAssign !== initialValue.customAssign ||
    draftValue.rfc4193 !== initialValue.rfc4193 ||
    draftValue.plane6 !== initialValue.plane6 ||
    JSON.stringify(draftValue.pools) !== JSON.stringify(initialValue.pools)

  return (
    <SettingsSectionCard title="IPv6分配" unsaved={unsaved}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="body1" sx={{ mb: 2 }}>
          {translateText('IPv6 默认不分配。只有自定义 IPv6 范围依赖手动填写子网；RFC4193 和 6PLANE 由控制器自动派生。')}
        </Typography>
        <TextField
          fullWidth
          label={translateText('IPv6 子网')}
          placeholder={translateText('例如 fd00::/48')}
          value={draftValue.subnet}
          onChange={(e) => onChange({ ...draftValue, subnet: e.target.value })}
        />
      </Box>

      <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, mb: draftValue.customAssign ? 2 : 0, flexWrap: 'wrap' }}>
          <Box>
            <Typography variant="h6">{translateText('从自定义 IPv6 范围分配')}</Typography>
            <Typography variant="body2" color="text.secondary">
              {translateText('需要先填写合法的 IPv6 子网，然后才能配置自定义 IPv6 地址池。')}
            </Typography>
          </Box>
          <Switch
            checked={draftValue.customAssign}
            disabled={!subnetValid}
            onChange={(e) => onChange({
              ...draftValue,
              customAssign: e.target.checked,
              ...(e.target.checked ? {} : { pools: [], poolStartDraft: '', poolEndDraft: '' }),
            })}
          />
        </Box>

        {draftValue.customAssign && (
          <>
            <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
              <TextField
                fullWidth
                label={translateText('起始 IPv6')}
                placeholder={translateText('例如 fd00::1000')}
                value={draftValue.poolStartDraft}
                onChange={(e) => onChange({ ...draftValue, poolStartDraft: e.target.value })}
                error={Boolean(draftValue.poolStartDraft || draftValue.poolEndDraft) && Boolean(draftRangeIssue)}
                helperText={draftValue.subnet ? `${translateText('必须落在 ')}${draftValue.subnet}${translateText(' 内')}` : translateText('请先填写 IPv6 子网')}
              />
              <TextField
                fullWidth
                label={translateText('结束 IPv6')}
                placeholder={translateText('例如 fd00::1fff')}
                value={draftValue.poolEndDraft}
                onChange={(e) => onChange({ ...draftValue, poolEndDraft: e.target.value })}
                error={Boolean(draftValue.poolStartDraft || draftValue.poolEndDraft) && Boolean(draftRangeIssue)}
                helperText={draftValue.subnet ? `${translateText('必须落在 ')}${draftValue.subnet}${translateText(' 内')}` : translateText('请先填写 IPv6 子网')}
              />
            </Box>

            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 2, flexWrap: 'wrap', mt: 2 }}>
              <Box>
                <Typography variant="subtitle1">{translateText('当前 IPv6 自动分配范围')}</Typography>
                <Typography variant="body2" color="text.secondary">{translateText('使用 `Set Range` 将当前输入加入 IPv6 地址池列表。')}</Typography>
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
              <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>{translateText('尚未配置 IPv6 地址池。')}</Typography>
            )}
          </>
        )}
      </Paper>

      <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, flexWrap: 'wrap' }}>
          <Box>
            <Typography variant="h6">{translateText('分配 RFC4193 唯一本地地址（每台设备 /128）')}</Typography>
            <Typography variant="body2" color="text.secondary">
              {translateText('自动为每个设备分配稳定的唯一本地 IPv6 地址。无需额外配置。')}
            </Typography>
          </Box>
          <Switch checked={draftValue.rfc4193} onChange={(e) => onChange({ ...draftValue, rfc4193: e.target.checked })} />
        </Box>
      </Paper>

      <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, flexWrap: 'wrap' }}>
          <Box>
            <Typography variant="h6">{translateText('分配 6PLANE 路由地址（每台设备 /80）')}</Typography>
            <Typography variant="body2" color="text.secondary">
              {translateText('为每个成员分配由节点 ID 派生的可路由 IPv6 地址。无需额外配置。')}
            </Typography>
          </Box>
          <Switch checked={draftValue.plane6} onChange={(e) => onChange({ ...draftValue, plane6: e.target.checked })} />
        </Box>
      </Paper>

      {configurationIssues.length > 0 && <Alert severity="warning" sx={{ mb: 3 }}>{configurationIssues.map(translateText).join(' ')}</Alert>}

      <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
        <Button variant="outlined" onClick={onReset} disabled={saving || !unsaved}>{translateText('重置更改')}</Button>
        <Button variant="contained" color="primary" onClick={onSave} disabled={saving || !unsaved}>{translateText('保存')}</Button>
      </Box>
    </SettingsSectionCard>
  )
}

export default IPv6AssignmentSection
