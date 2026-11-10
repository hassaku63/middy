import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useHearOnOpen, useSpaceToPlay } from './useHearOnOpen';

function press(key: string, target: EventTarget = window) {
  const e = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true });
  target.dispatchEvent(e);
  return e;
}

describe('useHearOnOpen (#315)', () => {
  afterEach(() => window.localStorage.clear());

  it('is on by default and remembers being turned off', () => {
    const { result, unmount } = renderHook(() => useHearOnOpen());
    expect(result.current.on).toBe(true);
    act(() => result.current.choose(false));
    expect(result.current.on).toBe(false);
    unmount();
    expect(renderHook(() => useHearOnOpen()).result.current.on).toBe(false);
  });
});

describe('useSpaceToPlay (#315)', () => {
  it('toggles on Space, but leaves Space to fields, buttons and chords', () => {
    const toggle = vi.fn();
    renderHook(() => useSpaceToPlay(true, toggle));
    expect(press(' ').defaultPrevented).toBe(true);
    expect(toggle).toHaveBeenCalledTimes(1);
    const option = document.createElement('div');
    option.setAttribute('role', 'option');
    for (const el of [document.createElement('input'), document.createElement('button'), option]) {
      document.body.append(el);
      press(' ', el);
      el.remove();
    }
    expect(toggle).toHaveBeenCalledTimes(1);
  });

  it('does nothing when off', () => {
    const toggle = vi.fn();
    renderHook(() => useSpaceToPlay(false, toggle));
    press(' ');
    expect(toggle).not.toHaveBeenCalled();
  });
});
