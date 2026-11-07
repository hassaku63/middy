import { useState, type FormEvent } from 'react';
import type { IdeaCandidate } from '@/shared/api/client';
import { CandidateCard } from './CandidateCard';
import type { Ideas } from './useIdeas';
import './ChatPanel.css';

export type ChatPanelProps = {
  ideas: Ideas;
  /** Called when the user opens a candidate in the player/timeline/panel. */
  onOpenCandidate: (candidate: IdeaCandidate) => void;
  /** Title of the candidate currently open, to mark it. */
  openTitle?: string | null;
};

const SUGGESTIONS = ['雨の日曜の午後', '出発の朝、晴れ', '夜の帰り道、少し寂しい'];

/** Conversation view for scenario S2: messages, candidate cards, input. */
export function ChatPanel({ ideas, onOpenCandidate, openTitle = null }: ChatPanelProps) {
  const [text, setText] = useState('');

  const submit = (e: FormEvent) => {
    e.preventDefault();
    const t = text;
    setText('');
    void ideas.send(t);
  };

  return (
    <section className="chat" aria-label="イメージから">
      <ol className="chat__log" aria-label="会話">
        {ideas.turns.map((turn, i) =>
          turn.role === 'user' ? (
            <li key={i} className="chat__turn chat__turn--user">
              <p>{turn.content}</p>
            </li>
          ) : (
            <li key={i} className="chat__turn chat__turn--assistant">
              <p>{turn.content}</p>
              {turn.mood.length > 0 && (
                <ul className="chat__mood" aria-label="雰囲気">
                  {turn.mood.map((m) => (
                    <li key={m}>{m}</li>
                  ))}
                </ul>
              )}
              <ul className="chat__candidates" aria-label="候補">
                {turn.candidates.map((c) => (
                  <li key={c.title}>
                    <CandidateCard
                      candidate={c}
                      open={openTitle === c.title}
                      onOpen={() => onOpenCandidate(c)}
                    />
                  </li>
                ))}
              </ul>
            </li>
          ),
        )}
        {ideas.busy && (
          <li className="chat__turn chat__turn--assistant" aria-live="polite">
            <p>考え中…</p>
          </li>
        )}
      </ol>
      {ideas.error && (
        <p role="alert" className="chat__error">
          {ideas.error}
        </p>
      )}
      {ideas.turns.length === 0 && (
        <p className="chat__hint">
          曲のイメージを言葉で伝えると、合いそうな調・テンポ・コード進行を提案します。例:{' '}
          {SUGGESTIONS.map((s) => (
            <button key={s} type="button" className="chat__suggestion" onClick={() => setText(s)}>
              {s}
            </button>
          ))}
        </p>
      )}
      <form className="chat__form" onSubmit={submit} aria-label="イメージの入力">
        <input
          type="text"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="どんな曲にしたい？（例: 雨の日曜の午後）"
          aria-label="イメージ"
          maxLength={1000}
          disabled={ideas.busy}
        />
        <button type="submit" disabled={ideas.busy || text.trim() === ''}>
          送る
        </button>
        {ideas.turns.length > 0 && (
          <button type="button" onClick={ideas.reset} disabled={ideas.busy}>
            最初から
          </button>
        )}
      </form>
    </section>
  );
}
