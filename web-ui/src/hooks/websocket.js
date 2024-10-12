import { useEffect, useState } from 'react';
import WsClient, { MessageType } from '../api/ws-client';
import { authProvider } from '../api';

const WS_URL = '/ws';
const PING_INTERVAL = 5000;

// TODO: review this hook not sure if it should be a hook
function UseWebsocket(auth, url) {
  return () => {
    const [wsClient, setWsClient] = useState(null);
    const [destroyListener, setDestroyListener] = useState(null);
    const [authToken, setAuthToken] = useState('');
    const [pingTimer, setPingTimer] = useState(null);

    const disconnect = () => {
      if (destroyListener) destroyListener();
      if (pingTimer) clearInterval(pingTimer);
      if (wsClient) setWsClient(null);
    };

    const updateWsClient = (token) => {
      if (!token) {
        return wsClient && disconnect();
      }

      if (wsClient) {
        return wsClient.setAuthToken(token);
      }

      const ws = new WsClient(url, token);

      ws.onOpen(() => {
        const pingTimer = setInterval(() => {
          ws.send(MessageType.PING);
        }, PING_INTERVAL);

        setPingTimer(pingTimer);
      });

      ws.onClose(disconnect);

      setWsClient(ws);
    };

    useEffect(() => {
      const removeIdentityListener = auth.onIdentityChange(
        ({ detail: authToken }) => setAuthToken(authToken));
      setDestroyListener(removeIdentityListener);

      return disconnect;
    }, []);

    useEffect(() => {
      if (authToken) {
        updateWsClient(authToken);
      }
    }, [authToken]);

    return wsClient;
  }
}

export default UseWebsocket(authProvider, WS_URL);
