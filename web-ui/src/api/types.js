export class TypesRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getTypes(projectId, limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const resourceUrl = `/projects/${projectId}/types`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { types } = response;
    return types.map(({ type_id, name, description }) => ({
      key: type_id, id: type_id, name, description,
    }));
  }

  async addType(projectId, { name, description }) {
    const response = await this.#request.post(`/projects/${projectId}/types`, {
      name,
      description,
    });

    if (!response.ok) {
      throw new Error('Failed to add type');
    }

    const json = await response.json();

    return {
      key: json.type_id, id: json.type_id, name: json.name, description: json.description,
    };
  }
}
