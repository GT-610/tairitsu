import React, { useState, useEffect, useCallback } from 'react';
import { Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom';
import Login from './pages/Login';
import Register from './pages/Register';
import SetupWizard from './pages/SetupWizard';
import Dashboard from './pages/Dashboard';
import Networks from './pages/Networks';
import NetworkDetail from './pages/NetworkDetail';
import Members from './pages/Members';
import Profile from './pages/Profile';
import Settings from './pages/Settings';
import Layout from './components/Layout';
import api from './services/api.js';
import './App.css';

function AppContent() {
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [isFirstRun, setIsFirstRun] = useState(null);
  const [loading, setLoading] = useState(true);
  const location = useLocation();

  // 监听API错误并在401时处理登出
  useEffect(() => {
    const handleApiError = (error) => {
      // 检查是否是401未授权错误
      if (error.response && error.response.status === 401) {
        // 清除认证信息
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        sessionStorage.removeItem('token');
        sessionStorage.removeItem('user');
        delete api.defaults.headers.common['Authorization'];
        
        // 更新用户状态
        setUser(null);
        
        // 重定向到登录页面（使用React Router而非window.location）
        if (location.pathname !== '/login') {
          navigate('/login');
        }
      }
    };

    // 添加响应拦截器
    const interceptor = api.interceptors.response.use(
      response => response,
      error => {
        handleApiError(error);
        return Promise.reject(error);
      }
    );

    // 清理函数
    return () => {
      api.interceptors.response.eject(interceptor);
    };
  }, [navigate, location.pathname]);

  // 检查系统是否已初始化（仅在应用启动时执行一次）
  useEffect(() => {
    const checkFirstRun = async () => {
      try {
        const response = await api.get('/system/status', {
          headers: {
            'Cache-Control': 'no-cache'
          }
        });
        
        // 确保response.data存在且包含initialized字段
        const isBackendInitialized = response.data && response.data.initialized;
        
        // 完全依赖后端API状态，不使用本地存储
        setIsFirstRun(!isBackendInitialized);
        
        // 记录API响应，帮助调试
        console.log('系统状态检查:', { 
          initialized: isBackendInitialized, 
          isFirstRun: !isBackendInitialized,
          // 如果后端返回了更多信息，也记录下来以便调试
          additionalInfo: {
            hasDatabase: response.data?.hasDatabase,
            hasAdmin: response.data?.hasAdmin,
            ztStatus: response.data?.ztStatus
          }
        });
      } catch (error) {
        console.error('获取后端初始化状态失败:', error);
        // 当后端不可用时，默认显示为首次运行，要求用户连接到后端
        // 不再使用本地存储作为回退机制，完全依赖后端API
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
        
        if (storedToken && storedUser) {
          api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`;
          const userData = JSON.parse(storedUser);
          setUser(userData);
        }
      } catch (error) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        sessionStorage.removeItem('token');
        sessionStorage.removeItem('user');
        delete api.defaults.headers.common['Authorization'];
      }
    }

    // 只有不是首次运行且系统状态检查完成时才检查认证状态
    if (isFirstRun !== null && !isFirstRun) {
      checkAuth();
    }
  }, [isFirstRun]);

  // 移除了会导致重复请求的storage事件监听器和refreshSystemStatus函数

  // 处理用户登出
  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
    delete api.defaults.headers.common['Authorization'];
    setUser(null);
  };

  // 处理注册成功
  const handleRegisterSuccess = () => {
    // 可以在这里添加注册成功后的逻辑
    console.log('注册成功');
  }

  if (loading || isFirstRun === null) {
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
                <Route path="/" element={<Layout user={user} onLogout={handleLogout} />}>
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

export default AppContent;