import { useState, useEffect, useCallback } from 'react';
import { authProvider } from '../api';

export function useCurrentUser() {
  const [currentUser, setCurrentUser] = useState(() => authProvider.getUser());

  const refreshUser = useCallback(() => {
    setCurrentUser(authProvider.getUser());
  }, []);

  useEffect(() => {
    const handleAuthChange = () => {
      refreshUser();
    };

    window.addEventListener('auth-status-changed', handleAuthChange);
    window.addEventListener('storage', handleAuthChange);

    return () => {
      window.removeEventListener('auth-status-changed', handleAuthChange);
      window.removeEventListener('storage', handleAuthChange);
    };
  }, [refreshUser]);

  return {
    currentUser,
    isAuthenticated: Boolean(currentUser),
    isAdmin: Boolean(currentUser?.isAdmin),
    refreshUser,
  };
}

export default useCurrentUser;
