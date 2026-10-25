import { TagBudgetRule } from '../models';

function formatDateOnly(date) {
  if (!date) return undefined;
  if (typeof date === 'string') {
    const match = date.trim().match(/^(\d{4}-\d{2}-\d{2})/);
    if (match) return match[1];
  }
  const d = new Date(date);
  if (Number.isNaN(d.getTime())) return undefined;
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function mapProject(item) {
  if (!item) return null;
  return {
    id: item.project_id,
    name: item.name,
    description: item.description,
    mainCurrency: item.mainCurrency || item.main_currency || 'UAH',
    startDate: item.start_date || null,
    endDate: item.end_date || null,
    annualPlannedBudget: typeof item.annual_planned_budget === 'number'
      ? item.annual_planned_budget / 100
      : 0,
    rawAnnualPlannedBudget: item.annual_planned_budget || 0,
    tagBudgets: Array.isArray(item.tag_budgets)
      ? item.tag_budgets.map(TagBudgetRule.fromDTO).filter(Boolean)
      : [],
    shareToken: item.share_token,
    isShared: Boolean(item.is_shared),
    owner: item.owner ? { id: item.owner.person_id, name: item.owner.display_name } : null,
    participants: item.participants || [],
  };
}

export class ProjectsRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getProjects(limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const url = query ? `/projects?${query}` : '/projects';
    const response = await this.#request.get(url);

    const { projects = [] } = response;
    return projects.map((item) => mapProject(item));
  }

  async getProject(projectId) {
    const response = await this.#request.get(`/projects/${projectId}`);
    return mapProject(response);
  }

  async createProject({
    name,
    description,
    mainCurrency,
    startDate,
    endDate,
    tagBudgets = [],
  }) {
    const payload = {
      name,
      description,
      mainCurrency: mainCurrency || undefined,
      start_date: formatDateOnly(startDate),
      end_date: formatDateOnly(endDate),
      tag_budgets: Array.isArray(tagBudgets) && tagBudgets.length > 0
        ? tagBudgets.map((tb) => (typeof tb.toDTO === 'function' ? tb.toDTO() : tb))
        : undefined,
    };
    const response = await this.#request.post('/projects', payload);
    return mapProject(response);
  }

  async acceptShare(shareToken) {
    const response = await this.#request.post(`/projects/share/${shareToken}`);
    return mapProject(response);
  }

  async updateProjectSettings(projectId, { mainCurrency, startDate, endDate, tagBudgets }) {
    const payload = {
      mainCurrency,
      start_date: formatDateOnly(startDate),
      end_date: formatDateOnly(endDate),
      tag_budgets: Array.isArray(tagBudgets)
        ? tagBudgets.map((tb) => (typeof tb.toDTO === 'function' ? tb.toDTO() : tb))
        : undefined,
    };
    const response = await this.#request.patch(`/projects/${projectId}`, payload);
    return mapProject(response);
  }

  async updateProjectBudget(projectId, tagBudgets = []) {
    const payload = {
      tag_budgets: Array.isArray(tagBudgets)
        ? tagBudgets.map((tb) => (typeof tb.toDTO === 'function' ? tb.toDTO() : tb))
        : [],
    };
    const response = await this.#request.put(`/projects/${projectId}/budget`, payload);
    return mapProject(response);
  }

  async getTotals(projectId) {
    const response = await this.#request.get(`/projects/${projectId}/totals`)
      .catch(() => ({
        totals: [
          { title: 'Total Invested', amount: 0, currency: 'UAH' },
        ],
      }));

    const { totals = [] } = response;

    return totals.map(({ title, amount, currency }) => ({
      key: (title || 'total').toLowerCase().replace(/\s+/g, '-'),
      title,
      amount: new Intl.NumberFormat().format(amount / 100),
      currency,
    }));
  }
}

export default ProjectsRepository;
