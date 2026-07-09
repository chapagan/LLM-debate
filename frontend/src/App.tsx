import { DebateArena } from './components/DebateArena';
import { StatusBar } from './components/StatusBar';
import { TopicComposer } from './components/TopicComposer';
import { useDebateSocket } from './hooks/useDebateSocket';

export default function App() {
  const { connectionStatus, debate, startDebate } = useDebateSocket();
  const connected = connectionStatus === 'open';
  return (
    <div className="flex h-screen min-h-[720px] flex-col bg-slate-100 text-slate-950">
      <header className="border-b border-slate-200 bg-slate-950 px-5 py-4 text-white">
        <h1 className="text-xl font-semibold">LLM Debate</h1>
      </header>
      <TopicComposer disabled={!connected} running={debate.status === 'running'} onStart={startDebate} />
      <StatusBar connectionStatus={connectionStatus} debate={debate} />
      {debate.error ? (
        <div role="alert" className="border-b border-red-200 bg-red-50 px-5 py-3 text-sm text-red-700">
          {debate.error}
        </div>
      ) : null}
      <DebateArena debate={debate} />
    </div>
  );
}
