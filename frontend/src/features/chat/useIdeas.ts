import { useCallback, useState } from 'react';
import { ApiError, imagine, type ChatMessage, type IdeaCandidate } from '@/shared/api/client';
import { historyMessage } from './history';

/** A rendered turn of the conversation: what was said and what came back. */
export type ChatTurn =
  | { role: 'user'; content: string }
  | { role: 'assistant'; content: string; mood: string[]; candidates: IdeaCandidate[] };

export type Ideas = {
  turns: ChatTurn[];
  busy: boolean;
  error: string | null;
  /** Sends the user's text with the recent history (recentHistory); resolves when the reply is in. */
  send: (text: string) => Promise<void>;
  reset: () => void;
};

/**
 * Owns the conversation of scenario S2. The server is stateless, so the
 * history (user texts, and assistant replies with the gist of their
 * examples, historyMessage) goes with every turn, cut to
 * the latest MAX_MESSAGES by recentHistory. An answer opens nothing by
 * itself: the user chooses (docs/product/experience.md, BL-219).
 */
export function useIdeas(): Ideas {
  const [turns, setTurns] = useState<ChatTurn[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const send = useCallback(
    async (text: string) => {
      const content = text.trim();
      if (!content || busy) return;
      const history = recentHistory([...turns.map(historyMessage), { role: 'user', content }]);
      setTurns((prev) => [...prev, { role: 'user', content }]);
      setBusy(true);
      setError(null);
      let res: Awaited<ReturnType<typeof imagine>>;
      try {
        res = await imagine(history);
      } catch (e) {
        setError(describe(e));
        return;
      } finally {
        setBusy(false);
      }
      setTurns((prev) => [
        ...prev,
        { role: 'assistant', content: res.reply, mood: res.mood, candidates: res.candidates },
      ]);
    },
    [turns, busy],
  );

  const reset = useCallback(() => {
    setTurns([]);
    setError(null);
  }, []);

  return { turns, busy, error, send, reset };
}

function describe(e: unknown): string {
  if (e instanceof ApiError) {
    switch (e.code) {
      case 'llm_unavailable':
        return 'AI に接続できませんでした。少し待ってからもう一度送ってください。';
      case 'llm_invalid_output':
        return 'AI の答えをうまく読めませんでした。言い方を変えてもう一度送ってください。';
      case 'llm_declined':
        return 'AI がこの言い方には答えられませんでした。情景や気分を別の言葉で伝えてみてください。';
      case 'internal_error':
        return 'AI の答えから例を作れませんでした。もう一度送ってください。';
      default:
        // The API's own messages are English and meant for developers, so
        // the learner gets a sentence in the UI's language.
        return 'うまく送れませんでした。もう一度送ってください。';
    }
  }
  return 'バックエンドに接続できません';
}

/** The most messages a turn may send (the API's `messages.maxItems`). */
export const MAX_MESSAGES = 20;

/**
 * The latest messages that fit the API, starting from a user message: a
 * long conversation keeps going on its recent turns instead of failing
 * once it passes the limit.
 */
export function recentHistory(messages: ChatMessage[]): ChatMessage[] {
  let out = messages.slice(-MAX_MESSAGES);
  while (out.length > 1 && out[0]?.role !== 'user') out = out.slice(1);
  return out;
}
