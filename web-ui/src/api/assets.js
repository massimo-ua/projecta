import { fromISO, toDateView, toISO, toPrice, toPriceView } from './mappers';

const toDomain = ({
                    asset_id, price, currency, description, type, acquired_at, name,
                  }) => ({
  key: asset_id,
  id: asset_id,
  description,
  price: toPriceView(price),
  currency,
  type: type?.name,
  category: type?.category?.name,
  acquiredAt: toDateView(acquired_at),
  name,
});

const toEditAssetView = ({
  asset_id, price, currency, description, type, acquired_at, name,
}) => ({
  id: asset_id, price: toPriceView(price), currency, description, typeId: type.type_id, acquiredAt: fromISO(acquired_at), name,
});

const toAddAssetDTO = ({ typeId, price, currency, acquiredAt, name, description, withPayment }) => ({
  type_id: typeId,
  price: toPrice(price),
  currency,
  acquired_at: toISO(acquiredAt),
  description,
  name,
  with_payment: withPayment,
});

const toUpdateAssetDTO = ({ typeId, price, currency, acquiredAt, name, description }) => ({
  type_id: typeId,
  price: toPrice(price),
  currency,
  acquired_at: toISO(acquiredAt),
  description,
  name,
});

export class AssetRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getAssets(projectId, limit = 10, offset = 0) {
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
      toUpdateAssetDTO(asset));
  }

  async addAsset(projectId, asset) {
    const response = await this.#request.post(`/projects/${projectId}/assets`, toAddAssetDTO(asset));

    return toDomain(response);
  }

  async removeAsset(projectId, assetId) {
    return await this.#request.delete(`/projects/${projectId}/assets/${assetId}`);
  }
}
