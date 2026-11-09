import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { usePinMemory } from './usePinMemory';

describe('usePinMemory', () => {
  it('recalls the chord pinned on each example as it was left', () => {
    const a = { title: 'a' } as IdeaCandidate;
    const b = { title: 'b' } as IdeaCandidate;
    const initialProps: { open: IdeaCandidate | null; pinned: number | null } = {
      open: a,
      pinned: 2,
    };
    const { result, rerender } = renderHook(({ open, pinned }) => usePinMemory(open, pinned), {
      initialProps,
    });
    rerender({ open: b, pinned: null });
    rerender({ open: b, pinned: 1 });
    expect(result.current(a)).toBe(2);
    expect(result.current(b)).toBe(1);
    expect(result.current({ title: 'c' } as IdeaCandidate)).toBeNull();
  });
});
