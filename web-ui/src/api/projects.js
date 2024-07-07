export class ProjectsRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getProjects(limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const url = query ? `/projects?${query}` : '/projects';
    const response = await this.#request.get(url);

    const { projects } = response;
    return projects.map(({ project_id, name, description }) => ({ id: project_id, name, description }));
  }

  async getTotals(projectId) {
    const response = await this.#request.get(`/projects/${projectId}/totals`)
      .catch(() => ({
        totals: [
          { title: 'Total Expenses', amount: 1000000, currency: 'UAH' },
        ],
      }));

    const { totals = [] } = response;

    return totals.map(({ title, amount, currency }) => ({
      key: name.toLowerCase().replace(' ', '-'),
      title,
      amount: new Intl.NumberFormat().format(amount / 100),
      currency,
    }));
  }
}
