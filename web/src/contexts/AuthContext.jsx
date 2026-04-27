import { createContext, useContext, useState, useEffect } from 'react';
import { api } from '../lib/api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  async function checkAuth() {
    try {
      if (api.getToken()) {
        const userData = await api.getCurrentUser();
        setUser(userData);
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      api.logout();
    } finally {
      setLoading(false);
    }
  }

  async function login(email, password) {
    const data = await api.login(email, password);
    setUser(data.user);
    return data;
  }

  async function register(email, password, displayName) {
    const data = await api.register(email, password, displayName);
    setUser(data.user);
    return data;
  }

  function logout() {
    api.logout();
    setUser(null);
  }

  const value = {
    user,
    loading,
    login,
    register,
    logout,
    isAuthenticated: !!user,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
