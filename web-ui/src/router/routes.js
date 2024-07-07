import {
  Projects,
  ProjectDetails,
  Types,
  Categories,
  Expenses,
  Total,
  Login,
  AuthenticatedOnly,
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
    path: 'types',
    Component: AuthenticatedOnly(Types),
  }, {
    path: 'categories',
    Component: AuthenticatedOnly(Categories),
  }, {
    path: 'expenses',
    Component: AuthenticatedOnly(Expenses),
  }, {
    path: 'total',
    Component: AuthenticatedOnly(Total),
  }],
}];

export default routes;
