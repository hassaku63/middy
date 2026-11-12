import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useCopyText } from './useCopyText';

describe('useCopyText (BL-300)', () => {
  afterEach(() => vi.useRealTimers());

  it('says it copied, then rests again', async () => {
    vi.useFakeTimers();
    const write = vi.fn(() => Promise.resolve());
    const { result } = renderHook(() => useCopyText(write));
    await act(() => result.current.copy('解説'));
    expect(write).toHaveBeenCalledWith('解説');
    expect(result.current.state).toBe('copied');
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(result.current.state).toBe('idle');
  });

  it('says it failed when the browser refuses, until the next try', async () => {
    const write = vi.fn<(text: string) => Promise<void>>(() =>
      Promise.reject(new Error('NotAllowedError')),
    );
    const { result } = renderHook(() => useCopyText(write));
    await act(() => result.current.copy('解説'));
    expect(result.current.state).toBe('failed');
    write.mockImplementationOnce(() => Promise.resolve());
    await act(() => result.current.copy('解説'));
    expect(result.current.state).toBe('copied');
  });
});
