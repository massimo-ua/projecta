import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authProvider } from '../api/index.js';

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
