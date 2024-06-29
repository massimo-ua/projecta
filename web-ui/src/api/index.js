import { Auth } from './auth';
import { Request } from './request.js';
import { ProjectsRepository } from './projects.js';
import { TypesRepository } from './types.js';
import { CategoriesRepository } from './categories.js';

const baseUrl = 'http://localhost:5173/api';
const authProvider = new Auth();
const request = new Request(baseUrl, authProvider);

export const projectsRepository = new ProjectsRepository(request);
export const typesRepository = new TypesRepository(request);
export const categoriesRepository = new CategoriesRepository(request);
