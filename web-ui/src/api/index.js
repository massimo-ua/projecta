import { Auth } from './auth';
import { Request } from './request';
import { ProjectsRepository } from './projects';
import { TypesRepository } from './types';
import { CategoriesRepository } from './categories';
import { PaymentRepository } from './payments';

const baseUrl = '/api';
export const authProvider = new Auth(baseUrl);
const request = new Request(baseUrl, authProvider);

export const projectsRepository = new ProjectsRepository(request);
export const typesRepository = new TypesRepository(request);
export const categoriesRepository = new CategoriesRepository(request);
export const paymentRepository = new PaymentRepository(request);
