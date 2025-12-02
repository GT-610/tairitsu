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
  Dialog, 
  DialogActions, 
  DialogContent, 
  DialogContentText, 
  DialogTitle, 
  TextField, 
  FormControlLabel, 
  Switch, 
  Snackbar, 
  Alert, 
  CircularProgress,
  IconButton
} from '@mui/material';
import { Add, Edit, Delete, Search, Refresh, ArrowBack } from '@mui/icons-material';
import { useParams, Link } from 'react-router-dom';
import { memberAPI, Member as ApiMember } from '../services/api';

// 格式化后的成员类型定义
interface Member {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'manager' | 'member';
  isActive: boolean;
  joinedAt: string;
}

// 表单数据类型定义
interface FormData {
  username: string;
  email: string;
  role: 'admin' | 'manager' | 'member';
  isActive: boolean;
}

// Snackbar状态类型定义
interface SnackbarState {
  open: boolean;
  message: string;
  severity: 'success' | 'error' | 'info' | 'warning';
}

// 确认对话框状态类型定义
interface ConfirmDialogState {
  open: boolean;
  memberId: string | null;
}

function Members() {
  const { networkId } = useParams<{ networkId: string }>();
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [openDialog, setOpenDialog] = useState<boolean>(false);
  const [currentMember, setCurrentMember] = useState<Member | null>(null);
  const [isEditing, setIsEditing] = useState<boolean>(false);
  const [formData, setFormData] = useState<FormData>({
    username: '',
    email: '',
    role: 'member',
    isActive: true
  });
  const [snackbar, setSnackbar] = useState<SnackbarState>({ open: false, message: '', severity: 'success' });
  const [confirmDialog, setConfirmDialog] = useState<ConfirmDialogState>({ open: false, memberId: null });

  // 获取成员数据
  useEffect(() => {
    const fetchMembers = async () => {
      try {
        setLoading(true);
        if (!networkId) {
          throw new Error('网络ID不能为空');
        }
        const response = await memberAPI.getMembers(networkId);
        // 格式化从API获取的数据以匹配前端组件的期望格式
        const formattedMembers: Member[] = response.data.map((member: ApiMember) => ({
          id: member.id || member.nodeId,
          username: member.name || member.nodeId,
          email: `${member.nodeId}@zt.local`, // ZeroTier成员通常没有真实邮箱
          role: (member.role as 'admin' | 'manager' | 'member') || 'member',
          isActive: member.authorized || false,
          joinedAt: member.createdAt ? new Date(member.createdAt).toLocaleDateString() : '未知'
        }));
        setMembers(formattedMembers);
      } catch (error: any) {
        console.error('获取成员列表失败:', error);
        setSnackbar({ 
          open: true, 
          message: '获取成员列表失败: ' + (error.response?.data?.message || error.message), 
          severity: 'error' 
        });
      } finally {
        setLoading(false);
      }
    };

    if (networkId) {
      fetchMembers();
    }
  }, [networkId]);

  // 处理搜索
  const filteredMembers = members.filter(member => 
    member.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    member.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // 处理添加成员
  const handleAddMember = () => {
    setCurrentMember(null);
    setIsEditing(false);
    setFormData({
      username: '',
      email: '',
      role: 'member',
      isActive: true
    });
    setOpenDialog(true);
  };

  // 处理编辑成员
  const handleEditMember = (member: Member) => {
    setCurrentMember(member);
    setIsEditing(true);
    setFormData({
      username: member.username,
      email: member.email,
      role: member.role,
      isActive: member.isActive
    });
    setOpenDialog(true);
  };

  // 处理删除确认
  const handleDeleteConfirm = (memberId: string) => {
    setConfirmDialog({ open: true, memberId });
  };

  // 处理删除成员
  const handleDeleteMember = async () => {
    try {
      if (!networkId || !confirmDialog.memberId) {
        throw new Error('网络ID或成员ID不能为空');
      }
      await memberAPI.deleteMember(networkId, confirmDialog.memberId);
      
      setMembers(prevMembers => 
        prevMembers.filter(member => member.id !== confirmDialog.memberId)
      );
      
      setSnackbar({ 
        open: true, 
        message: '成员删除成功', 
        severity: 'success' 
      });
    } catch (error: any) {
      console.error('删除成员失败:', error);
      setSnackbar({ 
        open: true, 
        message: '成员删除失败: ' + (error.response?.data?.message || error.message), 
        severity: 'error' 
      });
    } finally {
      setConfirmDialog({ open: false, memberId: null });
    }
  };

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      if (!networkId) {
        throw new Error('网络ID不能为空');
      }
      if (isEditing && currentMember) {
        // 更新现有成员
        const updatedData = {
          authorized: formData.isActive,
          name: formData.username
        };
        
        const response = await memberAPI.updateMember(networkId, currentMember.id, updatedData);
        
        // 更新本地状态
        const updatedMember: Member = {
          ...currentMember,
          ...formData,
          id: response.data.id || currentMember.id
        };
        
        setMembers(prevMembers => 
          prevMembers.map(member => 
            member.id === currentMember.id 
              ? updatedMember
              : member
          )
        );
        
        setSnackbar({ 
          open: true, 
          message: '成员信息更新成功', 
          severity: 'success' 
        });
      } else {
        // 添加新成员（在ZeroTier中实际上是授权一个新成员）
        // 这里我们简化处理，显示一个提示信息
        setSnackbar({ 
          open: true, 
          message: '在ZeroTier网络中，成员需要先加入网络，然后才能被授权。', 
          severity: 'info' 
        });
      }
      
      setOpenDialog(false);
    } catch (error: any) {
      console.error('成员操作失败:', error);
      setSnackbar({ 
        open: true, 
        message: (isEditing ? '成员信息更新失败: ' : '成员操作失败: ') + (error.response?.data?.message || error.message), 
        severity: 'error' 
      });
    }
  };

  // 处理状态切换
  const handleStatusToggle = async (memberId: string, newStatus: boolean) => {
    try {
      if (!networkId) {
        throw new Error('网络ID不能为空');
      }
      // 调用API更新成员状态
      const updatedData = {
        authorized: newStatus
      };
      
      await memberAPI.updateMember(networkId, memberId, updatedData);
      
      // 更新本地状态
      setMembers(prevMembers => 
        prevMembers.map(member => 
          member.id === memberId 
            ? { ...member, isActive: newStatus }
            : member
        )
      );
      
      setSnackbar({ 
        open: true, 
        message: `成员状态已${newStatus ? '启用' : '禁用'}`, 
        severity: 'success' 
      });
    } catch (error: any) {
      console.error('状态更新失败:', error);
      setSnackbar({ 
        open: true, 
        message: '状态更新失败: ' + (error.response?.data?.message || error.message), 
        severity: 'error' 
      });
    }
  };

  // 刷新成员列表
  const handleRefresh = async () => {
    try {
      if (!networkId) {
        throw new Error('网络ID不能为空');
      }
      setLoading(true);
      const response = await memberAPI.getMembers(networkId);
      // 格式化从API获取的数据以匹配前端组件的期望格式
      const formattedMembers: Member[] = response.data.map((member: ApiMember) => ({
        id: member.id || member.nodeId,
        username: member.name || member.nodeId,
        email: `${member.nodeId}@zt.local`, // ZeroTier成员通常没有真实邮箱
        role: (member.role as 'admin' | 'manager' | 'member') || 'member',
        isActive: member.authorized || false,
        joinedAt: member.createdAt ? new Date(member.createdAt).toLocaleDateString() : '未知'
      }));
      setMembers(formattedMembers);
      setSnackbar({ 
        open: true, 
        message: '成员列表已刷新', 
        severity: 'success' 
      });
    } catch (error: any) {
      console.error('刷新成员列表失败:', error);
      setSnackbar({ 
        open: true, 
        message: '刷新成员列表失败: ' + (error.response?.data?.message || error.message), 
        severity: 'error' 
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', gap: 2, alignItems: 'center', mb: 3 }}>
        <IconButton 
          onClick={() => window.history.back()}
          size="large"
        >
          <ArrowBack />
        </IconButton>
        <Typography variant="h4" component="h1">
          网络成员管理
        </Typography>
      </Box>
      
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Button 
            variant="contained" 
            startIcon={<Add />}
            onClick={handleAddMember}
          >
            添加成员
          </Button>
          <Button 
            variant="outlined" 
            startIcon={<Refresh />}
            onClick={handleRefresh}
            disabled={loading}
          >
            刷新
          </Button>
        </Box>
        
        <Box sx={{ position: 'relative', maxWidth: 300 }}>
          <Search sx={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', color: 'text.secondary' }} />
          <TextField
            placeholder="搜索用户名或邮箱"
            variant="outlined"
            fullWidth
            value={searchTerm}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: <span style={{ width: 24 }} />,
            }}
            size="small"
          />
        </Box>
      </Box>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table sx={{ minWidth: 650 }}>
            <TableHead>
              <TableRow>
                <TableCell>用户名</TableCell>
                <TableCell>邮箱</TableCell>
                <TableCell>角色</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>加入时间</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredMembers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                    <Typography variant="body1" color="text.secondary">
                      {searchTerm ? '没有找到匹配的成员' : '暂无成员'}
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                filteredMembers.map((member) => (
                  <TableRow
                    key={member.id}
                    sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                  >
                    <TableCell component="th" scope="row">
                      {member.username}
                    </TableCell>
                    <TableCell>{member.email}</TableCell>
                    <TableCell>
                      {member.role === 'admin' && '管理员'}
                      {member.role === 'manager' && '管理者'}
                      {member.role === 'member' && '普通成员'}
                    </TableCell>
                    <TableCell>
                      <FormControlLabel
                        control={
                          <Switch
                            checked={member.isActive}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleStatusToggle(member.id, e.target.checked)}
                          />
                        }
                        label={member.isActive ? '已启用' : '已禁用'}
                      />
                    </TableCell>
                    <TableCell>{member.joinedAt}</TableCell>
                    <TableCell align="right">
                      <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
                        <Button 
                          size="small" 
                          startIcon={<Edit />}
                          onClick={() => handleEditMember(member)}
                        >
                          编辑
                        </Button>
                        <Button 
                          size="small" 
                          startIcon={<Delete />}
                          color="error"
                          onClick={() => handleDeleteConfirm(member.id)}
                        >
                          删除
                        </Button>
                      </Box>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* 添加/编辑成员对话框 */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{isEditing ? '编辑成员' : '添加成员'}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {isEditing ? '请编辑成员信息' : '请输入新成员信息'}
          </DialogContentText>
          
          <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <TextField
              label="用户名"
              fullWidth
              required
              value={formData.username}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, username: e.target.value })}
            />
            
            <TextField
              label="邮箱"
              type="email"
              fullWidth
              required
              value={formData.email}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, email: e.target.value })}
            />
            
            <TextField
              label="角色"
              select
              fullWidth
              value={formData.role}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, role: e.target.value as 'admin' | 'manager' | 'member' })}
              SelectProps={{
                native: true,
              }}
            >
              <option value="member">普通成员</option>
              <option value="manager">管理者</option>
              <option value="admin">管理员</option>
            </TextField>
            
            <FormControlLabel
              control={
                <Switch
                  checked={formData.isActive}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, isActive: e.target.checked })}
                />
              }
              label="启用成员"
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          <Button onClick={handleSubmit} variant="contained">
            {isEditing ? '更新' : '添加'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* 删除确认对话框 */}
      <Dialog
        open={confirmDialog.open}
        onClose={() => setConfirmDialog({ ...confirmDialog, open: false })}
      >
        <DialogTitle>确认删除</DialogTitle>
        <DialogContent>
          <DialogContentText>
            确定要删除此成员吗？此操作不可撤销。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmDialog({ ...confirmDialog, open: false })}>
            取消
          </Button>
          <Button onClick={handleDeleteMember} color="error">
            删除
          </Button>
        </DialogActions>
      </Dialog>

      {/* 提示消息 */}
      <Snackbar 
        open={snackbar.open} 
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert 
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default Members;