import type { IdeaCandidate } from '@/shared/api/client';
import type { ChatTurn } from './useIdeas';

/** One exchange: the image the user sent and the examples that came back. */
export type Exchange = { number: number; image: string; candidates: IdeaCandidate[] };

/** The conversation as exchanges, numbered from 1 in the order they were sent. */
export function exchangesOf(turns: ChatTurn[]): Exchange[] {
  const out: Exchange[] = [];
  for (const t of turns) {
    if (t.role === 'user') out.push({ number: out.length + 1, image: t.content, candidates: [] });
    else out.at(-1)?.candidates.push(...t.candidates);
  }
  return out;
}
