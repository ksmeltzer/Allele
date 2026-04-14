import React, { createContext, useContext, useEffect, useState, useRef } from 'react';
import type { ReactNode } from 'react';

export type WSMessage = {
  type: string;
  ts: number;
  payload: any;
};

type WebSocketContextType = {
  connected: boolean;
  sendEvent: (type: string, payload?: any) => void;
  subscribe: (type: string, callback: (payload: any) => void) => () => void;
};

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const WebSocketProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const subscribersRef = useRef<Map<string, Set<(payload: any) => void>>>(new Map());

  useEffect(() => {
    let timeoutId: number;

    const connectWS = () => {
      // In dev mode or production, use standard token (hardcoded fallback for now)
      const token = 'dev-token';
      const wsUrl = `ws://${window.location.hostname}:8081/ws?auth_token=${token}`;
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        setConnected(true);
        wsRef.current = ws;
      };

      ws.onclose = () => {
        setConnected(false);
        wsRef.current = null;
        // Exponential backoff or simple reconnect
        timeoutId = window.setTimeout(connectWS, 3000);
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data) as WSMessage;
          const handlers = subscribersRef.current.get(msg.type);
          if (handlers) {
            handlers.forEach(fn => fn(msg.payload));
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      };
    };

    connectWS();

    return () => {
      window.clearTimeout(timeoutId);
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const sendEvent = (type: string, payload?: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload: payload || {} }));
    } else {
      console.warn(`WebSocket not connected. Could not send event: ${type}`);
    }
  };

  const subscribe = (type: string, callback: (payload: any) => void) => {
    if (!subscribersRef.current.has(type)) {
      subscribersRef.current.set(type, new Set());
    }
    const handlers = subscribersRef.current.get(type)!;
    handlers.add(callback);

    return () => {
      handlers.delete(callback);
    };
  };

  return (
    <WebSocketContext.Provider value={{ connected, sendEvent, subscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  const ctx = useContext(WebSocketContext);
  if (!ctx) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return ctx;
};
