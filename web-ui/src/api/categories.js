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

    const { categories, total } = response;
    return [categories.map(({ category_id, name, description }) => ({
      key: category_id,
      id: category_id,
      name,
      description,
    })), total];
  }

  async addCategory(projectId, categoryData) {
    const resourceUrl = `/projects/${projectId}/categories`;
    const response = await this.#request.post(resourceUrl, {
      name: categoryData.name,
      description: categoryData.description
    });

    const { category_id, name, description } = response;
    return {
      key: category_id,
      id: category_id,
      name,
      description,
    };
  }

  async removeCategory(projectId, categoryId) {
    const resourceUrl = `/projects/${projectId}/categories/${categoryId}`;
    await this.#request.delete(resourceUrl);
  }
}
