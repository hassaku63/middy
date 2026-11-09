import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { useCandidateKeys } from './useCandidateKeys';

const a = { title: 'a' } as IdeaCandidate;
const b = { title: 'b' } as IdeaCandidate;

function press(key: string, target: EventTarget = window) {
  const e = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true });
  target.dispatchEvent(e);
  return e;
}

describe('useCandidateKeys', () => {
  it('opens the latest answer’s examples by their number', () => {
    const open = vi.fn();
    renderHook(() => useCandidateKeys(true, [a, b], open));
    expect(press('2').defaultPrevented).toBe(true);
    expect(open).toHaveBeenLastCalledWith(b);
    press('1');
    expect(open).toHaveBeenLastCalledWith(a);
    expect(press('3').defaultPrevented).toBe(false);
    expect(open).toHaveBeenCalledTimes(2);
  });

  it('leaves digits typed into a field, and does nothing when off', () => {
    const open = vi.fn();
    renderHook(() => useCandidateKeys(true, [a, b], open));
    const input = document.createElement('input');
    document.body.append(input);
    press('1', input);
    input.remove();
    expect(open).not.toHaveBeenCalled();

    const off = vi.fn();
    renderHook(() => useCandidateKeys(false, [a, b], off));
    press('1');
    expect(off).not.toHaveBeenCalled();
  });
});
