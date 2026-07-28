import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { IntlayerProvider } from 'react-intlayer';

import './index.css';
import router from './router';
import { ErrorBoundary } from './components';
import { GOOGLE_CLIENT_ID } from './constants';

ReactDOM.createRoot(document.getElementById('root')).render(
  <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
    <React.StrictMode>
      <IntlayerProvider>
        <ErrorBoundary>
          <RouterProvider router={router} />
        </ErrorBoundary>
      </IntlayerProvider>
    </React.StrictMode>
  </GoogleOAuthProvider>,
);
