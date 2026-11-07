import { useCallback, useState } from 'react';
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

/**
 * Owns the conversation of scenario S2. The server is stateless, so the
 * full history (user texts and assistant replies) goes with every turn.
 */
export function useIdeas(): Ideas {
  const [turns, setTurns] = useState<ChatTurn[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const send = useCallback(
    async (text: string) => {
      const content = text.trim();
      if (!content || busy) return;
      const history: ChatMessage[] = [
        ...turns.map((t) => ({ role: t.role, content: t.content })),
        { role: 'user', content },
      ];
      setTurns((prev) => [...prev, { role: 'user', content }]);
      setBusy(true);
      setError(null);
      try {
        const res = await imagine(history);
        setTurns((prev) => [
          ...prev,
          { role: 'assistant', content: res.reply, mood: res.mood, candidates: res.candidates },
        ]);
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
      default:
        return e.message;
    }
  }
  return 'バックエンドに接続できません';
}
