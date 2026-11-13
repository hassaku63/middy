import { useState, type ReactNode, type SubmitEvent } from 'react';
import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { CandidateCard } from './CandidateCard';
import { ExampleList } from './ExampleList';
import { exchangesOf } from './exchanges';
import type { ChatTurn, Ideas } from './useIdeas';
import './ChatPanel.css';

export type ChatPanelProps = {
  ideas: Ideas;
  /** Called when the user opens a candidate in the player/timeline/panel. */
  onOpenCandidate: (candidate: IdeaCandidate) => void;
  /** The example currently open, to mark its card (compared by identity). */
  openExample?: IdeaCandidate | null;
  /** Catalogue for showing pattern names on cards. */
  patterns?: PatternsResponse | undefined;
  /**
   * Earlier exchanges as a list of one line per example instead of folded
   * away (BL-186). The app always asks for the list.
   */
  list?: boolean;
  /**
   * Settings for choosing an example, shown below the input, where the
   * examples are chosen (the switch for playing on choosing, BL-217).
   */
  options?: ReactNode;
};

const SUGGESTIONS = ['雨の日曜の午後', '出発の朝、晴れ', '夜の帰り道、少し寂しい'];

/** Conversation view for scenario S2: messages, candidate cards, input. */
export function ChatPanel({
  ideas,
  onOpenCandidate,
  openExample = null,
  patterns,
  list = false,
  options,
}: ChatPanelProps) {
  const [text, setText] = useState('');
  const heard = useReplyNotice(ideas);

  // Only the latest exchange stays open; earlier ones fold away so refining
  // never pushes playback out of sight (BL-068).
  const lastUser = ideas.turns.map((t) => t.role).lastIndexOf('user');
  const split = lastUser < 0 ? 0 : lastUser;
  const earlier = ideas.turns.slice(0, split);
  const latest = ideas.turns.slice(split);
  const earlierCount = earlier.filter((t) => t.role === 'user').length;

  const renderTurn = (turn: ChatTurn, i: number) =>
    turn.role === 'user' ? (
      <li key={i} className="chat__turn chat__turn--user">
        <p>{turn.content}</p>
      </li>
    ) : (
      <li key={i} className="chat__turn chat__turn--assistant">
        <ReplyHead turn={turn} />
        <ul className="chat__candidates" aria-label="候補">
          {turn.candidates.map((c, n) => (
            <li key={`${n}-${c.title}`}>
              <CandidateCard
                candidate={c}
                open={openExample === c}
                onOpen={() => onOpenCandidate(c)}
                patterns={patterns}
              />
            </li>
          ))}
        </ul>
      </li>
    );

  const submit = (e: SubmitEvent) => {
    e.preventDefault();
    if (ideas.busy) return; // Enter in the read-only field while waiting
    const t = text;
    setText('');
    void ideas.send(t);
  };

  return (
    <section className="chat" aria-label="イメージから">
      {/* What scrolls when the conversation has a column of its own (BL-184). */}
      <div className="chat__scroll">
        {list && earlier.length > 0 && (
          <ExampleList
            exchanges={exchangesOf(earlier)}
            openExample={openExample}
            onOpen={onOpenCandidate}
          />
        )}
        {!list && earlier.length > 0 && (
          <details className="chat__earlier">
            <summary>これまでのやり取り（{earlierCount}）</summary>
            <ol className="chat__log" aria-label="これまでの会話">
              {earlier.map(renderTurn)}
            </ol>
          </details>
        )}
        <ol className="chat__log" aria-label="会話">
          {latest.map((turn, i) => renderTurn(turn, earlier.length + i))}
          {ideas.busy && (
            <li className="chat__turn chat__turn--assistant">
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
      </div>
      <div className="chat__composer">
        <form className="chat__form" onSubmit={submit} aria-label="イメージの入力">
          <input
            type="text"
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="どんな曲にしたい？（例: 雨の日曜の午後）"
            aria-label="イメージ"
            maxLength={1000}
            readOnly={ideas.busy}
            aria-busy={ideas.busy}
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
        {options && <div className="chat__options">{options}</div>}
      </div>
      <span className="visually-hidden" role="status">
        {heard}
      </span>
    </section>
  );
}

type AssistantTurn = Extract<ChatTurn, { role: 'assistant' }>;

/**
 * What a screen reader hears of a send (#500): 「考え中」 when it goes out
 * and the number of examples when the reply comes. A failure is the alert's
 * to say. Nothing is said for a conversation already on the screen.
 */
function useReplyNotice(ideas: Ideas): string {
  const [wasBusy, setWasBusy] = useState(ideas.busy);
  const [said, setSaid] = useState('');
  if (wasBusy !== ideas.busy) {
    setWasBusy(ideas.busy);
    const last = ideas.turns.at(-1);
    const n = last?.role === 'assistant' ? last.candidates.length : null;
    setSaid(
      ideas.busy
        ? '考え中'
        : n === null
          ? ''
          : n > 0
            ? `例が ${n} 件届きました`
            : '例は届きませんでした',
    );
  }
  return said;
}

/** The reply's summary, what it could not express (ADR-0061), and its mood tags. */
function ReplyHead({ turn }: { turn: AssistantTurn }) {
  return (
    <>
      <p>{turn.content}</p>
      {turn.unconstrained && (
        <p className="chat__unconstrained">
          この例では表せなかったもの（AI の見立て）: {turn.unconstrained}
        </p>
      )}
      {turn.mood.length > 0 && (
        <ul className="chat__mood" aria-label="雰囲気">
          {turn.mood.map((m) => (
            <li key={m}>{m}</li>
          ))}
        </ul>
      )}
    </>
  );
}
