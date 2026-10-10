import {
  fromISO, toDateView, toISO, toPrice, toPriceView,
} from './mappers';
import { Asset } from '../models/Asset';

const toDomain = ({
  asset_id,
  name,
  description,
  status,
  start_date,
  completed_date,
  target_price,
  target_currency,
  direct_cost,
  total_cost,
  price,
  currency,
  home_amount,
  home_currency,
  homeAmount,
  homeCurrency,
  acquired_at,
  children = [],
  parents = [],
  type,
  project,
}) => {
  const finalPrice = total_cost !== undefined ? total_cost : (price !== undefined ? price : 0);
  const finalHomeAmount = home_amount !== undefined ? home_amount : (homeAmount !== undefined ? homeAmount : finalPrice);
  const finalHomeCurrency = home_currency || homeCurrency || project?.mainCurrency || currency;
  const finalCurrency = currency || target_currency || finalHomeCurrency;

  const childrenMapped = (children || []).map((c) => ({
    childId: c.child_id,
    childName: c.child_name,
    sharePercentage: c.share_percentage,
    totalCost: toPriceView(c.total_cost || 0),
    rawTotalCost: c.total_cost || 0,
  }));

  const parentsMapped = (parents || []).map((p) => ({
    parentId: p.parent_id,
    parentName: p.parent_name,
    sharePercentage: p.share_percentage,
  }));

  return new Asset({
    key: asset_id,
    id: asset_id,
    name,
    description,
    status: status || 'ACTIVE',
    startDate: toDateView(start_date || acquired_at),
    rawStartDate: start_date || acquired_at,
    completedDate: completed_date ? toDateView(completed_date) : null,
    rawCompletedDate: completed_date,
    targetPrice: target_price ? toPriceView(target_price) : null,
    rawTargetPrice: target_price,
    targetCurrency: target_currency || finalCurrency,
    directCost: toPriceView(direct_cost || 0),
    rawDirectCost: direct_cost || 0,
    totalCost: toPriceView(finalPrice),
    rawTotalCost: finalPrice,
    price: toPriceView(finalPrice),
    rawPrice: finalPrice,
    currency: finalCurrency,
    homeAmount: toPriceView(finalHomeAmount),
    homeCurrency: finalHomeCurrency,
    acquiredAt: toDateView(acquired_at || start_date),
    rawAcquiredAt: acquired_at || start_date,
    children: childrenMapped,
    parents: parentsMapped,
    type: type?.name,
    category: type?.category?.name,
  });
};

const toEditAssetView = ({
  asset_id,
  name,
  description,
  status,
  start_date,
  completed_date,
  target_price,
  target_currency,
  price,
  currency,
  acquired_at,
}) => ({
  id: asset_id,
  name,
  description,
  status: status || 'ACTIVE',
  startDate: fromISO(start_date || acquired_at),
  completedDate: completed_date ? fromISO(completed_date) : '',
  targetPrice: target_price ? toPriceView(target_price) : '',
  targetCurrency: target_currency || currency || 'UAH',
  price: toPriceView(price || 0),
  currency: currency || target_currency || 'UAH',
  acquiredAt: fromISO(acquired_at || start_date),
});

const toAddAssetDTO = ({
  name,
  description,
  status = 'ACTIVE',
  startDate,
  acquiredAt,
  completedDate,
  targetPrice,
  targetCurrency,
  price,
  currency,
  withPayment,
}) => {
  const chosenPrice = targetPrice ? toPrice(targetPrice) : (price ? toPrice(price) : 0);
  const chosenCurrency = targetCurrency || currency || 'UAH';
  const effectiveDate = startDate || acquiredAt || new Date();

  const dto = {
    name,
    description: description || '',
    status: status || 'ACTIVE',
    start_date: toISO(effectiveDate),
    acquired_at: toISO(effectiveDate),
    target_price: chosenPrice,
    target_currency: chosenCurrency,
    price: chosenPrice,
    currency: chosenCurrency,
    with_payment: Boolean(withPayment),
  };

  if (completedDate) {
    dto.completed_date = toISO(completedDate);
  }

  return dto;
};

const toUpdateAssetDTO = ({
  id,
  name,
  description,
  status,
  startDate,
  acquiredAt,
  completedDate,
  targetPrice,
  targetCurrency,
  price,
  currency,
}) => {
  const chosenPrice = targetPrice ? toPrice(targetPrice) : (price ? toPrice(price) : 0);
  const chosenCurrency = targetCurrency || currency || 'UAH';
  const effectiveDate = startDate || acquiredAt || new Date();

  const dto = {
    name,
    description: description || '',
    status: status || 'ACTIVE',
    start_date: toISO(effectiveDate),
    acquired_at: toISO(effectiveDate),
    target_price: chosenPrice,
    target_currency: chosenCurrency,
    price: chosenPrice,
    currency: chosenCurrency,
  };

  if (completedDate) {
    dto.completed_date = toISO(completedDate);
  }

  return dto;
};

export class AssetRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getAssets(projectId, limit = 50, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const resourceUrl = `/projects/${projectId}/assets`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { assets, total } = response;
    return [assets.map(toDomain), total];
  }

  async getAsset(projectId, assetId) {
    const response = await this.#request.get(`/projects/${projectId}/assets/${assetId}`);
    return toEditAssetView(response);
  }

  async updateAsset(projectId, asset) {
    return await this.#request.put(
      `/projects/${projectId}/assets/${asset.id}`,
      toUpdateAssetDTO(asset),
    );
  }

  async addAsset(projectId, asset) {
    const response = await this.#request.post(`/projects/${projectId}/assets`, toAddAssetDTO(asset));
    return toDomain(response);
  }

  async linkChild(projectId, parentId, childId, sharePercentage = 100) {
    return await this.#request.post(
      `/projects/${projectId}/assets/${parentId}/children`,
      {
        child_asset_id: childId,
        share_percentage: Number(sharePercentage),
      },
    );
  }

  async unlinkChild(projectId, parentId, childId) {
    return await this.#request.delete(
      `/projects/${projectId}/assets/${parentId}/children/${childId}`,
    );
  }

  async createAssetFromPayments(projectId, {
    paymentIds, name, description, acquiredAt, startDate, targetCurrency,
  }) {
    const payload = {
      payment_ids: paymentIds,
      name,
      description,
    };
    if (acquiredAt || startDate) payload.acquired_at = toISO(acquiredAt || startDate);
    if (targetCurrency) payload.target_currency = targetCurrency;

    const response = await this.#request.post(`/projects/${projectId}/assets/from-payments`, payload);
    return toDomain(response);
  }

  async removeAsset(projectId, assetId) {
    return await this.#request.delete(`/projects/${projectId}/assets/${assetId}`);
  }
}

export default AssetRepository;
