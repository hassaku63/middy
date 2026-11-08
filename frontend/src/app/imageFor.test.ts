import { describe, expect, it } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import type { ChatTurn } from '@/features/chat/useIdeas';
import { imageFor } from './imageFor';

const candidate = (title: string) => ({ title }) as IdeaCandidate;

describe('imageFor (review of BL-068)', () => {
  it('names the message an earlier answer replied to, not the latest one', () => {
    const first = candidate('雨の例');
    const second = candidate('明るい例');
    const turns: ChatTurn[] = [
      { role: 'user', content: '雨の日曜' },
      { role: 'assistant', content: '', mood: [], candidates: [first] },
      { role: 'user', content: 'もっと明るく' },
      { role: 'assistant', content: '', mood: [], candidates: [second] },
    ];
    expect(imageFor(turns, first)).toBe('雨の日曜');
    expect(imageFor(turns, second)).toBe('もっと明るく');
  });
});
