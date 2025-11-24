import { createContext, useContext, useState, useEffect } from 'react';

// 创建Auth上下文
const AuthContext = createContext(null);

// Auth Provider组件
export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [tempToken, setTempToken] = useState(null);
  const [isSetupWizard, setIsSetupWizard] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // 初始化时从存储中恢复认证状态
  useEffect(() => {
    const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
    const storedToken = localStorage.getItem('token') || sessionStorage.getItem('token');
    const storedTempToken = localStorage.getItem('tempToken') || sessionStorage.getItem('tempToken');
    const storedIsSetupWizard = localStorage.getItem('isSetupWizard') === 'true';

    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
    if (storedToken) {
      setToken(storedToken);
    }
    if (storedTempToken) {
      setTempToken(storedTempToken);
      setIsSetupWizard(storedIsSetupWizard);
    }
  }, []);

  // 保存认证状态到存储
  const saveAuthState = (userData, authToken, rememberMe = false) => {
    const storage = rememberMe ? localStorage : sessionStorage;
    if (userData) {
      storage.setItem('user', JSON.stringify(userData));
    }
    if (authToken) {
      storage.setItem('token', authToken);
    }
  };

  // 保存临时令牌状态
  const saveTempTokenState = (token, isSetup) => {
    // 临时令牌只保存在会话中，不长期存储
    sessionStorage.setItem('tempToken', token);
    sessionStorage.setItem('isSetupWizard', isSetup.toString());
  };

  // 登录函数
  const login = (userData, authToken, rememberMe = false) => {
    setUser(userData);
    setToken(authToken);
    setTempToken(null);
    setIsSetupWizard(false);
    saveAuthState(userData, authToken, rememberMe);
    // 清除临时令牌状态
    sessionStorage.removeItem('tempToken');
    sessionStorage.removeItem('isSetupWizard');
    return { success: true, user: userData, token: authToken };
  };

  // 设置向导登录（使用临时令牌）
  const loginWithSetupToken = (tempAuthToken) => {
    setTempToken(tempAuthToken);
    setIsSetupWizard(true);
    setUser(null);
    setToken(null);
    saveTempTokenState(tempAuthToken, true);
    // 清除普通认证状态
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    sessionStorage.removeItem('user');
    sessionStorage.removeItem('token');
    return { success: true, tempToken: tempAuthToken };
  };

  // 登出函数
  const logout = () => {
    setUser(null);
    setToken(null);
    setTempToken(null);
    setIsSetupWizard(false);
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    sessionStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('tempToken');
    sessionStorage.removeItem('isSetupWizard');
  };

  // 检查是否已认证
  const isAuthenticated = () => {
    return (!!user && !!token) || (!!tempToken && isSetupWizard);
  };

  // 检查是否是设置向导认证
  const isSetupWizardAuthenticated = () => {
    return !!tempToken && isSetupWizard;
  };

  // 获取当前使用的令牌
  const getCurrentToken = () => {
    return tempToken || token;
  };

  // 上下文值
  const value = {
    user,
    token,
    tempToken,
    isSetupWizard,
    isLoading,
    login,
    loginWithSetupToken,
    logout,
    isAuthenticated,
    isSetupWizardAuthenticated,
    getCurrentToken
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