import { format, formatISO, parseISO } from 'date-fns';

const toDomain = ({
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
});

const toAddExpenseDTO = ({ typeId, amount, currency, expenseDate, description, expenseKind, fromDownPayment }) => ({
  type_id: typeId,
  amount: amount * 100,
  currency,
  expense_date: formatISO(expenseDate, { representation: 'complete' }),
  description,
  kind: expenseKind,
  from_down_payment: fromDownPayment,
});

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

    const { expenses, total } = response;
    return [expenses.map(toDomain), total];
  }

  async addExpense(projectId, expense) {
    const response = await this.#request.post(`/projects/${projectId}/expenses`, toAddExpenseDTO(expense));

    if (!response.ok) {
      throw new Error('Failed to add expense');
    }

    const json = await response.json();

    return toDomain(json);
  }

  async removeExpense(projectId, expenseId) {
    const response = await this.#request.delete(`/projects/${projectId}/expenses/${expenseId}`);

    if (!response.ok) {
      throw new Error('Failed to remove expense');
    }
  }
}
