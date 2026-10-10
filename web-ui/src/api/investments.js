import {
  fromISO, toDateView, toISO, toPrice, toPriceView,
} from './mappers';
import { Investment } from '../models/Investment';

const toDomain = ({
  investment_id,
  project_id,
  asset_id,
  asset_name,
  contributor,
  resource_type,
  amount,
  currency,
  home_amount,
  home_currency,
  time_hours,
  time_hourly_rate,
  goods_quantity,
  goods_unit,
  goods_item_name,
  description,
  date,
  tags,
}) => {
  const finalHomeAmount = home_amount !== undefined ? home_amount : amount;
  const finalHomeCurrency = home_currency || currency;

  return new Investment({
    key: investment_id,
    id: investment_id,
    projectId: project_id,
    assetId: asset_id,
    assetName: asset_name,
    contributor,
    resourceType: resource_type || 'MONEY',
    amount: toPriceView(amount),
    rawAmount: amount,
    currency,
    homeAmount: toPriceView(finalHomeAmount),
    rawHomeAmount: finalHomeAmount,
    homeCurrency: finalHomeCurrency,
    timeHours: time_hours,
    timeHourlyRate: time_hourly_rate ? toPriceView(time_hourly_rate) : null,
    rawTimeHourlyRate: time_hourly_rate,
    goodsQuantity: goods_quantity,
    goodsUnit: goods_unit,
    goodsItemName: goods_item_name,
    description,
    date: toDateView(date),
    rawDate: date,
    tags: tags || [],
  });
};

const toEditInvestmentView = ({
  investment_id,
  asset_id,
  asset_name,
  resource_type,
  amount,
  currency,
  time_hours,
  time_hourly_rate,
  goods_quantity,
  goods_unit,
  goods_item_name,
  description,
  date,
  tags,
}) => ({
  id: investment_id,
  assetId: asset_id,
  assetName: asset_name,
  resourceType: resource_type || 'MONEY',
  amount: toPriceView(amount),
  currency: currency || 'UAH',
  timeHours: time_hours || '',
  timeHourlyRate: time_hourly_rate ? toPriceView(time_hourly_rate) : '',
  goodsQuantity: goods_quantity || '',
  goodsUnit: goods_unit || '',
  goodsItemName: goods_item_name || '',
  description: description || '',
  date: fromISO(date),
  tags: tags || [],
});

const toAddInvestmentDTO = ({
  assetId,
  resourceType = 'MONEY',
  amount,
  currency = 'UAH',
  timeHours,
  timeHourlyRate,
  goodsQuantity,
  goodsUnit,
  goodsItemName,
  description,
  date,
  tags,
}) => {
  const dto = {
    asset_id: assetId,
    resource_type: resourceType,
    amount: toPrice(amount),
    currency,
    description: description || '',
    date: toISO(date || new Date()),
    tags: Array.isArray(tags) ? tags : [],
  };

  if (timeHours !== undefined && timeHours !== '') {
    dto.time_hours = Number(timeHours);
  }
  if (timeHourlyRate !== undefined && timeHourlyRate !== '') {
    dto.time_hourly_rate = toPrice(timeHourlyRate);
  }
  if (goodsQuantity !== undefined && goodsQuantity !== '') {
    dto.goods_quantity = Number(goodsQuantity);
  }
  if (goodsUnit) {
    dto.goods_unit = goodsUnit;
  }
  if (goodsItemName) {
    dto.goods_item_name = goodsItemName;
  }

  return dto;
};

const toUpdateInvestmentDTO = ({
  assetId,
  resourceType = 'MONEY',
  amount,
  currency = 'UAH',
  timeHours,
  timeHourlyRate,
  goodsQuantity,
  goodsUnit,
  goodsItemName,
  description,
  date,
  tags,
}) => {
  const dto = {
    asset_id: assetId,
    resource_type: resourceType,
    amount: toPrice(amount),
    currency,
    description: description || '',
    date: toISO(date || new Date()),
    tags: Array.isArray(tags) ? tags : [],
  };

  if (timeHours !== undefined && timeHours !== '') {
    dto.time_hours = Number(timeHours);
  }
  if (timeHourlyRate !== undefined && timeHourlyRate !== '') {
    dto.time_hourly_rate = toPrice(timeHourlyRate);
  }
  if (goodsQuantity !== undefined && goodsQuantity !== '') {
    dto.goods_quantity = Number(goodsQuantity);
  }
  if (goodsUnit !== undefined) {
    dto.goods_unit = goodsUnit;
  }
  if (goodsItemName !== undefined) {
    dto.goods_item_name = goodsItemName;
  }

  return dto;
};

export class InvestmentRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getInvestments(projectId, limit = 50, offset = 0, filters = {}) {
    const params = { limit: String(limit), offset: String(offset) };
    if (filters.assetId) params.asset_id = filters.assetId;
    if (filters.resourceType) params.resource_type = filters.resourceType;
    if (filters.tag) params.tag = filters.tag;

    const query = new URLSearchParams(params).toString();
    const resourceUrl = `/projects/${projectId}/investments`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { investments, total } = response;
    return [investments.map(toDomain), total];
  }

  async getInvestment(projectId, investmentId) {
    const response = await this.#request.get(
      `/projects/${projectId}/investments/${investmentId}`,
    );
    return toEditInvestmentView(response);
  }

  async addInvestment(projectId, investment) {
    const response = await this.#request.post(
      `/projects/${projectId}/investments`,
      toAddInvestmentDTO(investment),
    );
    return toDomain(response);
  }

  async updateInvestment(projectId, investment) {
    return await this.#request.put(
      `/projects/${projectId}/investments/${investment.id}`,
      toUpdateInvestmentDTO(investment),
    );
  }

  async removeInvestment(projectId, investmentId) {
    return await this.#request.delete(
      `/projects/${projectId}/investments/${investmentId}`,
    );
  }

  async addInvestmentsBatch(projectId, { defaultAsset, items = [] }) {
    const payload = {
      default_asset: defaultAsset,
      items: items.map((it) => toAddInvestmentDTO({ ...it, assetId: it.assetId || defaultAsset })),
    };
    const response = await this.#request.post(
      `/projects/${projectId}/investments/batch`,
      payload,
    );
    return Array.isArray(response) ? response.map(toDomain) : [];
  }

  async getTags(projectId) {
    const response = await this.#request.get(`/projects/${projectId}/tags`);
    return Array.isArray(response) ? response : [];
  }

  async parseStatement(projectId, file) {
    const formData = new FormData();
    formData.append('file', file);
    return await this.#request.post(`/projects/${projectId}/payments/parse-statement`, formData);
  }
}

export default InvestmentRepository;
