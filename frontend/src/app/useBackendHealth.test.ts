import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { RETRY_MS, useBackendHealth } from './useBackendHealth';

const ok = () => new Response(JSON.stringify({ status: 'ok', version: 'dev' }), { status: 200 });

describe('useBackendHealth', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('does not probe while offline, and probes again when back online', async () => {
    const fetchMock = vi.fn(() => Promise.resolve(ok()));
    vi.stubGlobal('fetch', fetchMock);
    const { result, rerender } = renderHook(({ online }) => useBackendHealth(online), {
      initialProps: { online: false },
    });
    expect(fetchMock).not.toHaveBeenCalled();
    expect(result.current).toBe('checking');

    rerender({ online: true });
    await waitFor(() => expect(result.current).toBe('ok'));
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('reports a backend that does not answer', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('network')));
    const { result } = renderHook(() => useBackendHealth(true));
    await waitFor(() => expect(result.current).toBe('down'));
  });

  it('asks again while the backend does not answer, and clears the notice when it does (review)', async () => {
    vi.useFakeTimers();
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new TypeError('network'))
      .mockResolvedValue(ok());
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useBackendHealth(true));
    await act(() => vi.advanceTimersByTimeAsync(0));
    expect(result.current).toBe('down');
    await act(() => vi.advanceTimersByTimeAsync(RETRY_MS));
    expect(result.current).toBe('ok');
    expect(fetchMock).toHaveBeenCalledTimes(2);
    await act(() => vi.advanceTimersByTimeAsync(RETRY_MS * 3));
    expect(fetchMock).toHaveBeenCalledTimes(2); // no polling once it answers
  });
});
