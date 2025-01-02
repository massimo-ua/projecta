const CORRELATION_ID_HEADER = 'X-Request-ID';
const NO_CONTENT = 204;

export class Request {
  #baseUrl;

  #authProvider;

  #abortControllers;

  constructor(baseUrl, authProvider) {
    this.#baseUrl = baseUrl;
    this.#authProvider = authProvider;
    this.#abortControllers = new Map();
    window.addEventListener('unload', () => {
      this.#abortControllers.forEach(controller => controller.abort());
    });
  }

  async #buildRequestOptions(options = {}) {
    const {
      method = 'GET',
      body = null,
      ...rest
    } = options;

    const { headers = {}} = rest;
    const { [CORRELATION_ID_HEADER]: requestId = crypto.randomUUID(), ...restOfHeaders } = headers;

    const token = await this.#authProvider.getToken();
    const controller = new AbortController();

    const requestOptions = {
      ...rest,
      method,
      headers: {
        ...restOfHeaders,
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        [CORRELATION_ID_HEADER]: requestId,
      },
      ...(body && { body: JSON.stringify(body) }),
      signal: controller.signal,
    };

    return { requestOptions, controller };
  }

  #abortOngoingRequest(options = {}) {
    const requestId = options?.headers?.[CORRELATION_ID_HEADER];
    if (requestId) {
      const controller = this.#abortControllers.get(requestId);
      if (controller) {
        controller.abort();
      }
    }
  }

  async #call(url, options) {
    this.#abortOngoingRequest(options);
    const { requestOptions, controller } = await this.#buildRequestOptions(options);
    const requestId = requestOptions.headers[CORRELATION_ID_HEADER];
    this.#abortControllers.set(requestId, controller);

    try {
      const response = await fetch(`${this.#baseUrl}${url}`, requestOptions);

      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new Error(`${requestOptions.method} failed: ${response.status} ${error.message || response.statusText}`);
      }

      if (response.status !== NO_CONTENT) {
        return response.json().catch(() => ({}));
      }
    } finally {
      this.#abortControllers.delete(requestId);
    }
  }

  async get(url, options) {
    return await this.#call(url, options);
  }

  async post(url, data, options = {}) {
    return await this.#call(url, {
      ...options,
      method: 'POST',
      body: data,
    });
  }

  async put(url, data, options = {}) {
    return await this.#call(url, {
      ...options,
      method: 'PUT',
      body: data,
    });
  }

  async delete(url, options = {}) {
    return await this.#call(url, {
      ...options,
      method: 'DELETE',
    });
  }

  cancelRequest(requestId) {
    if (this.#abortControllers.has(requestId)) {
      this.#abortControllers.get(requestId).abort();
      this.#abortControllers.delete(requestId);
    }
  }
}
