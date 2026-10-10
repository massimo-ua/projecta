import React from 'react';
import { Navigate } from 'react-router-dom';
import { localeMap } from 'intlayer';
import {
  Projects,
  ProjectDetails,
  Total,
  Settings,
  Login,
  AuthenticatedOnly,
  AdminOnly,
  Assets,
  Payments,
  UserProfileSettings,
  RolesManagement,
  InvitationsManagement,
  AcceptInvitation,
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
    path: `${prefix}/admin/roles`,
    Component: AdminOnly(RolesManagement),
    errorElement,
  },
  {
    path: `${prefix}/admin/invitations`,
    Component: AdminOnly(InvitationsManagement),
    errorElement,
  },
  {
    path: `${prefix}/invite/:code`,
    Component: AcceptInvitation,
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
        Component: AuthenticatedOnly(Assets),
        errorElement,
      },
      {
        path: 'assets',
        Component: AuthenticatedOnly(Assets),
        errorElement,
      },
      {
        path: 'investments',
        Component: AuthenticatedOnly(Payments),
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
        path: 'settings',
        Component: AuthenticatedOnly(Settings),
        errorElement,
      },
      {
        path: 'types',
        Component: () => React.createElement(Navigate, { to: '../assets', replace: true }),
        errorElement,
      },
      {
        path: 'categories',
        Component: () => React.createElement(Navigate, { to: '../assets', replace: true }),
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
