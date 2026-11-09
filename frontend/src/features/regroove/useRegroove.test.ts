import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PatternsResponse, Timeline } from '@/shared/api/client';
import timelineFixture from '@/shared/api/fixtures/timeline.json';
import { useRegroove } from './useRegroove';

const timeline = timelineFixture as Timeline;
const rock: PatternsResponse['styles'][number] = {
  id: 'rock',
  name: 'ロック',
  description: '',
  accompaniments: [],
  drums: [],
  feels: [],
  grooves: [
    {
      meter: { numerator: 4, denominator: 4 },
      accompaniment: 'power',
      drums: 'rock',
      feel: 'straight',
    },
  ],
};

function json(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status });
}

describe('useRegroove', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('resolves to the new timeline and says what changed', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json(200, { timeline })));
    const { result } = renderHook(() => useRegroove());
    let next: Timeline | null = null;
    await act(async () => {
      next = await result.current.change(timeline, rock);
    });
    expect(next).toEqual(timeline);
    expect(result.current.error).toBeNull();
    expect(result.current.note).not.toBeNull();
    expect(result.current.busy).toBe(false);
  });

  it('turns an API failure into the error state', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(json(500, { error: { code: 'internal_error', message: 'x' } })),
    );
    const { result } = renderHook(() => useRegroove());
    let next: Timeline | null = timeline;
    await act(async () => {
      next = await result.current.change(timeline, rock);
    });
    expect(next).toBeNull();
    expect(result.current.error).toBe('スタイルを替えられませんでした');
    expect(result.current.busy).toBe(false);
  });

  it('lets a fault in its own request building through instead of hiding it (BL-157)', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const broken = { ...timeline, harmony: undefined } as unknown as Timeline;
    const { result } = renderHook(() => useRegroove());
    await act(async () => {
      await expect(result.current.change(broken, rock)).rejects.toThrow(TypeError);
    });
    expect(fetchMock).not.toHaveBeenCalled();
    expect(result.current.error).toBeNull();
    expect(result.current.busy).toBe(false);
  });
});
