import type { AgentId, DebateState, TurnState } from '../types';

const laneCopy: Record<AgentId, { title: string; subtitle: string }> = {
  visionary: {
    title: 'The Disruptive Visionary',
    subtitle: 'Strategic upside, speed, reinvention',
  },
  cfo: {
    title: 'The Skeptical CFO',
    subtitle: 'Cost control, migration risk, measurable ROI',
  },
};

type Props = {
  debate: DebateState;
};

export function DebateArena({ debate }: Props) {
  return (
    <main className="grid min-h-0 flex-1 gap-4 bg-slate-100 p-4 lg:grid-cols-2">
      <Lane agent="visionary" turns={debate.turns.filter((turn) => turn.agent === 'visionary')} />
      <Lane agent="cfo" turns={debate.turns.filter((turn) => turn.agent === 'cfo')} />
    </main>
  );
}

function Lane({ agent, turns }: { agent: AgentId; turns: TurnState[] }) {
  const copy = laneCopy[agent];
  return (
    <section className="flex min-h-[420px] flex-col rounded-lg border border-slate-200 bg-white">
      <header className="border-b border-slate-200 px-5 py-4">
        <h2 className="text-base font-semibold text-slate-950">{copy.title}</h2>
        <p className="mt-1 text-sm text-slate-500">{copy.subtitle}</p>
      </header>
      <div className="flex-1 space-y-3 overflow-auto p-4">
        {turns.length === 0 ? (
          <div className="flex h-full min-h-52 items-center justify-center rounded-md border border-dashed border-slate-300 text-sm text-slate-500">
            Waiting for this agent&apos;s turn
          </div>
        ) : (
          turns.map((turn) => <TurnCard key={`${turn.agent}-${turn.turn}`} turn={turn} />)
        )}
      </div>
    </section>
  );
}

function TurnCard({ turn }: { turn: TurnState }) {
  return (
    <article className="rounded-md border border-slate-200 bg-slate-50 p-4">
      <div className="mb-2 flex items-center justify-between text-xs font-semibold uppercase tracking-wide text-slate-500">
        <span>Turn {turn.turn}</span>
        <span>{turn.streaming ? 'Streaming' : 'Complete'}</span>
      </div>
      <p className="whitespace-pre-wrap text-sm leading-6 text-slate-800">{turn.content || '...'}</p>
    </article>
  );
}
