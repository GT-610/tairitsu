import { createContext, useContext, useState, useEffect, useCallback, useMemo, ReactNode } from 'react';
import api, { authAPI, User, UserSession } from './api';
import { clearPersistedAuthState, restoreAuthState } from './authStorage';

// 定义Auth上下文类型
interface AuthContextType {
  user: User | null;
  token: string | null;
  session: UserSession | null;
  isHydrated: boolean;
  login: (userData: User, authToken: string, userSession: UserSession) => { success: boolean; user: User; token: string; session: UserSession };
  refreshUser: (userData: User) => void;
  logout: () => Promise<void>;
  clearAuth: () => void;
  isAuthenticated: () => boolean;
}

// 创建Auth上下文，使用AuthContextType类型
const AuthContext = createContext<AuthContextType | null>(null);

// Auth Provider组件的props类型
interface AuthProviderProps {
  children: ReactNode;
}

// Auth Provider组件
export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [session, setSession] = useState<UserSession | null>(null);
  const [isHydrated, setIsHydrated] = useState(false);

  // 在应用启动时从存储中恢复认证状态
  useEffect(() => {
    const restored = restoreAuthState();
    if (restored) {
      setUser(restored.user);
      setToken(restored.token);
      setSession(restored.session);
    }
    setIsHydrated(true);
  }, []);

  // 登录函数
  const login = useCallback((userData: User, authToken: string, userSession: UserSession) => {
    setUser(userData);
    setToken(authToken);
    setSession(userSession);
    api.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
    return { success: true, user: userData, token: authToken, session: userSession };
  }, []);

  const refreshUser = useCallback((userData: User) => {
    setUser(userData);

    if (localStorage.getItem('token')) {
      localStorage.setItem('user', JSON.stringify(userData));
    } else if (sessionStorage.getItem('token')) {
      sessionStorage.setItem('user', JSON.stringify(userData));
    }
  }, []);

  const clearAuth = useCallback(() => {
    setUser(null);
    setToken(null);
    setSession(null);
    clearPersistedAuthState();
  }, []);

  // 登出函数
  const logout = useCallback(async () => {
    try {
      if (token) {
        await authAPI.logout();
      }
    } catch {
      // Ignore logout API failures and still clear local auth state.
    } finally {
      clearAuth();
    }
  }, [clearAuth, token]);

  // 检查是否已认证
  const isAuthenticated = useCallback((): boolean => {
    return !!user && !!token;
  }, [user, token]);

  // 上下文值
  const value = useMemo<AuthContextType>(() => ({
    user,
    token,
    session,
    isHydrated,
    login,
    refreshUser,
    logout,
    clearAuth,
    isAuthenticated
  }), [user, token, session, isHydrated, login, refreshUser, logout, clearAuth, isAuthenticated]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// 自定义钩子，用于在组件中访问Auth上下文
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// 默认导出
export default useAuth;
