import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { useLayoutTrial } from './useLayoutTrial';

describe('useLayoutTrial', () => {
  afterEach(() => window.history.replaceState(null, '', '/'));

  it("keeps today's layout unless the URL asks for the desk", () => {
    window.history.replaceState(null, '', '/?lab');
    const { result } = renderHook(() => useLayoutTrial());
    expect(result.current.desk).toBe(false);
    expect(result.current.lab).toBe(true);
  });

  it('reads ?layout=desk and writes the choice back to the URL', () => {
    window.history.replaceState(null, '', '/?layout=desk&zoom=page');
    const { result } = renderHook(() => useLayoutTrial());
    expect(result.current.desk).toBe(true);

    act(() => result.current.choose(false));
    expect(result.current.desk).toBe(false);
    expect(window.location.search).toBe('?zoom=page');

    act(() => result.current.choose(true));
    expect(window.location.search).toBe('?zoom=page&layout=desk');
  });
});
