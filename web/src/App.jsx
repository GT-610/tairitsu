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
import SetupWizard from './pages/SetupWizard.jsx'
import api from './services/api.js'
import { AuthProvider, useAuth } from './services/auth.jsx'
import './App.css'

// 创建使用AuthContext的内部应用组件
function AppContent() {
  const [loading, setLoading] = useState(true)
  const [isFirstRun, setIsFirstRun] = useState(false)
  const { user, token, login, logout } = useAuth() || {};

  // 检查是否是首次运行
  useEffect(() => {
    const checkFirstRun = async () => {
      try {
        // 尝试获取系统设置状态来判断是否已完成设置
        const response = await api.get('/system/status');
        // 如果没有管理员用户，则认为是首次运行
        setIsFirstRun(!response.data.hasAdmin);
      } catch (error) {
        // 如果无法获取状态，可能是首次运行
        setIsFirstRun(true);
      } finally {
        setLoading(false);
      }
    };

    checkFirstRun();
  }, []);

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
      }
    }

    // 只有不是首次运行时才检查认证状态
    if (!isFirstRun) {
      checkAuth();
    }
  }, [login, isFirstRun]);

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
        {/* 首次运行时显示设置向导 */}
        {isFirstRun ? (
          <>
            <Route path="/setup" element={<SetupWizard />}></Route>
            <Route path="*" element={<Navigate to="/setup" replace />}></Route>
          </>
        ) : (
          <>
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