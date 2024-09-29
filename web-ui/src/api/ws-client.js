export default class WsClient {
  constructor(url) {
    this.url = url;
    this.ws = new WebSocket(url);
  }

  onOpen(callback) {
    this.ws.onopen = callback;
  }

  onClose(callback) {
    this.ws.onclose = callback;
  }

  onError(callback) {
    this.ws.onerror = callback;
  }

  onMessage(callback) {
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      callback(data);
    };
  }

  send(data) {
    this.ws.send(JSON.stringify(data));
  }
}
