import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authProvider } from '../api/index.js';
import useCurrentUser from '../hooks/useCurrentUser.js';

export const AuthenticatedOnly = (WrappedComponent) => function (props) {
  const navigate = useNavigate();

  useEffect(() => {
    const isAuthenticated = authProvider.isAuthenticated();
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [navigate]);

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
