import { describe, expect, it } from 'vitest';
import type { ChatMessage } from '@/shared/api/client';
import { MAX_MESSAGES, recentHistory } from './useIdeas';

describe('recentHistory (review)', () => {
  it('keeps a long conversation within the limit, starting from a user message', () => {
    const msgs: ChatMessage[] = Array.from({ length: 23 }, (_, i) => ({
      role: i % 2 === 0 ? 'user' : 'assistant',
      content: `m${i}`,
    }));
    const out = recentHistory(msgs);
    expect(out.length).toBeLessThanOrEqual(MAX_MESSAGES);
    expect(out[0]?.role).toBe('user');
    expect(out.at(-1)?.content).toBe('m22');
  });

  it('leaves a short conversation as it is', () => {
    const msgs: ChatMessage[] = [{ role: 'user', content: '雨' }];
    expect(recentHistory(msgs)).toEqual(msgs);
  });
});
