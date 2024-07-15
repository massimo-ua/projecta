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

    const { types, total } = response;
    return [types.map((type) => this.toType(type)), total];
  }

  async addType(projectId, { categoryId, name, description }) {
    const response = await this.#request.post(`/projects/${projectId}/types`, {
      category_id: categoryId,
      name,
      description,
    });

    if (!response.ok) {
      throw new Error('Failed to add type');
    }

    const json = await response.json();

    return this.toType(json);
  }

  async removeType(projectId, typeId) {
    const response = await this.#request.delete(`/projects/${projectId}/types/${typeId}`);

    if (!response.ok) {
      throw new Error('Failed to remove type');
    }
  }

  toType({ type_id, name, description, category }) {
    return {
      key: type_id,
      id: type_id,
      name,
      description,
      category: category.name,
    };
  }
}
