import { format, formatISO, parseISO } from 'date-fns';

const toDomain = ({
                    asset_id, price, currency, description, type, category, acquired_at,
                  }) => ({
  key: asset_id,
  id: asset_id,
  description,
  price: (price / 100).toFixed(2),
  currency,
  type: type?.name,
  category: category?.name,
  paymentDate: format(parseISO(acquired_at), 'dd/MM/yyyy', { awareOfUnicodeTokens: true }),
});

const toAddAssetDTO = ({ typeId, price, currency, acquiredAt, name, description, withPayment }) => ({
  type_id: typeId,
  price: price * 100,
  currency,
  acquired_at: formatISO(acquiredAt, { representation: 'complete' }),
  description,
  name,
  with_payment: withPayment,
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

  async addAsset(projectId, asset) {
    const response = await this.#request.post(`/projects/${projectId}/assets`, toAddAssetDTO(asset));

    if (!response.ok) {
      throw new Error('Failed to add asset');
    }

    const json = await response.json();

    return toDomain(json);
  }

  async removeAsset(projectId, assetId) {
    const response = await this.#request.delete(`/projects/${projectId}/assets/${assetId}`);

    if (!response.ok) {
      throw new Error('Failed to remove asset');
    }
  }
}
