import { format, parseISO } from 'date-fns';

export class ExpensesRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getExpenses(projectId, limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const resourceUrl = `/projects/${projectId}/expenses`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { expenses } = response;
    return expenses.map(({
      expense_id, amount, currency, description, type, category, expense_date,
    }) => ({
      key: expense_id,
      id: expense_id,
      description,
      amount: (amount / 100).toFixed(2),
      currency,
      type: type?.name,
      category: category?.name,
      expenseDate: format(parseISO(expense_date), 'dd/MM/yyyy', { awareOfUnicodeTokens: true }),
    }));
  }
}
