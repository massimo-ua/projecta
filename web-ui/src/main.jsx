import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';

import HomeLayout from './Layout';
import './index.css';
import router from './router';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <RouterProvider router={router}>
      <HomeLayout />
    </RouterProvider>
  </React.StrictMode>,
);
