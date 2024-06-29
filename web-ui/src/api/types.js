export class TypesRepository {
  #request;
  constructor(request) {
    this.#request = request;
  }

  async getTypes(projectId, limit = 10, offset = 0 ) {
    const query = new URLSearchParams({limit: String(limit), offset: String(offset)}).toString();
    const resourceUrl = `/projects/${projectId}/types`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { types } = response;
    return types.map(({ type_id, name, description }) => ({ key: type_id, id: type_id, name, description }));
  }
}
