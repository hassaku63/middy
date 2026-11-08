import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { usePatterns } from './usePatterns';

const catalogue = { accompaniments: [], drums: [], feels: [], styles: [{ id: 'rock' }] };

describe('usePatterns (review)', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('loads again once the backend answers after a failed start', async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new TypeError('network'))
      .mockResolvedValue(new Response(JSON.stringify(catalogue), { status: 200 }));
    vi.stubGlobal('fetch', fetchMock);
    const { result, rerender } = renderHook(({ reachable }) => usePatterns(reachable), {
      initialProps: { reachable: true },
    });
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(result.current.styles).toEqual([]);

    rerender({ reachable: false }); // the health probe found it down
    rerender({ reachable: true }); // and then up again
    await waitFor(() => expect(result.current.styles).toHaveLength(1));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    rerender({ reachable: false });
    rerender({ reachable: true });
    expect(fetchMock).toHaveBeenCalledTimes(2); // loaded once is enough
  });
});
