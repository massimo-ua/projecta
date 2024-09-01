export class Request {
  #baseUrl;

  #authProvider;

  constructor(baseUrl, authProvider) {
    this.#baseUrl = baseUrl;
    this.#authProvider = authProvider;
  }

  async get(url, options) {
    const token = await this.#authProvider.getToken();
    const response = await fetch(`${this.#baseUrl}${url}`, {
      ...options,
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to get data');
    }

    return response.json();
  }

  async #withBody(method, url, data, options) {
    const token = await this.#authProvider.getToken();
    return await fetch(`${this.#baseUrl}${url}`, {
      ...options,
      method,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });
  }

  async post(url, data, options) {
    return await this.#withBody('POST', url, data, options);
  }

  async put(url, data, options) {
    return await this.#withBody('PUT', url, data, options);
  }

  async delete(url, options) {
    const token = await this.#authProvider.getToken();
    return await fetch(`${this.#baseUrl}${url}`, {
      ...options,
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    });
  }
}
