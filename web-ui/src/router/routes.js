import React from 'react';
import { Navigate } from 'react-router-dom';
import { localeMap } from 'intlayer';
import {
  Projects,
  ProjectDetails,
  Types,
  Categories,
  Total,
  Settings,
  Login,
  AuthenticatedOnly,
  Assets,
  Payments,
  UserProfileSettings,
  ErrorPage,
  AcceptShare,
} from '../components';

const NotFoundRedirect = () => React.createElement(Navigate, { to: '/', replace: true });
const errorElement = React.createElement(ErrorPage);

const createRoutesForPrefix = (prefix) => [
  {
    path: prefix || '/',
    Component: AuthenticatedOnly(Projects),
    errorElement,
  },
  {
    path: `${prefix}/projects`,
    Component: AuthenticatedOnly(Projects),
    exact: true,
    errorElement,
  },
  {
    path: `${prefix}/projects/share/:shareToken`,
    Component: AuthenticatedOnly(AcceptShare),
    errorElement,
  },
  {
    path: `${prefix}/profile`,
    Component: AuthenticatedOnly(UserProfileSettings),
    errorElement,
  },
  {
    path: `${prefix}/login`,
    Component: Login,
    errorElement,
  },
  {
    path: `${prefix}/projects/:projectId`,
    Component: AuthenticatedOnly(ProjectDetails),
    errorElement,
    children: [
      {
        index: true,
        Component: AuthenticatedOnly(Payments),
        errorElement,
      },
      {
        path: 'types',
        Component: AuthenticatedOnly(Types),
        errorElement,
      },
      {
        path: 'categories',
        Component: AuthenticatedOnly(Categories),
        errorElement,
      },
      {
        path: 'settings',
        Component: AuthenticatedOnly(Settings),
        errorElement,
      },
      {
        path: 'payments',
        Component: AuthenticatedOnly(Payments),
        exact: true,
        errorElement,
      },
      {
        path: 'total',
        Component: AuthenticatedOnly(Total),
        errorElement,
      },
      {
        path: 'assets',
        Component: AuthenticatedOnly(Assets),
        errorElement,
      },
    ],
  },
];

const routes = [
  ...localeMap(({ urlPrefix }) => createRoutesForPrefix(urlPrefix)).flat(),
  {
    path: '*',
    Component: NotFoundRedirect,
  },
];

export default routes;
