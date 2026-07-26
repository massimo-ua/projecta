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
    return projects.map((item) => ({
      id: item.project_id,
      name: item.name,
      description: item.description,
      mainCurrency: item.main_currency || 'UAH',
    }));
  }

  async getProject(projectId) {
    const response = await this.#request.get(`/projects/${projectId}`);
    return {
      id: response.project_id,
      name: response.name,
      description: response.description,
      mainCurrency: response.main_currency || 'UAH',
    };
  }

  async createProject({ name, description }) {
    const response = await this.#request.post('/projects', { name, description });
    return {
      id: response.project_id,
      name: response.name,
      description: response.description,
      mainCurrency: response.main_currency || 'UAH',
    };
  }

  async updateProjectSettings(projectId, { mainCurrency }) {
    const response = await this.#request.patch(`/projects/${projectId}`, {
      main_currency: mainCurrency,
    });
    return {
      id: response.project_id,
      name: response.name,
      description: response.description,
      mainCurrency: response.main_currency || 'UAH',
    };
  }

  async getTotals(projectId) {
    const response = await this.#request.get(`/projects/${projectId}/totals`)
      .catch(() => ({
        totals: [
          { title: 'Total Expenses', amount: 0, currency: 'UAH' },
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
