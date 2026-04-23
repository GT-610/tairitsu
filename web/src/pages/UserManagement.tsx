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
  Snackbar
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { User, userAPI, authAPI } from '../services/api';
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
        <Typography variant="body2" color="text.secondary">
          当前系统仅保留一个管理员。你可以将管理员身份转让给某个普通用户，转让后自己会自动降为普通用户。
        </Typography>
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
                    <Button
                      variant="contained"
                      color="primary"
                      onClick={() => { void handleTransferAdmin(user); }}
                      disabled={updating}
                    >
                      转让管理员
                    </Button>
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
    </Box>
  );
}

export default UserManagement
