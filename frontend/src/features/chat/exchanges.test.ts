import { describe, expect, it } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { exchangesOf } from './exchanges';

describe('exchangesOf', () => {
  it('numbers the exchanges and gathers each answer’s examples', () => {
    const a = { title: 'a' } as IdeaCandidate;
    const b = { title: 'b' } as IdeaCandidate;
    expect(
      exchangesOf([
        { role: 'user', content: '雨' },
        { role: 'assistant', content: '', mood: [], candidates: [a, b] },
        { role: 'user', content: '明るく' },
      ]),
    ).toEqual([
      { number: 1, image: '雨', candidates: [a, b] },
      { number: 2, image: '明るく', candidates: [] },
    ]);
  });
});
