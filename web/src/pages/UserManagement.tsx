import React, { useState, useEffect } from 'react';
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
import { User, userAPI, authAPI } from '../services/api';

function UserManagement() {
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
      } catch (error) {
        console.error('获取用户数据失败:', error);
        setMessage({ text: '获取用户数据失败', severity: 'error' });
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // 处理用户角色更新
  const handleUpdateUserRole = async (userId: string, currentRole: 'admin' | 'user') => {
    try {
      setUpdating(true);
      
      const newRole = currentRole === 'admin' ? 'user' : 'admin';
      await userAPI.updateUserRole(userId, newRole);
      
      // 更新本地用户列表
      setUsers(prevUsers => 
        prevUsers.map(user => 
          user.id === userId ? { ...user, role: newRole } : user
        )
      );
      
      setMessage({
        text: `成功${newRole === 'admin' ? '赋予' : '撤销'}管理员权限`,
        severity: 'success'
      });
    } catch (error) {
      console.error('更新用户角色失败:', error);
      setMessage({ text: '更新用户角色失败', severity: 'error' });
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
      <Typography variant="h4" component="h1" gutterBottom>
        用户管理
      </Typography>

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
              <TableCell>邮箱</TableCell>
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
                <TableCell>{user.email}</TableCell>
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
                <TableCell>{new Date(user.createdAt).toLocaleString()}</TableCell>
                <TableCell>
                  {/* 不能对自己进行操作 */}
                  {currentUser && user.id !== currentUser.id ? (
                    <Button
                      variant="contained"
                      color={user.role === 'admin' ? 'error' : 'primary'}
                      onClick={() => handleUpdateUserRole(user.id, user.role)}
                      disabled={updating}
                    >
                      {user.role === 'admin' ? '撤销管理员' : '赋予管理员'}
                    </Button>
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