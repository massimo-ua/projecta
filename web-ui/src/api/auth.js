import { differenceInMinutes } from 'date-fns';

const ACCESS_TOKEN_KEY = 'access-token';
const REFRESH_TOKEN_KEY = 'refresh-token';
const TOKEN_EXPIRES_AT_KEY = 'access-token-expires-at';

export class Auth {
  #baseUrl;

  #tokenKey;

  #refreshTokenKey;

  #tokenExpiresAtKey;

  #pendingTokenRequest;

  constructor(baseURL) {
    this.#baseUrl = baseURL;
    this.#tokenKey = ACCESS_TOKEN_KEY;
    this.#refreshTokenKey = REFRESH_TOKEN_KEY;
    this.#tokenExpiresAtKey = TOKEN_EXPIRES_AT_KEY;
  }

  #isTokenExpired() {
    const expiresAt = parseInt(localStorage.getItem(this.#tokenExpiresAtKey), 10);
    return !expiresAt || differenceInMinutes(new Date(expiresAt), new Date()) < 2;
  }

  async getToken() {
    if (this.#isTokenExpired()) {
      if (this.#pendingTokenRequest) {
        return this.#pendingTokenRequest;
      }

      this.#pendingTokenRequest = this.#refreshToken();

      this.#pendingTokenRequest.finally(() => {
        this.#pendingTokenRequest = null;
      });

      return this.#pendingTokenRequest;
    }

    return localStorage.getItem(this.#tokenKey);
  }

  async #refreshToken() {
    const refreshToken = localStorage.getItem(this.#refreshTokenKey);
    const accessToken = localStorage.getItem(this.#tokenKey);
    const response = await fetch(`${this.#baseUrl}/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: refreshToken,
        access_token: accessToken,
      }),
    });

    if (!response.ok) {
      this.logout();
      throw new Error('Failed to refresh token');
    }

    const json = await response.json();
    return this.#handleAuthResponse(json);
  }

  #handleAuthResponse(response) {
    const {
      access_token: accessToken,
      refresh_token: refreshToken,
      expires_at: expiresAt,
    } = response;

    localStorage.setItem(this.#tokenKey, accessToken);
    localStorage.setItem(this.#refreshTokenKey, refreshToken);
    localStorage.setItem(this.#tokenExpiresAtKey, String(expiresAt * 1000));
    return accessToken;
  }

  async login(username, password) {
    if (this.#pendingTokenRequest) {
      return this.#pendingTokenRequest;
    }

    this.#pendingTokenRequest = fetch(`${this.#baseUrl}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: username,
        token: password,
        identity_provider: 'LOCAL',
      }),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error('Failed to login');
        }

        return response.json();
      })
      .then((json) => this.#handleAuthResponse(json))
      .finally(() => {
        this.#pendingTokenRequest = null;
      });

    return this.#pendingTokenRequest;
  }

  async loginSocial(token, provider) {
    if (this.#pendingTokenRequest) {
      return this.#pendingTokenRequest;
    }

    this.#pendingTokenRequest = fetch(`${this.#baseUrl}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        token,
        identity_provider: provider,
      }),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error('Failed to login');
        }

        return response.json();
      })
      .then((json) => this.#handleAuthResponse(json))
      .finally(() => {
        this.#pendingTokenRequest = null;
      });

    return this.#pendingTokenRequest;
  }

  logout() {
    localStorage.removeItem(this.#tokenKey);
    localStorage.removeItem(this.#refreshTokenKey);
    localStorage.removeItem(this.#tokenExpiresAtKey);
    this.#pendingTokenRequest = null;
  }

  isAuthenticated() {
    return !!localStorage.getItem(this.#tokenKey);
  }
}
