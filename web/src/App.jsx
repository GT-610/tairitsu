import React, { useState, useEffect, lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import Layout from './components/Layout.jsx'
import { CSSTransition, TransitionGroup } from 'react-transition-group'
import './TransitionStyles.css'
// 使用懒加载优化资源加载性能
const Login = lazy(() => import('./pages/Login.jsx'))
const Register = lazy(() => import('./pages/Register.jsx'))
const Dashboard = lazy(() => import('./pages/Dashboard.jsx'))
const Networks = lazy(() => import('./pages/Networks.jsx'))
const NetworkDetail = lazy(() => import('./pages/NetworkDetail.jsx'))
const Members = lazy(() => import('./pages/Members.jsx'))
const Profile = lazy(() => import('./pages/Profile.jsx'))
const Settings = lazy(() => import('./pages/Settings.jsx'))
const SetupWizard = lazy(() => import('./pages/SetupWizard.jsx'))
import api from './services/api.js'
import { AuthProvider, useAuth } from './services/auth.jsx'
import './App.css'

// 路由内容组件，用于添加过渡效果
function RouteContent({ user, logout, isFirstRun, handleRegisterSuccess }) {
  // 在BrowserRouter内部使用useLocation
  const location = useLocation();
  const nodeRef = React.useRef(null);
  
  return (
    <TransitionGroup>
      <CSSTransition
        key={location.pathname}
        timeout={300}
        classNames="page-transition"
        nodeRef={nodeRef}
      >
        <div ref={nodeRef}>
          <Routes location={location}>
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
                    <Route path="/" element={<Navigate to="/login" replace />}></Route>
                    <Route path="/*" element={<Navigate to="/login" replace />}></Route>
                  </>
                )}
              </>
            )}
          </Routes>
        </div>
      </CSSTransition>
    </TransitionGroup>
  )
}

// 内部应用内容组件，在BrowserRouter内部使用
function AppContentInternal() {
  const [loading, setLoading] = useState(true)
  const [isFirstRun, setIsFirstRun] = useState(false)
  const { user, login, logout } = useAuth() || {};

  // 合并系统状态检查逻辑，优化性能
  useEffect(() => {
    const checkSystemStatus = async () => {
      try {
        // 先检查URL是否是设置向导页面
        const isSetupWizardPage = window.location.pathname === '/setup';
        
        // 检查localStorage中的初始化状态（优先级最高）
        const localStorageInitialized = localStorage.getItem('tairitsu_initialized') === 'true';
        const setupWizardStarted = localStorage.getItem('tairitsu_setup_started') === 'true';
        
        // 如果有初始化标记，直接设为非首次运行
        if (localStorageInitialized) {
          setIsFirstRun(false);
          return;
        }
        
        // 如果是设置向导页面或设置已开始，直接设为首次运行
        if (isSetupWizardPage || setupWizardStarted) {
          setIsFirstRun(true);
          return;
        }
        
        // 只有在必要时才发送请求检查系统状态
        const response = await api.get('/system/status', {
          headers: {
            'Cache-Control': 'no-cache'
          }
        });
        const initialized = response.data.initialized;
        setIsFirstRun(!initialized);
        
        // 如果后端返回已初始化，更新localStorage
        if (initialized) {
          localStorage.setItem('tairitsu_initialized', 'true');
        }
      } catch (error) {
        console.error('获取系统状态失败:', error);
        setIsFirstRun(true);
      } finally {
        setLoading(false);
      }
    };

    checkSystemStatus();
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

  // 监听系统初始化状态变化
  useEffect(() => {
    // 监听存储事件，当有其他标签页更新时，刷新系统状态
    const handleStorageChange = (event) => {
      if (event.key === 'tairitsu_initialized' && event.newValue === 'true') {
        console.log('[App] 检测到初始化状态变化，更新为已初始化');
        setIsFirstRun(false);
      }
    };

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
      <Suspense fallback={
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          height: '100vh',
          fontSize: '18px'
        }}>
          加载中...
        </div>
      }>
        <RouteContent 
          user={user} 
          logout={logout}
          isFirstRun={isFirstRun}
          handleRegisterSuccess={handleRegisterSuccess}
        />
      </Suspense>
    </div>
  )
}

// AppContent组件，直接返回内部内容组件，使用main.jsx中提供的BrowserRouter
function AppContent() {
  return <AppContentInternal />;
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