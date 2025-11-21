import React, { useState, useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout.jsx'
import Login from './pages/Login.jsx'
import Register from './pages/Register.jsx'
import Dashboard from './pages/Dashboard.jsx'
import Networks from './pages/Networks.jsx'
import NetworkDetail from './pages/NetworkDetail.jsx'
import Members from './pages/Members.jsx'
import Profile from './pages/Profile.jsx'
import Settings from './pages/Settings.jsx'
import api from './services/api.js'
import { AuthProvider, useAuth } from './services/auth.jsx'
import './App.css'

// 创建使用AuthContext的内部应用组件
function AppContent() {
  const [loading, setLoading] = useState(true)
  const { user, token, login, logout } = useAuth() || {};

  // 检查用户登录状态
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const storedToken = localStorage.getItem('token') || sessionStorage.getItem('token');
        const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
        
        if (storedToken && storedUser && login) {
          api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`;
          const userData = JSON.parse(storedUser);
          login(userData, storedToken);
        }
      } catch (error) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        sessionStorage.removeItem('token');
        sessionStorage.removeItem('user');
        delete api.defaults.headers.common['Authorization'];
      } finally {
        setLoading(false);
      }
    }

    checkAuth();
  }, [login]);

  // 处理注册成功
  const handleRegisterSuccess = () => {
    // 可以在这里添加注册成功后的逻辑
    console.log('注册成功');
  }

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        fontSize: '18px'
      }}>
        加载中...
      </div>
    )
  }

  return (
    <div className="app">
      <Routes>
        <Route path="/login" element={<Login />}></Route>
        <Route path="/register" element={<Register onRegisterSuccess={handleRegisterSuccess} />}></Route>
        
        {user ? (
          <>
            <Route path="/" element={<Layout user={user} onLogout={logout} />}>
              <Route index element={<Dashboard />}></Route>
              <Route path="networks" element={<Networks />}></Route>
              <Route path="networks/:id" element={<NetworkDetail />}></Route>
              <Route path="networks/:id/members" element={<Members />}></Route>
              <Route path="profile" element={<Profile user={user} />}></Route>
              <Route path="settings" element={<Settings />}></Route>
            </Route>
          </>
        ) : (
          <>
            <Route path="/*" element={<Navigate to="/login" replace />}></Route>
          </>
        )}
      </Routes>
    </div>
  )
}

// 主App组件，使用AuthProvider包裹
function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  )
}

export default App