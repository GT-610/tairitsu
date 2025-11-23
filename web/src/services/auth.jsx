/**
 * 认证服务模块
 * 提供认证上下文(AuthContext)和相关认证功能
 */
import { createContext, useContext, useState } from 'react';

/**
 * 认证上下文对象，用于在组件树中共享认证状态
 * @type {React.Context<{user: Object|null, token: string|null, isLoading: boolean, login: Function, logout: Function, isAuthenticated: Function}|null>}
 */
const AuthContext = createContext(null);

/**
 * 认证提供者组件，为子组件提供认证状态和方法
 * 
 * @component
 * @param {Object} props - 组件属性
 * @param {React.ReactNode} props.children - 子组件
 * @returns {React.ReactNode}
 */
export const AuthProvider = ({ children }) => {
  // 认证状态管理
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  /**
   * 用户登录方法
   * 
   * @param {Object} userData - 用户数据对象
   * @param {string} authToken - 认证令牌
   * @returns {Object} 登录结果对象
   * @returns {boolean} return.success - 登录是否成功
   * @returns {Object} return.user - 用户数据
   * @returns {string} return.token - 认证令牌
   */
  const login = (userData, authToken) => {
    setUser(userData);
    setToken(authToken);
    return { success: true, user: userData, token: authToken };
  };

  /**
   * 用户登出方法
   * 清除用户状态、令牌并移除存储中的认证信息
   * 
   * @returns {void}
   */
  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    sessionStorage.removeItem('user');
    sessionStorage.removeItem('token');
  };

  /**
   * 检查用户是否已认证
   * 
   * @returns {boolean} 是否已认证
   */
  const isAuthenticated = () => {
    return !!user && !!token;
  };

  /**
   * 认证上下文值
   * 
   * @type {Object}
   */
  const authContextValue = {
    user,
    token,
    isLoading,
    setIsLoading,
    login,
    logout,
    isAuthenticated
  };

  return (
    <AuthContext.Provider value={authContextValue}>
      {children}
    </AuthContext.Provider>
  );
};

/**
 * 自定义钩子，用于在组件中访问认证上下文
 * 
 * @returns {Object} 认证上下文对象
 * @throws {Error} 当在AuthProvider外部使用时发出警告
 */
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    console.warn('useAuth钩子必须在AuthProvider组件内部使用');
  }
  return context;
};

// 默认导出
export default useAuth;