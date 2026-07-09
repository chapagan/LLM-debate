import { useCallback, useEffect, useReducer, useRef, useState } from 'react';
import { debateReducer, initialDebateState } from '../state/debateReducer';
import type { ConnectionStatus, InboundEvent, OutboundEvent } from '../types';

export function useDebateSocket() {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('connecting');
  const [state, dispatch] = useReducer(debateReducer, initialDebateState);
  const socketRef = useRef<WebSocket | null>(null);
  const debateStatusRef = useRef(state.status);

  useEffect(() => {
    debateStatusRef.current = state.status;
  }, [state.status]);

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const socket = new WebSocket(`${protocol}://${window.location.host}/ws`);
    socketRef.current = socket;
    socket.onopen = () => {
      if (socketRef.current !== socket) return;
      setConnectionStatus('open');
    };
    socket.onclose = () => {
      if (socketRef.current !== socket) return;
      setConnectionStatus('closed');
      if (debateStatusRef.current === 'running') {
        dispatch({ type: 'ERROR', message: 'WebSocket connection closed' });
      }
    };
    socket.onerror = () => {
      if (socketRef.current !== socket) return;
      setConnectionStatus('error');
      dispatch({ type: 'ERROR', message: 'WebSocket connection error' });
    };
    socket.onmessage = (message) => {
      if (socketRef.current !== socket) return;
      try {
        const event = JSON.parse(message.data) as unknown;
        if (isOutboundEvent(event)) {
          dispatch(event);
          return;
        }
        dispatch({ type: 'ERROR', message: 'Received an invalid server message' });
      } catch {
        dispatch({ type: 'ERROR', message: 'Received an invalid server message' });
      }
    };
    return () => {
      socket.onopen = null;
      socket.onclose = null;
      socket.onerror = null;
      socket.onmessage = null;
      socket.close();
      if (socketRef.current === socket) {
        socketRef.current = null;
      }
    };
  }, []);

  const startDebate = useCallback((topic: string) => {
    const event: InboundEvent = { action: 'START_DEBATE', topic };
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      dispatch({ type: 'ERROR', message: 'WebSocket is not connected' });
      return;
    }
    socket.send(JSON.stringify(event));
  }, []);

  return { connectionStatus, debate: state, startDebate };
}

function isOutboundEvent(value: unknown): value is OutboundEvent {
  if (!isRecord(value) || typeof value.type !== 'string') {
    return false;
  }

  switch (value.type) {
    case 'SESSION_STARTED':
      return isString(value.sessionId) && isString(value.topic) && isNumber(value.totalTurns);
    case 'TURN_STARTED':
      return isString(value.sessionId) && isNumber(value.turn) && isAgent(value.agent) && isString(value.agentName);
    case 'TOKEN':
      return isString(value.sessionId) && isNumber(value.turn) && isAgent(value.agent) && isString(value.content);
    case 'TURN_ENDED':
      return isString(value.sessionId) && isNumber(value.turn) && isAgent(value.agent);
    case 'DEBATE_ENDED':
      return isString(value.sessionId);
    case 'ERROR':
      return isString(value.message) && (value.sessionId === undefined || isString(value.sessionId));
    default:
      return false;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isString(value: unknown): value is string {
  return typeof value === 'string';
}

function isNumber(value: unknown): value is number {
  return typeof value === 'number';
}

function isAgent(value: unknown): value is 'visionary' | 'cfo' {
  return value === 'visionary' || value === 'cfo';
}
