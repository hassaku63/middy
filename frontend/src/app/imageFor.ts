import type { IdeaCandidate } from '@/shared/api/client';
import type { ChatTurn } from '@/features/chat/useIdeas';

/**
 * The image an example answered: the user's message before the answer
 * that holds it, so an example opened from an earlier, folded turn
 * (BL-068) is saved with its own image, not the latest refinement.
 */
export function imageFor(turns: ChatTurn[], candidate: IdeaCandidate): string {
  const at = turns.findIndex((t) => t.role === 'assistant' && t.candidates.includes(candidate));
  const before = at < 0 ? turns : turns.slice(0, at);
  return [...before].reverse().find((t) => t.role === 'user')?.content ?? '';
}
