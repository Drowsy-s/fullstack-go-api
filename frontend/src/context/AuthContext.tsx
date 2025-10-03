import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { fetchProfile, loginUser, registerUser } from '../api';
import type { AuthResponse, User } from '../types';

interface AuthContextValue {
  token: string | null;
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);
const STORAGE_KEY = 'fullstack-go-token';

function usePersistedToken(): [string | null, (token: string | null) => void] {
  const [token, setTokenState] = useState<string | null>(() => localStorage.getItem(STORAGE_KEY));

  const setToken = useCallback((value: string | null) => {
    setTokenState(value);
    if (value) {
      localStorage.setItem(STORAGE_KEY, value);
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  }, []);

  return [token, setToken];
}

function handleAuthResponse(response: AuthResponse, setToken: (token: string | null) => void, setUser: (user: User | null) => void) {
  setToken(response.token);
  setUser(response.user);
}

export const AuthProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [token, setToken] = usePersistedToken();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    let active = true;

    const loadProfile = async () => {
      if (!token) {
        setUser(null);
        setLoading(false);
        return;
      }

      setLoading(true);
      try {
        const { user: profile } = await fetchProfile(token);
        if (active) {
          setUser(profile);
        }
      } catch (error) {
        if (active) {
          setToken(null);
          setUser(null);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    };

    void loadProfile();

    return () => {
      active = false;
    };
  }, [token, setToken]);

  const login = useCallback(
    async (email: string, password: string) => {
      setLoading(true);
      try {
        const response = await loginUser({ email, password });
        handleAuthResponse(response, setToken, setUser);
      } finally {
        setLoading(false);
      }
    },
    [setToken, setUser],
  );

  const register = useCallback(
    async (name: string, email: string, password: string) => {
      setLoading(true);
      try {
        const response = await registerUser({ name, email, password });
        handleAuthResponse(response, setToken, setUser);
      } finally {
        setLoading(false);
      }
    },
    [setToken, setUser],
  );

  const logout = useCallback(() => {
    setToken(null);
    setUser(null);
  }, [setToken]);

  const refreshProfile = useCallback(async () => {
    if (!token) {
      setUser(null);
      return;
    }
    const { user: profile } = await fetchProfile(token);
    setUser(profile);
  }, [token, setUser]);

  const value = useMemo(
    () => ({ token, user, loading, login, register, logout, refreshProfile }),
    [token, user, loading, login, register, logout, refreshProfile],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
