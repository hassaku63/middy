import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { useLayoutTrial } from './useLayoutTrial';

describe('useLayoutTrial', () => {
  afterEach(() => window.history.replaceState(null, '', '/'));

  it("keeps today's layout unless the URL asks for another", () => {
    window.history.replaceState(null, '', '/?lab&layout=nope');
    const { result } = renderHook(() => useLayoutTrial());
    expect(result.current.mode).toBe('stacked');
    expect(result.current.desk).toBe(false);
    expect(result.current.lab).toBe(true);
  });

  it('reads ?layout=desk and writes the choice back to the URL', () => {
    window.history.replaceState(null, '', '/?layout=desk&zoom=page');
    const { result } = renderHook(() => useLayoutTrial());
    expect(result.current.desk).toBe(true);
    expect(result.current.wide).toBe(false);

    act(() => result.current.choose('stacked'));
    expect(result.current.desk).toBe(false);
    expect(window.location.search).toBe('?zoom=page');

    act(() => result.current.choose('desk'));
    expect(window.location.search).toBe('?zoom=page&layout=desk');
  });

  it('reads ?layout=wide as a desk with the timeline across the stage (#314)', () => {
    window.history.replaceState(null, '', '/?layout=wide');
    const { result } = renderHook(() => useLayoutTrial());
    expect(result.current.mode).toBe('wide');
    expect(result.current.desk).toBe(true);
    expect(result.current.wide).toBe(true);
  });
});
