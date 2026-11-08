import { useCallback, useEffect, useRef, useState } from 'react';
import { ApiError, imagine, type ChatMessage, type IdeaCandidate } from '@/shared/api/client';

/** A rendered turn of the conversation: what was said and what came back. */
export type ChatTurn =
  | { role: 'user'; content: string }
  | { role: 'assistant'; content: string; mood: string[]; candidates: IdeaCandidate[] };

export type Ideas = {
  turns: ChatTurn[];
  busy: boolean;
  error: string | null;
  /** Sends the user's text with the whole history; resolves when the reply is in. */
  send: (text: string) => Promise<void>;
  reset: () => void;
};

/** What came back for one image: the text the user sent and the examples. */
export type Answer = { image: string; candidates: IdeaCandidate[] };

/**
 * Owns the conversation of scenario S2. The server is stateless, so the
 * full history (user texts and assistant replies) goes with every turn.
 * `onAnswer` runs once per reply, so the caller can open the first example
 * at once (docs/product/experience.md: an example arrives ready to hear).
 */
export function useIdeas(onAnswer?: (answer: Answer) => void): Ideas {
  const [turns, setTurns] = useState<ChatTurn[]>([]);
  const onAnswerRef = useRef(onAnswer);
  useEffect(() => {
    onAnswerRef.current = onAnswer;
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const send = useCallback(
    async (text: string) => {
      const content = text.trim();
      if (!content || busy) return;
      const history = recentHistory([
        ...turns.map((t) => ({ role: t.role, content: t.content })),
        { role: 'user', content },
      ]);
      setTurns((prev) => [...prev, { role: 'user', content }]);
      setBusy(true);
      setError(null);
      try {
        const res = await imagine(history);
        setTurns((prev) => [
          ...prev,
          { role: 'assistant', content: res.reply, mood: res.mood, candidates: res.candidates },
        ]);
        onAnswerRef.current?.({ image: content, candidates: res.candidates });
      } catch (e) {
        setError(describe(e));
      } finally {
        setBusy(false);
      }
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
      default:
        return e.message;
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
