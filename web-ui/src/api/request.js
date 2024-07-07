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

    return response.json();
  }
}
