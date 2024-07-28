import { Auth } from './auth';
import { Request } from './request.js';
import { ProjectsRepository } from './projects.js';
import { TypesRepository } from './types.js';
import { CategoriesRepository } from './categories.js';
import { ExpensesRepository } from './expenses.js';

const baseUrl = '/api';
export const authProvider = new Auth(baseUrl);
const request = new Request(baseUrl, authProvider);

export const projectsRepository = new ProjectsRepository(request);
export const typesRepository = new TypesRepository(request);
export const categoriesRepository = new CategoriesRepository(request);
export const expensesRepository = new ExpensesRepository(request);
