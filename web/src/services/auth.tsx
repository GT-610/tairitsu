import { createContext, useContext, useState, useEffect, useCallback, useMemo, ReactNode } from 'react';
import { authAPI, User } from './api';

// The session credential lives in an HttpOnly cookie that JavaScript cannot
// read, so the client cannot inspect it to decide whether it is signed in.
// Authentication state is therefore derived from the server on startup and
// kept in memory only.
interface AuthContextType {
  user: User | null;
  isHydrated: boolean;
  login: (userData: User) => void;
  refreshUser: (userData: User) => void;
  logout: () => Promise<void>;
  clearAuth: () => void;
  isAuthenticated: () => boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<User | null>(null);
  const [isHydrated, setIsHydrated] = useState(false);

  // Restore the session by asking the server who we are. A missing or expired
  // cookie answers 401, which simply means "signed out".
  useEffect(() => {
    let cancelled = false;

    const restoreSession = async () => {
      try {
        const response = await authAPI.getProfile();
        if (!cancelled) {
          setUser(response.data);
        }
      } catch {
        if (!cancelled) {
          setUser(null);
        }
      } finally {
        if (!cancelled) {
          setIsHydrated(true);
        }
      }
    };

    void restoreSession();

    return () => {
      cancelled = true;
    };
  }, []);

  const login = useCallback((userData: User) => {
    setUser(userData);
    setIsHydrated(true);
  }, []);

  const refreshUser = useCallback((userData: User) => {
    setUser(userData);
  }, []);

  const clearAuth = useCallback(() => {
    setUser(null);
  }, []);

  const logout = useCallback(async () => {
    try {
      await authAPI.logout();
    } catch {
      // The server clears the cookie when it can. Clear local state regardless
      // so the UI never keeps showing a session the user asked to end.
    } finally {
      clearAuth();
    }
  }, [clearAuth]);

  const isAuthenticated = useCallback((): boolean => {
    return user !== null;
  }, [user]);

  const value = useMemo<AuthContextType>(() => ({
    user,
    isHydrated,
    login,
    refreshUser,
    logout,
    clearAuth,
    isAuthenticated
  }), [user, isHydrated, login, refreshUser, logout, clearAuth, isAuthenticated]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook for accessing the auth context from components
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export default useAuth;
