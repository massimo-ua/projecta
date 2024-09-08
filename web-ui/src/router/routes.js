import {
  Projects,
  ProjectDetails,
  Types,
  Categories,
  Payments,
  Total,
  Login,
  AuthenticatedOnly,
  Assets,
} from '../components';

const routes = [{
  path: '/',
  Component: AuthenticatedOnly(Projects),
}, {
  path: '/projects',
  Component: AuthenticatedOnly(Projects),
  exact: true,
}, {
  path: '/login',
  Component: Login,
}, {
  path: '/projects/:projectId',
  Component: AuthenticatedOnly(ProjectDetails),
  children: [{
    index: true,
    Component: AuthenticatedOnly(Payments),
  },{
    path: 'types',
    Component: AuthenticatedOnly(Types),
  }, {
    path: 'categories',
    Component: AuthenticatedOnly(Categories),
  }, {
    path: 'payments',
    Component: AuthenticatedOnly(Payments),
  }, {
    path: 'total',
    Component: AuthenticatedOnly(Total),
  }, {
    path: 'assets',
    Component: AuthenticatedOnly(Assets),
  }],
}];

export default routes;
