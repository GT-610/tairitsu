import React, { useState, useEffect } from 'react';
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
import api from './services/api';
import { useAuth } from './services/auth';
import './App.css';

function AppContent() {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();
  const [isFirstRun, setIsFirstRun] = useState<boolean | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const location = useLocation();

  // Listen for API errors and handle logout on 401 unauthorized responses
  useEffect(() => {
    const handleApiError = (error: any) => {
      // Check if the error is a 401 unauthorized error
      if (error.response && error.response.status === 401) {
        // Clear authentication information
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        sessionStorage.removeItem('token');
        sessionStorage.removeItem('user');
        delete api.defaults.headers.common['Authorization'];
        
        // Redirect to login page (using React Router instead of window.location)
        if (location.pathname !== '/login') {
          navigate('/login');
        }
      }
    };

    // Add response interceptor
    const interceptor = api.interceptors.response.use(
      response => response,
      error => {
        handleApiError(error);
        return Promise.reject(error);
      }
    );

    // Cleanup function
    return () => {
      api.interceptors.response.eject(interceptor);
    };
  }, [navigate, location.pathname]);

  // Check if system is initialized (executed only once on app startup)
  useEffect(() => {
    const checkFirstRun = async () => {
      try {
        const response = await api.get('/system/status', {
          headers: {
            'Cache-Control': 'no-cache'
          }
        });
        
        // Ensure response.data exists and contains initialized field
        const isBackendInitialized = response.data && response.data.initialized;
        
        // Fully rely on backend API status, not using local storage
        setIsFirstRun(!isBackendInitialized);
        
        // Log API response for debugging
        console.log('系统状态检查:', { 
          initialized: isBackendInitialized, 
          isFirstRun: !isBackendInitialized,
          // If backend returns more information, also log it for debugging
          additionalInfo: {
            hasDatabase: response.data?.hasDatabase,
            hasAdmin: response.data?.hasAdmin,
            ztStatus: response.data?.ztStatus
          }
        });
      } catch (error) {
        console.error('Failed to get backend initialization status:', error);
        // When backend is unavailable, default to first run status, requiring user to connect to backend
        // No longer use local storage as fallback mechanism, fully rely on backend API
        setIsFirstRun(true);
      } finally {
        setLoading(false);
      }
    };

    checkFirstRun();
  }, []);

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
        {/* Show setup wizard on first run */}
        {isFirstRun ? (
          <>
            <Route path="/setup" element={<SetupWizard />}></Route>
            {/* Use replace instead of push to prevent users from returning to setup wizard via browser back button */}
            <Route path="*" element={<Navigate to="/setup" replace />}></Route>
          </>
        ) : (
          <>
            <Route path="/login" element={<Login />}></Route>
            <Route path="/register" element={<Register />}></Route>
            
            {isAuthenticated() ? (
              <>
                <Route path="/" element={<Layout user={user} />}>
                  <Route path="dashboard" element={<Dashboard />}></Route>
                  <Route path="networks" element={<Networks />}></Route>
                  <Route path="networks/:id" element={<NetworkDetail />}></Route>
                  <Route path="networks/:id/members" element={<Members />}></Route>
                  <Route path="profile" element={<Profile user={user} />}></Route>
                  <Route path="settings" element={<Settings />}></Route>
                </Route>
              </>
            ) : (
              <>
                {/* Add root path direct redirect to login page */}
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