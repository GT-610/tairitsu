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
        // 设置请求头中的Cache-Control为no-cache，确保获取最新状态
        const response = await api.get('/system/status', {
          headers: {
            'Cache-Control': 'no-cache'
          }
        });
        // 根据系统的initialized状态来判断是否是首次运行
        // 如果系统未初始化，则认为是首次运行
        setIsFirstRun(!response.data.initialized);
        console.log('系统状态检查:', { initialized: response.data.initialized, isFirstRun: !response.data.initialized });
      } catch (error) {
        console.error('获取系统状态失败:', error);
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

  // 监听系统初始化状态变化，当设置向导完成后，重新检查系统状态
  useEffect(() => {
    // 定义一个函数来刷新系统状态
    const refreshSystemStatus = async () => {
      try {
        // 首先检查localStorage中的初始化状态
        const localStorageInitialized = localStorage.getItem('tairitsu_initialized');
        if (localStorageInitialized === 'true') {
          setIsFirstRun(false);
          console.log('[App] 从localStorage检测到系统已初始化');
          // 如果localStorage中标记为已初始化，同步到后端检查
          try {
            const response = await api.get('/system/status', {
              headers: {
                'Cache-Control': 'no-cache'
              }
            });
            // 以后端状态为准，更新前端状态
            setIsFirstRun(!response.data.initialized);
            console.log('[App] 后端验证系统初始化状态:', response.data.initialized);
          } catch (backendError) {
            console.warn('[App] 后端验证失败，但仍使用localStorage状态');
          }
          return;
        }
        
        // 如果localStorage中没有，从后端获取
        const response = await api.get('/system/status', {
          headers: {
            'Cache-Control': 'no-cache'
          }
        });
        // 根据系统的initialized状态来更新
        setIsFirstRun(!response.data.initialized);
        console.log('[App] 从后端刷新系统状态，初始化状态:', response.data.initialized);
        
        // 如果后端返回已初始化，更新localStorage
        if (response.data.initialized) {
          localStorage.setItem('tairitsu_initialized', 'true');
        }
      } catch (error) {
        console.error('[App] 刷新系统状态失败:', error);
      }
    };

    // 监听存储事件，当有其他标签页更新时，也刷新系统状态
    const handleStorageChange = (event) => {
      if (event.key === 'tairitsu_initialized') {
        console.log('[App] 检测到初始化状态变化，刷新系统状态');
        refreshSystemStatus();
      }
    };
    
    // 组件挂载时立即检查localStorage
    refreshSystemStatus();

    // 添加监听器
    window.addEventListener('storage', handleStorageChange);

    // 清理监听器
    return () => {
      window.removeEventListener('storage', handleStorageChange);
    };
  }, []);

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
            {/* 使用replace而不是push，防止用户通过浏览器返回按钮回到设置向导 */}
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
                {/* 添加根路径直接重定向到登录页面 */}
                <Route path="/" element={<Navigate to="/login" replace />}></Route>
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