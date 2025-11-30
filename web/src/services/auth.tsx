import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import api, { User } from './api';

// 定义Auth上下文类型
export interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (userData: User, authToken: string) => { success: boolean; user: User; token: string };
  logout: () => void;
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
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // 在应用启动时从存储中恢复认证状态
  useEffect(() => {
    const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
    const storedToken = localStorage.getItem('token') || sessionStorage.getItem('token');
    
    if (storedUser && storedToken) {
      try {
        const userData = JSON.parse(storedUser) as User;
        setUser(userData);
        setToken(storedToken);
        // 更新API实例的默认请求头
        api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`;
      } catch (error) {
        console.error('解析存储的用户数据失败:', error);
        // 如果解析失败，清除存储的数据
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        sessionStorage.removeItem('user');
        sessionStorage.removeItem('token');
      }
    }
  }, []);

  // 登录函数
  const login = (userData: User, authToken: string) => {
    setUser(userData);
    setToken(authToken);
    // 更新API实例的默认请求头
    api.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
    return { success: true, user: userData, token: authToken };
  };

  // 登出函数
  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    sessionStorage.removeItem('user');
    sessionStorage.removeItem('token');
  };

  // 检查是否已认证
  const isAuthenticated = (): boolean => {
    return !!user && !!token;
  };

  // 上下文值
  const value: AuthContextType = {
    user,
    token,
    isLoading,
    login,
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