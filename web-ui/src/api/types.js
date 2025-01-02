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

    return this.toType(response);
  }

  async removeType(projectId, typeId) {
    return await this.#request.delete(`/projects/${projectId}/types/${typeId}`);
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
