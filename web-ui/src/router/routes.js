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
} from '../components';

const NotFoundRedirect = () => React.createElement(Navigate, { to: '/', replace: true });

const createRoutesForPrefix = (prefix) => [
  {
    path: prefix || '/',
    Component: AuthenticatedOnly(Projects),
  },
  {
    path: `${prefix}/projects`,
    Component: AuthenticatedOnly(Projects),
    exact: true,
  },
  {
    path: `${prefix}/profile`,
    Component: AuthenticatedOnly(UserProfileSettings),
  },
  {
    path: `${prefix}/login`,
    Component: Login,
  },
  {
    path: `${prefix}/projects/:projectId`,
    Component: AuthenticatedOnly(ProjectDetails),
    children: [
      {
        index: true,
        Component: AuthenticatedOnly(Payments),
      },
      {
        path: 'types',
        Component: AuthenticatedOnly(Types),
      },
      {
        path: 'categories',
        Component: AuthenticatedOnly(Categories),
      },
      {
        path: 'settings',
        Component: AuthenticatedOnly(Settings),
      },
      {
        path: 'payments',
        Component: AuthenticatedOnly(Payments),
        exact: true,
      },
      {
        path: 'total',
        Component: AuthenticatedOnly(Total),
      },
      {
        path: 'assets',
        Component: AuthenticatedOnly(Assets),
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
