import { format, formatISO, parseISO } from 'date-fns';
import {
  fromISO, toISO, toPrice, toPriceView,
} from './mappers';
import { Payment } from '../models/Payment';

const toDomain = (item) => {
  if (!item) return null;
  const id = item.payment_id || item.investment_id;
  const date = item.payment_date || item.date;
  const amount = item.amount;
  const finalHomeAmount = item.home_amount !== undefined ? item.home_amount : (item.homeAmount !== undefined ? item.homeAmount : amount);
  const finalHomeCurrency = item.home_currency || item.homeCurrency || item.project?.mainCurrency || item.currency;

  let categoryLabel = item.type?.category?.name;
  if (!categoryLabel && Array.isArray(item.tags) && item.tags.length > 0) {
    categoryLabel = item.tags.join(', ');
  } else if (!categoryLabel) {
    categoryLabel = item.resource_type || 'Capital';
  }

  let typeLabel = item.type?.name;
  if (!typeLabel) {
    typeLabel = item.asset_name || 'General Operations';
  }

  let formattedDate = '';
  if (date) {
    try {
      formattedDate = format(parseISO(date), 'dd/MM/yyyy', { awareOfUnicodeTokens: true });
    } catch {
      formattedDate = date;
    }
  }

  return new Payment({
    key: id,
    id: id,
    description: item.description,
    amount: toPriceView(amount),
    rawAmount: amount,
    currency: item.currency,
    homeAmount: toPriceView(finalHomeAmount),
    rawHomeAmount: finalHomeAmount,
    homeCurrency: finalHomeCurrency,
    type: typeLabel,
    typeId: item.asset_id || item.type?.type_id,
    category: categoryLabel,
    paymentDate: formattedDate,
    rawDate: date,
    kind: item.resource_type || item.kind,
  });
};

const toAddPaymentDTO = ({
  typeId, assetId, amount, currency, paymentDate, date, description, paymentKind, resourceType, tags,
}) => ({
  asset_id: assetId || typeId,
  amount: toPrice(amount),
  currency: currency || 'UAH',
  date: formatISO(paymentDate || date || new Date(), { representation: 'complete' }),
  payment_date: formatISO(paymentDate || date || new Date(), { representation: 'complete' }),
  description: description || '',
  kind: paymentKind,
  resource_type: resourceType || 'MONEY',
  tags: tags || [],
});

const toUpdatePaymentDTO = ({
  typeId, assetId, amount, currency, paymentDate, date, description, paymentKind, resourceType, tags,
}) => ({
  asset_id: assetId || typeId,
  amount: toPrice(amount),
  currency: currency || 'UAH',
  date: toISO(paymentDate || date || new Date()),
  payment_date: toISO(paymentDate || date || new Date()),
  description: description || '',
  kind: paymentKind,
  resource_type: resourceType || 'MONEY',
  tags: tags || [],
});

const toEditPaymentView = (item) => ({
  id: item.payment_id || item.investment_id,
  amount: toPriceView(item.amount),
  currency: item.currency,
  description: item.description,
  assetId: item.asset_id,
  typeId: item.type?.type_id || item.asset_id,
  categoryId: item.type?.category?.category_id,
  paymentDate: fromISO(item.payment_date || item.date),
  kind: item.resource_type || item.kind,
  tags: item.tags || [],
});

export class PaymentRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getPayments(projectId, limit = 10, offset = 0, filters = {}) {
    const params = { limit: String(limit), offset: String(offset) };
    if (filters.assetId || filters.typeId) params.asset_id = filters.assetId || filters.typeId;
    if (filters.fromDate) params.from_date = filters.fromDate;
    if (filters.toDate) params.to_date = filters.toDate;
    if (filters.tag) params.tag = filters.tag;
    if (filters.resourceType) params.resource_type = filters.resourceType;

    const query = new URLSearchParams(params).toString();
    try {
      const url = `/projects/${projectId}/investments${query ? `?${query}` : ''}`;
      const response = await this.#request.get(url);
      const items = response.investments || [];
      const total = response.total !== undefined ? response.total : items.length;
      return [items.map(toDomain), total];
    } catch {
      const resourceUrl = `/projects/${projectId}/payments`;
      const url = query ? `${resourceUrl}?${query}` : resourceUrl;
      const response = await this.#request.get(url);
      const { payments, total } = response;
      return [(payments || []).map(toDomain), total];
    }
  }

  async addPayment(projectId, payment) {
    try {
      const response = await this.#request.post(
        `/projects/${projectId}/investments`,
        toAddPaymentDTO(payment),
      );
      return toDomain(response);
    } catch {
      const response = await this.#request.post(
        `/projects/${projectId}/payments`,
        toAddPaymentDTO(payment),
      );
      return toDomain(response);
    }
  }

  async removePayment(projectId, paymentId) {
    try {
      return await this.#request.delete(`/projects/${projectId}/investments/${paymentId}`);
    } catch {
      return await this.#request.delete(`/projects/${projectId}/payments/${paymentId}`);
    }
  }

  async getPayment(projectId, paymentId) {
    try {
      const response = await this.#request.get(`/projects/${projectId}/investments/${paymentId}`);
      return toEditPaymentView(response);
    } catch {
      const response = await this.#request.get(`/projects/${projectId}/payments/${paymentId}`);
      return toEditPaymentView(response);
    }
  }

  async updatePayment(projectId, payment) {
    try {
      return await this.#request.put(
        `/projects/${projectId}/investments/${payment.id}`,
        toUpdatePaymentDTO(payment),
      );
    } catch {
      return await this.#request.put(
        `/projects/${projectId}/payments/${payment.id}`,
        toUpdatePaymentDTO(payment),
      );
    }
  }

  async parseStatement(projectId, file) {
    const formData = new FormData();
    formData.append('file', file);
    return await this.#request.post(`/projects/${projectId}/payments/parse-statement`, formData);
  }

  async addPaymentsBatch(projectId, payments) {
    const dtos = payments.map(toAddPaymentDTO);
    try {
      const response = await this.#request.post(
        `/projects/${projectId}/investments/batch`,
        { items: dtos },
      );
      return Array.isArray(response) ? response.map(toDomain) : [];
    } catch {
      const response = await this.#request.post(`/projects/${projectId}/payments/batch`, dtos);
      return Array.isArray(response) ? response.map(toDomain) : [];
    }
  }
}

export default PaymentRepository;
