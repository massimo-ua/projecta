import { format, formatISO, parseISO } from 'date-fns';

const toDomain = ({
                    payment_id, amount, currency, description, type, category, payment_date,
                  }) => ({
  key: payment_id,
  id: payment_id,
  description,
  amount: (amount / 100).toFixed(2),
  currency,
  type: type?.name,
  category: category?.name,
  paymentDate: format(parseISO(payment_date), 'dd/MM/yyyy', { awareOfUnicodeTokens: true }),
});

const toAddPaymentDTO = ({ typeId, amount, currency, paymentDate, description, expenseKind }) => ({
  type_id: typeId,
  amount: amount * 100,
  currency,
  payment_date: formatISO(paymentDate, { representation: 'complete' }),
  description,
  kind: expenseKind,
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

    if (!response.ok) {
      throw new Error('Failed to add payment');
    }

    const json = await response.json();

    return toDomain(json);
  }

  async removePayment(projectId, paymentId) {
    const response = await this.#request.delete(`/projects/${projectId}/payments/${paymentId}`);

    if (!response.ok) {
      throw new Error('Failed to remove expense');
    }
  }
}
