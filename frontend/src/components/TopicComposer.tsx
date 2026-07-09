import { SendHorizonal } from 'lucide-react';
import { FormEvent, useState } from 'react';

type Props = {
  disabled: boolean;
  running: boolean;
  onStart: (topic: string) => void;
};

export function TopicComposer({ disabled, running, onStart }: Props) {
  const [topic, setTopic] = useState('Should we rewrite our legacy core monolith into microservices using Rust?');

  function submit(event: FormEvent) {
    event.preventDefault();
    const trimmed = topic.trim();
    if (trimmed) {
      onStart(trimmed);
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-3 border-b border-slate-200 bg-white px-5 py-4 lg:flex-row">
      <input
        value={topic}
        onChange={(event) => setTopic(event.target.value)}
        className="min-h-11 flex-1 rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-cyan-600 focus:ring-2 focus:ring-cyan-100"
        aria-label="Debate topic"
      />
      <button
        disabled={disabled || topic.trim().length === 0}
        className="inline-flex min-h-11 items-center justify-center gap-2 rounded-md bg-cyan-700 px-4 text-sm font-semibold text-white hover:bg-cyan-800 disabled:cursor-not-allowed disabled:bg-slate-300"
      >
        <SendHorizonal size={17} />
        {running ? 'Restart Debate' : 'Start Debate'}
      </button>
    </form>
  );
}
