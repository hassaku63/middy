import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import timelineFixture from '@/shared/api/fixtures/timeline.json';
import { useKeyChange } from './useKeyChange';

const timeline = timelineFixture as Timeline;

function json(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status });
}

describe('useKeyChange', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('resolves to the timeline the engine returns', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json(200, { timeline })));
    const { result } = renderHook(() => useKeyChange());
    let next: Timeline | null = null;
    await act(async () => {
      next = await result.current.change(timeline, 7);
    });
    expect(next).toEqual(timeline);
    expect(result.current.error).toBeNull();
    expect(result.current.busy).toBe(false);
  });

  it('turns an API failure into the error state', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(json(500, { error: { code: 'internal_error', message: 'x' } })),
    );
    const { result } = renderHook(() => useKeyChange());
    let next: Timeline | null = timeline;
    await act(async () => {
      next = await result.current.change(timeline, 7);
    });
    expect(next).toBeNull();
    expect(result.current.error).toBe('キーを変えられませんでした');
    expect(result.current.busy).toBe(false);
  });

  it('lets a fault in its own request building through instead of hiding it (BL-157)', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const broken = { ...timeline, harmony: undefined } as unknown as Timeline;
    const { result } = renderHook(() => useKeyChange());
    await act(async () => {
      await expect(result.current.change(broken, 7)).rejects.toThrow(TypeError);
    });
    expect(fetchMock).not.toHaveBeenCalled();
    expect(result.current.error).toBeNull();
    expect(result.current.busy).toBe(false);
  });
});
