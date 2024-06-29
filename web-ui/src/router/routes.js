import { Projects, ProjectDetails, Types, Categories, Expenses } from '../components';

export const routes = [ {
  path: '/',
  Component: Projects
}, {
  path: '/projects',
  Component: Projects,
  exact: true,
}, {
  path: '/projects/:projectId',
  Component: ProjectDetails,
  children: [{
    path: 'types',
    Component: Types,
  }, {
    path: 'categories',
    Component: Categories,
  }, {
    path: 'expenses',
    Component: Expenses,
  }],
}];
