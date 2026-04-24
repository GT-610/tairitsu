import { useState, useEffect, lazy, Suspense, useCallback } from 'react';
import { Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom';
import { CircularProgress, Box } from '@mui/material';
import Login from './pages/Login';
import Register from './pages/Register';
import SetupWizard, { setupCompletedEvent } from './pages/SetupWizard';
import Networks from './pages/Networks';
import NotFound from './pages/NotFound';
import Layout from './components/Layout';
import api, { type SetupStatus } from './services/api';
import { hasStatus } from './services/errors';
import { useAuth } from './services/auth';

const lazyPages = {
  UserManagement: lazy(() => import('./pages/UserManagement')),
  Planet: lazy(() => import('./pages/Planet')),
  ImportNetwork: lazy(() => import('./pages/ImportNetwork')),
  Dashboard: lazy(() => import('./pages/Dashboard')),
  Settings: lazy(() => import('./pages/Settings')),
  Profile: lazy(() => import('./pages/Profile')),
  NetworkDetail: lazy(() => import('./pages/NetworkDetail')),
};

function Loading() {
  return (
    <Box sx={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100vh'
    }}>
      <CircularProgress />
    </Box>
  );
}
// import './App.css';

function AppContent() {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();
  const [isFirstRun, setIsFirstRun] = useState<boolean | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const location = useLocation();

  const checkFirstRun = useCallback(async () => {
    try {
      const response = await api.get<SetupStatus>('/system/status', {
        headers: {
          'Cache-Control': 'no-cache'
        }
      });

      const isBackendInitialized = response.data.initialized;
      setIsFirstRun(!isBackendInitialized);
    } catch {
      setIsFirstRun(true);
    } finally {
      setLoading(false);
    }
  }, []);

  // Listen for API errors and handle logout on 401 unauthorized responses
  useEffect(() => {
    const handleApiError = (error: unknown) => {
      // Check if the error is a 401 unauthorized error
      if (hasStatus(error, 401)) {
        // Clear authentication information
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('session');
        sessionStorage.removeItem('token');
        sessionStorage.removeItem('user');
        sessionStorage.removeItem('session');
        delete api.defaults.headers.common['Authorization'];
        
        // Redirect to login page (using React Router instead of window.location)
        if (location.pathname !== '/login') {
          void navigate('/login');
        }
      }
    };

    // 添加响应拦截器
    const interceptor = api.interceptors.response.use(
      response => response,
      error => {
        handleApiError(error);
        return Promise.reject(error instanceof Error ? error : new Error(String(error)));
      }
    );

    // 清理函数
    return () => {
      api.interceptors.response.eject(interceptor);
    };
  }, [navigate, location.pathname]);

  // 检查系统是否已初始化（仅在应用启动时执行一次）
  useEffect(() => {
    void checkFirstRun();
  }, [checkFirstRun]);

  useEffect(() => {
    const handleSetupComplete = () => {
      setLoading(true);
      void checkFirstRun();
    };

    window.addEventListener(setupCompletedEvent, handleSetupComplete);
    return () => {
      window.removeEventListener(setupCompletedEvent, handleSetupComplete);
    };
  }, [checkFirstRun]);

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
    );
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
            <Route path="/register" element={<Register />}></Route>
            {isAuthenticated() ? (
              <>
                <Route path="/" element={<Layout user={user} />}>
                  {/* 公共路由 */}
                  <Route path="networks" element={<Networks />}></Route>
                  <Route path="networks/:id" element={<Suspense fallback={<Loading />}><lazyPages.NetworkDetail /></Suspense>}></Route>
                  <Route path="profile" element={<Suspense fallback={<Loading />}><lazyPages.Profile user={user} /></Suspense>}></Route>
                  <Route path="settings" element={<Suspense fallback={<Loading />}><lazyPages.Settings /></Suspense>}></Route>
                  
                  {/* 管理员路由 */}
                  {user && user.role === 'admin' && (
                    <>
                      <Route path="dashboard" element={<Suspense fallback={<Loading />}><lazyPages.Dashboard /></Suspense>}></Route>
                      <Route path="user-management" element={<Suspense fallback={<Loading />}><lazyPages.UserManagement /></Suspense>}></Route>
                      <Route path="import-network" element={<Suspense fallback={<Loading />}><lazyPages.ImportNetwork /></Suspense>}></Route>
                      <Route path="planet" element={<Suspense fallback={<Loading />}><lazyPages.Planet /></Suspense>}></Route>
                    </>
                  )}
                  
                  {/* 默认路由重定向 */}
                  <Route path="" element={<Navigate to="/networks" replace />}></Route>
                  <Route path="*" element={<NotFound />}></Route>
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
  );
}

export default AppContent;
