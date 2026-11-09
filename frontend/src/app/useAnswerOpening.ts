import { useRef } from 'react';
import { useIdeas, type Answer, type Ideas } from '@/features/chat/useIdeas';
import type { IdeaCandidate } from '@/shared/api/client';

/**
 * The conversation, with the rule for opening an answer (BL-064): its first
 * example opens at once, unless the user opened something (another
 * candidate, a saved example) while the answer was being prepared. Then the
 * answer only adds its cards, and what the user chose stays open.
 */
export function useAnswerOpening(open: (c: IdeaCandidate, image: string) => void) {
  const opens = useRef(0); // examples the user opened
  const sentAt = useRef(0); // opens when the last image was sent
  const chat = useIdeas((answer: Answer) => {
    const first = answer.candidates[0];
    if (first && opens.current === sentAt.current) open(first, answer.image);
  });
  const ideas: Ideas = {
    ...chat,
    send: (text) => {
      sentAt.current = opens.current;
      return chat.send(text);
    },
  };
  /** Call when the user opens an example themselves. */
  const opened = () => {
    opens.current += 1;
  };
  return { ideas, opened };
}
