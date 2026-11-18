import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authProvider } from '../api/index.js';
import useCurrentUser from '../hooks/useCurrentUser.js';

export const AuthenticatedOnly = (WrappedComponent) => function AuthenticatedGuard(props) {
  const navigate = useNavigate();
  const isAuthenticated = authProvider.isAuthenticated();

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [navigate, isAuthenticated]);

  if (!isAuthenticated) {
    return null;
  }

  return <WrappedComponent {...props} />;
};

export const AdminOnly = (WrappedComponent) => function AdminGuard(props) {
  const navigate = useNavigate();
  const { currentUser, isAuthenticated, isAdmin } = useCurrentUser();

  useEffect(() => {
    if (!authProvider.isAuthenticated()) {
      navigate('/login');
      return;
    }

    if (currentUser && !isAdmin) {
      navigate('/');
    }
  }, [navigate, currentUser, isAuthenticated, isAdmin]);

  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return <WrappedComponent {...props} />;
};
