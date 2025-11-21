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
  CircularProgress
} from '@mui/material';
import { Add, Edit, Delete, Search, Refresh } from '@mui/icons-material';
import { useParams } from 'react-router-dom';

function Members() {
  const { networkId } = useParams();
  const [members, setMembers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [openDialog, setOpenDialog] = useState(false);
  const [currentMember, setCurrentMember] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    role: 'member',
    isActive: true
  });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [confirmDialog, setConfirmDialog] = useState({ open: false, memberId: null });

  // 模拟数据获取
  useEffect(() => {
    const fetchMembers = async () => {
      try {
        setLoading(true);
        // 模拟API调用延迟
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // 模拟成员数据
        const mockMembers = [
          { id: '1', username: 'alice', email: 'alice@example.com', role: 'admin', isActive: true, joinedAt: '2023-11-01' },
          { id: '2', username: 'bob', email: 'bob@example.com', role: 'member', isActive: true, joinedAt: '2023-11-02' },
          { id: '3', username: 'charlie', email: 'charlie@example.com', role: 'member', isActive: false, joinedAt: '2023-11-03' },
          { id: '4', username: 'david', email: 'david@example.com', role: 'member', isActive: true, joinedAt: '2023-11-04' },
          { id: '5', username: 'eve', email: 'eve@example.com', role: 'manager', isActive: true, joinedAt: '2023-11-05' },
        ];
        
        setMembers(mockMembers);
      } catch (error) {
        setSnackbar({ 
          open: true, 
          message: '获取成员列表失败', 
          severity: 'error' 
        });
      } finally {
        setLoading(false);
      }
    };

    fetchMembers();
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
  const handleEditMember = (member) => {
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
  const handleDeleteConfirm = (memberId) => {
    setConfirmDialog({ open: true, memberId });
  };

  // 处理删除成员
  const handleDeleteMember = async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      setMembers(prevMembers => 
        prevMembers.filter(member => member.id !== confirmDialog.memberId)
      );
      
      setSnackbar({ 
        open: true, 
        message: '成员删除成功', 
        severity: 'success' 
      });
    } catch (error) {
      setSnackbar({ 
        open: true, 
        message: '成员删除失败', 
        severity: 'error' 
      });
    } finally {
      setConfirmDialog({ open: false, memberId: null });
    }
  };

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      if (isEditing) {
        // 更新现有成员
        setMembers(prevMembers => 
          prevMembers.map(member => 
            member.id === currentMember.id 
              ? { ...member, ...formData }
              : member
          )
        );
        setSnackbar({ 
          open: true, 
          message: '成员信息更新成功', 
          severity: 'success' 
        });
      } else {
        // 添加新成员
        const newMember = {
          id: Date.now().toString(),
          ...formData,
          joinedAt: new Date().toISOString().split('T')[0]
        };
        setMembers(prevMembers => [...prevMembers, newMember]);
        setSnackbar({ 
          open: true, 
          message: '成员添加成功', 
          severity: 'success' 
        });
      }
      
      setOpenDialog(false);
    } catch (error) {
      setSnackbar({ 
        open: true, 
        message: isEditing ? '成员信息更新失败' : '成员添加失败', 
        severity: 'error' 
      });
    }
  };

  // 处理状态切换
  const handleStatusToggle = async (memberId, newStatus) => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
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
    } catch (error) {
      setSnackbar({ 
        open: true, 
        message: '状态更新失败', 
        severity: 'error' 
      });
    }
  };

  // 刷新成员列表
  const handleRefresh = () => {
    setLoading(true);
    // 模拟刷新操作
    setTimeout(() => {
      // 这里可以重新调用获取成员的函数
      setLoading(false);
      setSnackbar({ 
        open: true, 
        message: '成员列表已刷新', 
        severity: 'info' 
      });
    }, 500);
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        网络成员管理
      </Typography>
      
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
            onChange={(e) => setSearchTerm(e.target.value)}
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
                            onChange={(e) => handleStatusToggle(member.id, e.target.checked)}
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
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
            />
            
            <TextField
              label="邮箱"
              type="email"
              fullWidth
              required
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            />
            
            <TextField
              label="角色"
              select
              fullWidth
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value })}
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
                  onChange={(e) => setFormData({ ...formData, isActive: e.target.checked })}
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