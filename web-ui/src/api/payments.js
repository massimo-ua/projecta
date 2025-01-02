import { format, formatISO, parseISO } from 'date-fns';
import { fromISO, toISO, toPrice, toPriceView } from './mappers';

const toDomain = ({
                    payment_id, amount, currency, description, type, payment_date, kind,
                  }) => ({
  key: payment_id,
  id: payment_id,
  description,
  amount: toPriceView(amount),
  currency,
  type: type?.name,
  category: type.category?.name,
  paymentDate: format(parseISO(payment_date), 'dd/MM/yyyy', { awareOfUnicodeTokens: true }),
  kind,
});

const toAddPaymentDTO = ({ typeId, amount, currency, paymentDate, description, paymentKind }) => ({
  type_id: typeId,
  amount: toPrice(amount),
  currency,
  payment_date: formatISO(paymentDate, { representation: 'complete' }),
  description,
  kind: paymentKind,
});

const toUpdatePaymentDTO = ({ typeId, amount, currency, paymentDate, description, paymentKind }) => ({
  type_id: typeId,
  amount: toPrice(amount),
  currency,
  payment_date: toISO(paymentDate),
  description,
  kind: paymentKind,
});

const toEditPaymentView = ({
  payment_id, amount, currency, description, type, payment_date, kind,
}) => ({
  id: payment_id,
  amount: toPriceView(amount),
  currency,
  description,
  typeId: type.type_id,
  categoryId: type.category?.category_id,
  paymentDate: fromISO(payment_date),
  kind,
});

export class PaymentRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getPayments(projectId, limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const resourceUrl = `/projects/${projectId}/payments`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { payments, total } = response;
    return [payments.map(toDomain), total];
  }

  async addPayment(projectId, payment) {
    const response = await this.#request.post(`/projects/${projectId}/payments`, toAddPaymentDTO(payment));

    return toDomain(response);
  }

  async removePayment(projectId, paymentId) {
    return await this.#request.delete(`/projects/${projectId}/payments/${paymentId}`);
  }

  async getPayment(projectId, paymentId) {
    const response = await this.#request.get(`/projects/${projectId}/payments/${paymentId}`);

    return toEditPaymentView(response);
  }

  async updatePayment(projectId, payment) {
    return await this.#request.put(
      `/projects/${projectId}/payments/${payment.id}`,
      toUpdatePaymentDTO(payment));
  }
}
