import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Button,
  Alert,
  CircularProgress,
  Snackbar,
  Stack,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { User, userAPI, authAPI, type ResetUserPasswordResponse, type CreateUserResponse, type DeleteUserResponse } from '../services/api';
import { getErrorMessage } from '../services/errors';
import { useAuth } from '../services/auth';

function UserManagement() {
  const navigate = useNavigate();
  const { refreshUser } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [updating, setUpdating] = useState<boolean>(false);
  const [message, setMessage] = useState<{ text: string; severity: 'success' | 'error' | 'info' } | null>(null);
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [createUsername, setCreateUsername] = useState('');
  const [createUsernameError, setCreateUsernameError] = useState('');
  const [createResult, setCreateResult] = useState<CreateUserResponse | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null);
  const [deleteResult, setDeleteResult] = useState<DeleteUserResponse | null>(null);
  const [resetTarget, setResetTarget] = useState<User | null>(null);
  const [resetResult, setResetResult] = useState<ResetUserPasswordResponse | null>(null);

  // 获取所有用户和当前用户信息
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // 获取当前用户信息
        const currentUserResponse = await authAPI.getProfile();
        setCurrentUser(currentUserResponse.data);
        
        // 获取所有用户
        const usersResponse = await userAPI.getAllUsers();
        setUsers(usersResponse.data);
      } catch (error: unknown) {
        setMessage({ 
          text: getErrorMessage(error, '获取用户数据失败'), 
          severity: 'error' 
        });
      } finally {
        setLoading(false);
      }
    };

    void fetchData();
  }, []);

  // 转让管理员身份
  const handleTransferAdmin = async (user: User) => {
    try {
      setUpdating(true);

      const response = await userAPI.transferAdmin(user.id);
      const nextAdmin = response.data.user;
      const refreshedProfile = await authAPI.getProfile();

      setUsers(prevUsers =>
        prevUsers.map((item) => {
          if (item.id === currentUser?.id) {
            return { ...item, role: 'user' };
          }
          if (item.id === nextAdmin.id) {
            return { ...item, role: 'admin' };
          }
          return item;
        })
      );
      setCurrentUser(refreshedProfile.data);
      refreshUser(refreshedProfile.data);
      void navigate('/networks', {
        replace: true,
        state: { message: `管理员身份已转让给 ${nextAdmin.username}` },
      });
    } catch (error) {
      setMessage({ text: getErrorMessage(error, '转让管理员身份失败'), severity: 'error' });
    } finally {
      setUpdating(false);
    }
  };

  const handleCreateUser = async () => {
    if (!createUsername.trim()) {
      setCreateUsernameError('请输入用户名');
      return;
    }

    try {
      setUpdating(true);
      const response = await userAPI.createUser({ username: createUsername.trim() });
      setUsers((previous) => [...previous, response.data.user]);
      setOpenCreateDialog(false);
      setCreateUsername('');
      setCreateUsernameError('');
      setCreateResult(response.data);
      setMessage({
        text: `已创建用户 ${response.data.user.username}，请立即通过其他方式告知其临时密码`,
        severity: 'success',
      });
    } catch (error: unknown) {
      setCreateUsernameError(getErrorMessage(error, '创建用户失败'));
    } finally {
      setUpdating(false);
    }
  };

  const handleResetPassword = async () => {
    if (!resetTarget) {
      return;
    }

    try {
      setUpdating(true);
      const response = await userAPI.resetPassword(resetTarget.id);
      setResetTarget(null);
      setResetResult(response.data);
      setMessage({
        text: `已为 ${response.data.user.username} 生成新的临时密码，并吊销 ${response.data.revoked_sessions} 个现有会话`,
        severity: 'success',
      });
    } catch (error: unknown) {
      setMessage({
        text: getErrorMessage(error, '重置用户密码失败'),
        severity: 'error',
      });
    } finally {
      setUpdating(false);
    }
  };

  const handleDeleteUser = async () => {
    if (!deleteTarget) {
      return;
    }

    try {
      setUpdating(true);
      const response = await userAPI.deleteUser(deleteTarget.id);
      setUsers((previous) => previous.filter((item) => item.id !== deleteTarget.id));
      setDeleteTarget(null);
      setDeleteResult(response.data);
      setMessage({
        text: `已删除用户 ${response.data.user.username}，并转移 ${response.data.transferred_networks} 个网络`,
        severity: 'success',
      });
    } catch (error: unknown) {
      setMessage({
        text: getErrorMessage(error, '删除用户失败'),
        severity: 'error',
      });
    } finally {
      setUpdating(false);
    }
  };

  // 关闭提示消息
  const handleCloseMessage = () => {
    setMessage(null);
  };

  if (loading) {
    return (
      <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '50vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          用户管理
        </Typography>
        <Stack direction="row" spacing={2} alignItems="center">
          <Typography variant="body2" color="text.secondary">
            当前系统仅保留一个管理员。你可以创建普通用户、重置普通用户密码，或将管理员身份转让给某个普通用户。
          </Typography>
          <Button variant="contained" onClick={() => setOpenCreateDialog(true)} disabled={updating}>
            创建用户
          </Button>
        </Stack>
      </Box>

      {message && (
        <Snackbar
          open={!!message}
          autoHideDuration={3000}
          onClose={handleCloseMessage}
          anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        >
          <Alert
            onClose={handleCloseMessage}
            severity={message.severity}
            sx={{ width: '100%' }}
          >
            {message.text}
          </Alert>
        </Snackbar>
      )}

      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="user management table">
          <TableHead>
            <TableRow>
              <TableCell>用户名</TableCell>
              <TableCell>角色</TableCell>
              <TableCell>创建时间</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((user) => (
              <TableRow
                key={user.id}
                sx={{
                  '&:last-child td, &:last-child th': { border: 0 }
                }}
              >
                <TableCell component="th" scope="row">
                  {user.username}
                </TableCell>
                <TableCell>
                  <Box
                    sx={{
                      display: 'inline-block',
                      padding: '2px 8px',
                      borderRadius: '4px',
                      backgroundColor: user.role === 'admin' ? '#e3f2fd' : '#f1f8e9',
                      color: user.role === 'admin' ? '#1565c0' : '#388e3c',
                      fontWeight: 'bold'
                    }}
                  >
                    {user.role === 'admin' ? '管理员' : '普通用户'}
                  </Box>
                </TableCell>
                <TableCell>{user.createdAt ? new Date(user.createdAt).toLocaleString() : '未知'}</TableCell>
                <TableCell>
                  {/* 不能对自己进行操作 */}
                  {currentUser && user.id !== currentUser.id && user.role !== 'admin' ? (
                    <Stack direction="row" spacing={1.5}>
                      <Button
                        variant="outlined"
                        color="error"
                        onClick={() => setDeleteTarget(user)}
                        disabled={updating}
                      >
                        删除用户
                      </Button>
                      <Button
                        variant="outlined"
                        color="warning"
                        onClick={() => setResetTarget(user)}
                        disabled={updating}
                      >
                        重置密码
                      </Button>
                      <Button
                        variant="contained"
                        color="primary"
                        onClick={() => { void handleTransferAdmin(user); }}
                        disabled={updating}
                      >
                        转让管理员
                      </Button>
                    </Stack>
                  ) : currentUser && user.id === currentUser.id ? (
                    <Typography variant="body2" color="text.secondary">
                      当前管理员
                    </Typography>
                  ) : user.role === 'admin' ? (
                    <Typography variant="body2" color="text.secondary">
                      已是管理员
                    </Typography>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={openCreateDialog} onClose={() => !updating && setOpenCreateDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>创建用户</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Alert severity="info">
              新用户会自动创建为普通用户，并生成一个一次性临时密码。
            </Alert>
            <TextField
              label="用户名"
              value={createUsername}
              onChange={(event) => {
                setCreateUsername(event.target.value)
                if (createUsernameError) {
                  setCreateUsernameError('')
                }
              }}
              fullWidth
              error={Boolean(createUsernameError)}
              helperText={createUsernameError || '建议使用可识别、便于通知的用户名'}
              disabled={updating}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenCreateDialog(false)} disabled={updating}>取消</Button>
          <Button variant="contained" onClick={() => { void handleCreateUser(); }} disabled={updating}>
            {updating ? '创建中...' : '确认创建'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(createResult)} onClose={() => setCreateResult(null)} maxWidth="sm" fullWidth>
        <DialogTitle>一次性临时密码</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Alert severity="warning">
              该密码只会展示这一次。关闭后将无法再次查看，请立即通过其他方式安全告知用户。
            </Alert>
            <Typography variant="body2" color="text.secondary">
              新用户：{createResult?.user.username || '未知用户'}
            </Typography>
            <TextField
              label="临时密码"
              value={createResult?.temporary_password || ''}
              fullWidth
              InputProps={{ readOnly: true }}
            />
            <Typography variant="body2" color="text.secondary">
              用户首次收到密码后，应尽快登录并到“设置”中修改为自己的新密码。
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={() => setCreateResult(null)}>
            我已记录
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(deleteTarget)} onClose={() => !updating && setDeleteTarget(null)} maxWidth="sm" fullWidth>
        <DialogTitle>删除用户</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Alert severity="error">
              删除 {deleteTarget?.username || '该用户'} 后，账号将不可恢复。
            </Alert>
            <Typography variant="body2" color="text.secondary">
              该用户拥有的网络将自动转让给当前管理员，ZeroTier 控制器内网络本身不会被删除。
            </Typography>
            <Typography variant="body2" color="text.secondary">
              该用户当前所有登录会话会立即失效，后续请求会被强制退出。
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} disabled={updating}>取消</Button>
          <Button
            variant="contained"
            color="error"
            onClick={() => { void handleDeleteUser(); }}
            disabled={updating}
          >
            {updating ? '删除中...' : '确认删除'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(deleteResult)} onClose={() => setDeleteResult(null)} maxWidth="sm" fullWidth>
        <DialogTitle>用户已删除</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Alert severity="success">
              {deleteResult?.user.username || '该用户'} 已被删除。
            </Alert>
            <Typography variant="body2" color="text.secondary">
              转移网络：{deleteResult?.transferred_networks ?? 0} 个
            </Typography>
            <Typography variant="body2" color="text.secondary">
              吊销会话：{deleteResult?.revoked_sessions ?? 0} 个
            </Typography>
            <Typography variant="body2" color="text.secondary">
              这些网络现在已归当前管理员所有，可在网络列表中继续管理。
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={() => setDeleteResult(null)}>
            我知道了
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(resetTarget)} onClose={() => !updating && setResetTarget(null)} maxWidth="sm" fullWidth>
        <DialogTitle>重置用户密码</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Alert severity="warning">
              将为 {resetTarget?.username || '该用户'} 生成一个新的随机密码，并立即吊销该用户当前所有登录会话。
            </Alert>
            <Typography variant="body2" color="text.secondary">
              新密码只会展示一次。请通过其他安全方式告知用户，并提醒其在登录后尽快到“设置”中修改密码。
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setResetTarget(null)} disabled={updating}>取消</Button>
          <Button
            variant="contained"
            color="warning"
            onClick={() => { void handleResetPassword(); }}
            disabled={updating}
          >
            {updating ? '重置中...' : '确认重置'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(resetResult)} onClose={() => setResetResult(null)} maxWidth="sm" fullWidth>
        <DialogTitle>一次性临时密码</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <Alert severity="warning">
              该密码只会展示这一次。关闭后将无法再次查看，请立即通过其他方式安全告知用户。
            </Alert>
            <Typography variant="body2" color="text.secondary">
              目标用户：{resetResult?.user.username || '未知用户'}
            </Typography>
            <TextField
              label="临时密码"
              value={resetResult?.temporary_password || ''}
              fullWidth
              InputProps={{ readOnly: true }}
            />
            <Typography variant="body2" color="text.secondary">
              已吊销会话：{resetResult?.revoked_sessions ?? 0} 个。用户收到密码后应尽快登录并修改为自己的新密码。
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={() => setResetResult(null)}>
            我已记录
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default UserManagement
