import type { DebateState, OutboundEvent, TurnState } from '../types';

export const initialDebateState: DebateState = {
  sessionId: null,
  topic: '',
  totalTurns: 4,
  status: 'idle',
  error: null,
  turns: [],
};

export function debateReducer(state: DebateState, event: OutboundEvent): DebateState {
  if ('sessionId' in event && event.sessionId && state.sessionId && event.sessionId !== state.sessionId && event.type !== 'SESSION_STARTED') {
    return state;
  }
  switch (event.type) {
    case 'SESSION_STARTED':
      return {
        sessionId: event.sessionId,
        topic: event.topic,
        totalTurns: event.totalTurns,
        status: 'running',
        error: null,
        turns: [],
      };
    case 'TURN_STARTED':
      return {
        ...state,
        turns: [
          ...state.turns,
          { turn: event.turn, agent: event.agent, agentName: event.agentName, content: '', streaming: true },
        ],
      };
    case 'TOKEN':
      return {
        ...state,
        turns: state.turns.map((turn) =>
          turn.turn === event.turn && turn.agent === event.agent
            ? { ...turn, content: turn.content + event.content }
            : turn,
        ),
      };
    case 'TURN_ENDED':
      return { ...state, turns: markTurnDone(state.turns, event.turn, event.agent) };
    case 'DEBATE_ENDED':
      return { ...state, status: 'complete', turns: state.turns.map((turn) => ({ ...turn, streaming: false })) };
    case 'ERROR':
      return { ...state, status: 'error', error: event.message, turns: state.turns.map((turn) => ({ ...turn, streaming: false })) };
    default:
      return state;
  }
}

function markTurnDone(turns: TurnState[], turnNumber: number, agent: TurnState['agent']): TurnState[] {
  return turns.map((turn) => (turn.turn === turnNumber && turn.agent === agent ? { ...turn, streaming: false } : turn));
}
