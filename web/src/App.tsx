import { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { CircularProgress, Box } from '@mui/material';
import Login from './pages/Login';
import Register from './pages/Register';
import SetupWizard from './pages/SetupWizard';
import Networks from './pages/Networks';
import NotFound from './pages/NotFound';
import Layout from './components/Layout';
import { useAuth } from './services/auth';
import { useSetupGate, useUnauthorizedRedirect } from './hooks/useAppRuntime';

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
  const { user, isAuthenticated } = useAuth();
  const { isFirstRun, loading } = useSetupGate();
  useUnauthorizedRedirect();

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
