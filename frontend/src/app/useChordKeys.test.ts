import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { HarmonyEvent } from '@/shared/api/client';
import { useChordKeys } from './useChordKeys';

const chord = (start: number) => ({ start, duration: 480 }) as HarmonyEvent;
const harmony = [chord(0), chord(480), chord(960)];

function press(key: string, target: EventTarget = window) {
  const e = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true });
  target.dispatchEvent(e);
  return e;
}

function setup(selected: number | null, positionTick = 500, enabled = true) {
  const onSelect = vi.fn();
  renderHook(() => useChordKeys({ enabled, harmony, positionTick, selected, onSelect }));
  return onSelect;
}

describe('useChordKeys', () => {
  it('moves from the chord under the playhead and pins it', () => {
    const onSelect = setup(null);
    expect(press('ArrowRight').defaultPrevented).toBe(true);
    expect(onSelect).toHaveBeenLastCalledWith(2);
    press('ArrowLeft');
    expect(onSelect).toHaveBeenLastCalledWith(0);
  });

  it('moves from the pinned chord and stops at the ends', () => {
    const onSelect = setup(2);
    press('ArrowRight');
    expect(onSelect).toHaveBeenLastCalledWith(2);
  });

  it('starts at the first chord when none is under the playhead', () => {
    const onSelect = setup(null, 5000);
    press('ArrowRight');
    expect(onSelect).toHaveBeenLastCalledWith(0);
  });

  it('lets Esc follow the playhead again, and leaves Esc alone when nothing is pinned', () => {
    setup(null);
    expect(press('Escape').defaultPrevented).toBe(false);
    const onSelect = setup(1);
    press('Escape');
    expect(onSelect).toHaveBeenLastCalledWith(null);
  });

  it('leaves keys typed into a field, and does nothing when off', () => {
    const onSelect = setup(1);
    const input = document.createElement('input');
    document.body.append(input);
    press('ArrowRight', input);
    input.remove();
    expect(onSelect).not.toHaveBeenCalled();

    const off = setup(1, 500, false);
    press('ArrowRight');
    expect(off).not.toHaveBeenCalled();
  });
});
