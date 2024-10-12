export const MessageType = {
  PING: 'ping',
  PONG: 'pong',
};

export default class WsClient {
  #authToken;
  #url;
  #ws

  constructor(url, authToken = '') {
    this.#url = url;
    this.#authToken = authToken;

    if (this.#authToken) {
      this.open();
    }
  }

  setAuthToken(authToken) {
    this.#authToken = authToken;
  }

  onOpen(callback) {
    if (!this.#ws) {
      throw Error('Websocket client is not connected');
    }

    this.#ws.onopen = callback;
  }

  onClose(callback) {
    this.#ws.onclose = callback;
  }

  onError(callback) {
    if (!this.#ws) {
      throw Error('Websocket client is not connected');
    }

    this.#ws.onerror = callback;
  }

  onMessage(callback) {
    if (!this.#ws) {
      throw Error('Websocket client is not connected');
    }

    this.#ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      callback(data);
    };
  }

  send(type, data = null) {
    this.#ws.send(JSON.stringify({
      type,
      token: this.#authToken,
      data,
    }));
  }

  close() {
    if (!this.#ws) {
      return;
    }

    this.#ws.close();
    this.#ws = null;
  }

  open() {
    if (this.#ws) {
      return;
    }

    if (!this.#authToken) {
      throw Error('Websocket client is not authenticated');
    }

    this.#ws = new WebSocket(`${this.#url}?token=${this.#authToken}`);
  }
}
