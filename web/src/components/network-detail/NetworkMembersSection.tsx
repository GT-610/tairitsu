import { MoreHoriz } from '@mui/icons-material'
import { Alert, Box, Button, Card, CardContent, Chip, Grid, IconButton, Paper, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material'
import type { MouseEvent } from 'react'
import type { NetworkMemberDevice } from './types'

interface NetworkMembersSectionProps {
  memberDevices: NetworkMemberDevice[];
  pendingMembers: NetworkMemberDevice[];
  authorizedMembers: NetworkMemberDevice[];
  filteredMembers: NetworkMemberDevice[];
  memberSearchTerm: string;
  saving: boolean;
  hidePendingBanner: boolean;
  onMemberSearchTermChange: (value: string) => void;
  onHidePendingBanner: () => void;
  onQuickApprove: () => void;
  onQuickReject: () => void;
  onOpenMemberMenu?: (event: MouseEvent<HTMLElement>, member: NetworkMemberDevice) => void;
  readOnly?: boolean;
}

function NetworkMembersSection({
  memberDevices,
  pendingMembers,
  authorizedMembers,
  filteredMembers,
  memberSearchTerm,
  saving,
  hidePendingBanner,
  onMemberSearchTermChange,
  onHidePendingBanner,
  onQuickApprove,
  onQuickReject,
  onOpenMemberMenu,
  readOnly = false,
}: NetworkMembersSectionProps) {
  return (
    <>
      {pendingMembers.length > 0 && !hidePendingBanner && !readOnly && (
        <Alert
          severity="warning"
          sx={{ mb: 3 }}
          action={(
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
              <Button color="inherit" variant="outlined" size="small" disabled={saving} onClick={onQuickApprove}>
                授权第一个待审批成员
              </Button>
              <Button color="inherit" variant="text" size="small" disabled={saving} onClick={onQuickReject}>
                拒绝第一个待审批成员
              </Button>
              <Button color="inherit" size="small" onClick={onHidePendingBanner}>
                关闭
              </Button>
            </Stack>
          )}
        >
          当前有 {pendingMembers.length} 个待授权设备。你可以在此快速审批，也可以在下方成员列表中逐个处理。
        </Alert>
      )}

      <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, sm: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  设备总数
                </Typography>
                <Typography variant="h4">
                  {memberDevices.length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  已授权设备
                </Typography>
                <Typography variant="h4">
                  {authorizedMembers.length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  待授权设备
                </Typography>
                <Typography variant="h4">
                  {pendingMembers.length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Paper>

      <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h5">
            成员设备
          </Typography>
          <TextField
            size="small"
            placeholder="搜索设备名称或设备 ID"
            value={memberSearchTerm}
            onChange={(event) => onMemberSearchTermChange(event.target.value)}
            sx={{ width: { xs: '100%', sm: 280 } }}
          />
        </Box>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>设备 ID</TableCell>
                <TableCell>名称</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>Managed IPs</TableCell>
                <TableCell>ZT 版本</TableCell>
                {!readOnly && <TableCell align="right">操作</TableCell>}
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredMembers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={readOnly ? 5 : 6} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                    {memberSearchTerm ? '没有找到匹配的成员设备' : '暂无设备连接'}
                  </TableCell>
                </TableRow>
              ) : (
                filteredMembers.map((member) => (
                  <TableRow key={member.id} hover>
                    <TableCell>{member.id}</TableCell>
                    <TableCell>{member.name}</TableCell>
                    <TableCell>
                      <Chip
                        label={member.authorized ? '已授权' : '待授权'}
                        color={member.authorized ? 'success' : 'warning'}
                        variant={member.authorized ? 'filled' : 'outlined'}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>{member.ipAssignments.length > 0 ? member.ipAssignments.join(', ') : '-'}</TableCell>
                    <TableCell>{member.clientVersion}</TableCell>
                    {!readOnly && onOpenMemberMenu && (
                      <TableCell align="right">
                        <IconButton onClick={(event) => onOpenMemberMenu(event, member)}>
                          <MoreHoriz />
                        </IconButton>
                      </TableCell>
                    )}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </>
  )
}

export default NetworkMembersSection
