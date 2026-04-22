import { Box, Button, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material'
import SettingsSectionCard from './SettingsSectionCard'
import type { ManagedRoutesSettingsDraft } from './types'

interface ManagedRoutesSectionProps {
  saving: boolean;
  initialValue: ManagedRoutesSettingsDraft;
  draftValue: ManagedRoutesSettingsDraft;
  onChange: (next: ManagedRoutesSettingsDraft) => void;
  onReset: () => void;
  onSave: () => void;
  onAddRoute: () => void;
  onRemoveRoute: (index: number) => void;
}

function ManagedRoutesSection({
  saving,
  initialValue,
  draftValue,
  onChange,
  onReset,
  onSave,
  onAddRoute,
  onRemoveRoute,
}: ManagedRoutesSectionProps) {
  const unsaved = JSON.stringify(draftValue.routes) !== JSON.stringify(initialValue.routes)

  return (
    <SettingsSectionCard title="Managed Routes" unsaved={unsaved}>
      <Typography variant="body1" sx={{ mb: 3 }}>
        定义该网络还能到达的附加网段。主 IPv4/IPv6 子网由上面的分配区块维护，这里只管理额外的托管路由。
      </Typography>

      <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: '1.5fr 1fr 160px' }, mb: 2 }}>
        <TextField
          fullWidth
          label="目标网络 (CIDR)"
          placeholder="例如 10.1.2.0/24 或 fd00:1::/64"
          value={draftValue.routeDraft.target}
          onChange={(e) => onChange({ ...draftValue, routeDraft: { ...draftValue.routeDraft, target: e.target.value } })}
        />
        <TextField
          fullWidth
          label="下一跳地址"
          placeholder="可选"
          value={draftValue.routeDraft.via || ''}
          onChange={(e) => onChange({ ...draftValue, routeDraft: { ...draftValue.routeDraft, via: e.target.value } })}
        />
        <Button fullWidth variant="outlined" onClick={onAddRoute} sx={{ height: '100%' }}>
          Add Route
        </Button>
      </Box>

      {draftValue.routes.length > 0 ? (
        <TableContainer component={Paper} variant="outlined" sx={{ mb: 3 }}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Destination (CIDR)</TableCell>
                <TableCell>Via</TableCell>
                <TableCell align="right">Action</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {draftValue.routes.map((route, index) => (
                <TableRow key={`${route.target}-${route.via || 'direct'}`}>
                  <TableCell>{route.target}</TableCell>
                  <TableCell>{route.via || '-'}</TableCell>
                  <TableCell align="right">
                    <Button variant="outlined" color="error" size="small" onClick={() => onRemoveRoute(index)}>删除</Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>尚未配置额外托管路由。</Typography>
      )}

      <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
        <Button variant="outlined" onClick={onReset} disabled={saving || !unsaved}>重置更改</Button>
        <Button variant="contained" color="primary" onClick={onSave} disabled={saving || !unsaved}>保存</Button>
      </Box>
    </SettingsSectionCard>
  )
}

export default ManagedRoutesSection
