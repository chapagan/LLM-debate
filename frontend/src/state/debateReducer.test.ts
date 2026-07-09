import { describe, expect, it } from 'vitest';
import type { DebateState } from '../types';
import { initialDebateState, debateReducer } from './debateReducer';

describe('debateReducer', () => {
  it('starts a new session and clears previous turns', () => {
    const state = debateReducer(
      { ...initialDebateState, turns: [{ turn: 99, agent: 'cfo', agentName: 'Old', content: 'old', streaming: false }] },
      { type: 'SESSION_STARTED', sessionId: 's1', topic: 'Rust rewrite', totalTurns: 4 },
    );
    expect(state.sessionId).toBe('s1');
    expect(state.topic).toBe('Rust rewrite');
    expect(state.turns).toEqual([]);
    expect(state.status).toBe('running');
  });

  it('appends tokens to the matching active turn', () => {
    const started = debateReducer(initialDebateState, { type: 'SESSION_STARTED', sessionId: 's1', topic: 'Rust rewrite', totalTurns: 4 });
    const turn = debateReducer(started, { type: 'TURN_STARTED', sessionId: 's1', turn: 1, agent: 'visionary', agentName: 'The Disruptive Visionary' });
    const token = debateReducer(turn, { type: 'TOKEN', sessionId: 's1', turn: 1, agent: 'visionary', content: 'Hello' });
    expect(token.turns[0].content).toBe('Hello');
  });

  it('ignores stale events from canceled sessions', () => {
    const state = debateReducer(initialDebateState, { type: 'SESSION_STARTED', sessionId: 'new', topic: 'New', totalTurns: 4 });
    const stale = debateReducer(state, { type: 'TOKEN', sessionId: 'old', turn: 1, agent: 'visionary', content: 'stale' });
    expect(stale).toBe(state);
  });

  it('marks the matching turn done when a turn ends', () => {
    const state: DebateState = {
      ...initialDebateState,
      sessionId: 's1',
      turns: [
        { turn: 1, agent: 'visionary', agentName: 'The Disruptive Visionary', content: 'Hello', streaming: true },
        { turn: 1, agent: 'cfo', agentName: 'The Skeptical CFO', content: 'No', streaming: true },
      ],
    };

    const ended = debateReducer(state, { type: 'TURN_ENDED', sessionId: 's1', turn: 1, agent: 'visionary' });

    expect(ended.turns).toEqual([
      { turn: 1, agent: 'visionary', agentName: 'The Disruptive Visionary', content: 'Hello', streaming: false },
      { turn: 1, agent: 'cfo', agentName: 'The Skeptical CFO', content: 'No', streaming: true },
    ]);
  });

  it('marks the debate complete and clears streaming turns when debate ends', () => {
    const state: DebateState = {
      ...initialDebateState,
      sessionId: 's1',
      status: 'running' as const,
      turns: [
        { turn: 1, agent: 'visionary', agentName: 'The Disruptive Visionary', content: 'Hello', streaming: true },
        { turn: 1, agent: 'cfo', agentName: 'The Skeptical CFO', content: 'No', streaming: true },
      ],
    };

    const ended = debateReducer(state, { type: 'DEBATE_ENDED', sessionId: 's1' });

    expect(ended.status).toBe('complete');
    expect(ended.turns.map((turn) => turn.streaming)).toEqual([false, false]);
  });

  it('records errors and clears streaming turns', () => {
    const state: DebateState = {
      ...initialDebateState,
      sessionId: 's1',
      status: 'running' as const,
      turns: [
        { turn: 1, agent: 'visionary', agentName: 'The Disruptive Visionary', content: 'Hello', streaming: true },
        { turn: 1, agent: 'cfo', agentName: 'The Skeptical CFO', content: 'No', streaming: true },
      ],
    };

    const errored = debateReducer(state, { type: 'ERROR', sessionId: 's1', message: 'Socket closed' });

    expect(errored.status).toBe('error');
    expect(errored.error).toBe('Socket closed');
    expect(errored.turns.map((turn) => turn.streaming)).toEqual([false, false]);
  });
});
