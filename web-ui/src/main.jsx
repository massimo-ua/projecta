import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { GoogleOAuthProvider } from '@react-oauth/google';
import { IntlayerProvider } from 'react-intlayer';

import HomeLayout from './Layout';
import './index.css';
import router from './router';
import { GOOGLE_CLIENT_ID } from './constants';

ReactDOM.createRoot(document.getElementById('root')).render(
  <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
    <React.StrictMode>
      <IntlayerProvider>
        <RouterProvider router={router}>
          <HomeLayout />
        </RouterProvider>
      </IntlayerProvider>
    </React.StrictMode>
  </GoogleOAuthProvider>,
);
