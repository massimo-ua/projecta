export class CategoriesRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getCategories(projectId, limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const resourceUrl = `/projects/${projectId}/categories`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { categories } = response;
    return categories.map(({ category_id, name, description }) => ({
      key: category_id,
      id: category_id,
      name,
      description,
    }));
  }
}
