import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import api, { authAPI, User, UserSession } from './api';
import { clearPersistedAuthState, restoreAuthState } from './authStorage';

// 定义Auth上下文类型
export interface AuthContextType {
  user: User | null;
  token: string | null;
  session: UserSession | null;
  isLoading: boolean;
  login: (userData: User, authToken: string, userSession: UserSession) => { success: boolean; user: User; token: string; session: UserSession };
  refreshUser: (userData: User) => void;
  logout: () => Promise<void>;
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
  const [isLoading] = useState<boolean>(false);

  // 在应用启动时从存储中恢复认证状态
  useEffect(() => {
    const restored = restoreAuthState();
    if (restored) {
      setUser(restored.user);
      setToken(restored.token);
      setSession(restored.session);
    }
  }, []);

  // 登录函数
  const login = (userData: User, authToken: string, userSession: UserSession) => {
    setUser(userData);
    setToken(authToken);
    setSession(userSession);
    // 更新API实例的默认请求头
    api.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
    return { success: true, user: userData, token: authToken, session: userSession };
  };

  const refreshUser = (userData: User) => {
    setUser(userData);

    if (localStorage.getItem('token')) {
      localStorage.setItem('user', JSON.stringify(userData));
    } else if (sessionStorage.getItem('token')) {
      sessionStorage.setItem('user', JSON.stringify(userData));
    }
  };

  // 登出函数
  const clearAuthState = () => {
    setUser(null);
    setToken(null);
    setSession(null);
    clearPersistedAuthState();
  };

  const logout = async () => {
    try {
      if (token) {
        await authAPI.logout();
      }
    } catch {
      // Ignore logout API failures and still clear local auth state.
    } finally {
      clearAuthState();
    }
  };

  // 检查是否已认证
  const isAuthenticated = (): boolean => {
    return !!user && !!token;
  };

  // 上下文值
  const value: AuthContextType = {
    user,
    token,
    session,
    isLoading,
    login,
    refreshUser,
    logout,
    isAuthenticated
  };

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
