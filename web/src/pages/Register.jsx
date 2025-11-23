import React, { useState, useCallback } from 'react'
import { Box, TextField, Button, Typography, Container, Paper, CircularProgress }
from '@mui/material'
import { Link, useNavigate } from 'react-router-dom'
import { authAPI } from '../services/api.js'
import ErrorAlert from '../components/ErrorAlert.jsx'

function Register() {
  const navigate = useNavigate()
  
  // 合并状态为单个对象，减少渲染次数
  const [state, setState] = useState({
    formData: {
      username: '',
      email: '',
      password: '',
      confirmPassword: ''
    },
    error: '',
    success: '',
    loading: false
  })

  // 使用useCallback缓存handleChange函数
  const handleChange = useCallback((e) => {
    setState(prev => ({
      ...prev,
      formData: {
        ...prev.formData,
        [e.target.name]: e.target.value
      }
    }))
  }, [])

  // 使用useCallback缓存handleSubmit函数
  const handleSubmit = useCallback(async (e) => {
    e.preventDefault()
    
    // 防止重复提交
    if (state.loading) return
    
    setState(prev => ({ ...prev, error: '', success: '' }))

    // 验证密码
    if (state.formData.password !== state.formData.confirmPassword) {
      setState(prev => ({ ...prev, error: '两次输入的密码不一致' }))
      return
    }

    setState(prev => ({ ...prev, loading: true }))
    
    try {
      // 提取表单数据
      const { username, email, password } = state.formData
      
      await authAPI.register({
        username,
        email,
        password
      })
      
      setState(prev => ({ ...prev, success: '注册成功，请登录' }))
      
      // 3秒后跳转到登录页
      setTimeout(() => {
        navigate('/login')
      }, 3000)
      
    } catch (err) {
      const errorMessage = err.response?.data?.message || '注册失败，请稍后重试'
      setState(prev => ({ ...prev, error: errorMessage }))
      console.error('Registration error:', err)
    } finally {
      setState(prev => ({ ...prev, loading: false }))
    }
  }, [state.formData, state.loading, navigate])

  return (
    <Container component="main" maxWidth="xs" sx={{ mt: 8 }}>
      <Paper elevation={3} sx={{ p: 4, borderRadius: 2 }}>
        <Typography component="h1" variant="h5" align="center" sx={{ mb: 4 }}>
          用户注册
        </Typography>
        
        {state.error && (
          <ErrorAlert 
            message={state.error} 
            onClose={() => setState(prev => ({ ...prev, error: '' }))} 
          />
        )}
        
        {state.success && (
          <ErrorAlert 
            message={state.success} 
            severity="success"
            onClose={() => setState(prev => ({ ...prev, success: '' }))} 
          />
        )}
        
        <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
          <TextField
            margin="normal"
            required
            fullWidth
            id="username"
            label="用户名"
            name="username"
            autoComplete="username"
            autoFocus
            value={state.formData.username}
            onChange={handleChange}
            disabled={state.loading}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            id="email"
            label="邮箱"
            name="email"
            type="email"
            autoComplete="email"
            value={state.formData.email}
            onChange={handleChange}
            disabled={state.loading}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            name="password"
            label="密码"
            type="password"
            id="password"
            autoComplete="new-password"
            value={state.formData.password}
            onChange={handleChange}
            disabled={state.loading}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            name="confirmPassword"
            label="确认密码"
            type="password"
            id="confirmPassword"
            value={state.formData.confirmPassword}
            onChange={handleChange}
            disabled={state.loading}
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={state.loading}
          >
            {state.loading ? 
              <CircularProgress size={24} color="inherit" /> : 
              '注册'
            }
          </Button>
          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="body2">
              已有账号？ 
              <Link to="/login" style={{ color: '#64b5f6', textDecoration: 'none' }}>
                立即登录
              </Link>
            </Typography>
          </Box>
        </Box>
      </Paper>
    </Container>
  )
}

export default Register