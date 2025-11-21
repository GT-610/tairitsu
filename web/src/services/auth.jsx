import { createContext, useContext, useState } from 'react';

// 创建Auth上下文
const AuthContext = createContext(null);

// Auth Provider组件
export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  // 登录函数
  const login = (userData, authToken) => {
    setUser(userData);
    setToken(authToken);
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
  const isAuthenticated = () => {
    return !!user && !!token;
  };

  // 上下文值
  const value = {
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
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    console.warn('useAuth must be used within an AuthProvider');
  }
  return context;
};

// 默认导出
export default useAuth;