import type { ConnectionStatus, DebateState } from '../types';

type Props = {
  connectionStatus: ConnectionStatus;
  debate: DebateState;
};

export function StatusBar({ connectionStatus, debate }: Props) {
  const completedTurns = debate.turns.filter((turn) => !turn.streaming).length;
  return (
    <div className="grid gap-3 border-b border-slate-200 bg-slate-50 px-5 py-3 text-sm text-slate-700 sm:grid-cols-3">
      <div>
        Connection: <strong className="capitalize">{connectionStatus}</strong>
      </div>
      <div>
        Session: <strong>{debate.sessionId ? debate.sessionId.slice(0, 8) : 'none'}</strong>
      </div>
      <div>
        Turns: <strong>{completedTurns}/{debate.totalTurns}</strong>
      </div>
    </div>
  );
}
