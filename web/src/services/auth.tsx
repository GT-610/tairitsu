import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import api, { User } from './api';

// Define Auth context type
export interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (userData: User, authToken: string) => { success: boolean; user: User; token: string };
  logout: () => void;
  isAuthenticated: () => boolean;
}

// Create Auth context with AuthContextType
const AuthContext = createContext<AuthContextType | null>(null);

// Auth Provider component props type
interface AuthProviderProps {
  children: ReactNode;
}

// Auth Provider component
export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Restore authentication state from storage on app startup
  useEffect(() => {
    const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
    const storedToken = localStorage.getItem('token') || sessionStorage.getItem('token');
    
    if (storedUser && storedToken) {
      try {
        const userData = JSON.parse(storedUser) as User;
        setUser(userData);
        setToken(storedToken);
        // Update API instance default request headers
        api.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`;
      } catch (error) {
        console.error('Failed to parse stored user data:', error);
        // If parsing fails, clear stored data
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        sessionStorage.removeItem('user');
        sessionStorage.removeItem('token');
      }
    }
  }, []);

  // Login function
  const login = (userData: User, authToken: string) => {
    setUser(userData);
    setToken(authToken);
    // Update API instance default request headers
    api.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
    return { success: true, user: userData, token: authToken };
  };

  // Logout function
  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    sessionStorage.removeItem('user');
    sessionStorage.removeItem('token');
  };

  // Check if authenticated
  const isAuthenticated = (): boolean => {
    return !!user && !!token;
  };

  // Context value
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

// Custom hook for accessing Auth context in components
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// Default export
export default useAuth;