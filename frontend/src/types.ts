export type AgentId = 'visionary' | 'cfo';

export type InboundEvent = {
  action: 'START_DEBATE';
  topic: string;
};

export type OutboundEvent =
  | { type: 'SESSION_STARTED'; sessionId: string; topic: string; totalTurns: number }
  | { type: 'TURN_STARTED'; sessionId: string; turn: number; agent: AgentId; agentName: string }
  | { type: 'TOKEN'; sessionId: string; turn: number; agent: AgentId; content: string }
  | { type: 'TURN_ENDED'; sessionId: string; turn: number; agent: AgentId }
  | { type: 'DEBATE_ENDED'; sessionId: string }
  | { type: 'ERROR'; sessionId?: string; message: string };

export type ConnectionStatus = 'connecting' | 'open' | 'closed' | 'error';

export type TurnState = {
  turn: number;
  agent: AgentId;
  agentName: string;
  content: string;
  streaming: boolean;
};

export type DebateState = {
  sessionId: string | null;
  topic: string;
  totalTurns: number;
  status: 'idle' | 'running' | 'complete' | 'error';
  error: string | null;
  turns: TurnState[];
};
