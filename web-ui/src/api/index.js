import { Auth } from './auth';
import { Request } from './request';
import { ProjectsRepository } from './projects';
import { TypesRepository } from './types';
import { CategoriesRepository } from './categories';
import { PaymentRepository } from './payments';
import { AssetRepository } from './assets';
import WsClient from './ws-client';

const baseUrl = '/api';
export const authProvider = new Auth(baseUrl);
const request = new Request(baseUrl, authProvider);
const ws = new WsClient('/ws');

ws.onOpen(() => {
  setInterval(() => {
    ws.send({ type: 'ping' });
  }, 5000);
});
ws.onClose(console.log);
ws.onError(console.error);
ws.onMessage(console.log);

export const projectsRepository = new ProjectsRepository(request);
export const typesRepository = new TypesRepository(request);
export const categoriesRepository = new CategoriesRepository(request);
export const paymentRepository = new PaymentRepository(request);
export const assetRepository = new AssetRepository(request);
