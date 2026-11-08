import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PatternsResponse, Timeline } from '@/shared/api/client';
import { useExampleChanges } from './useExampleChanges';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [{ start: 0, duration: 1920, chord: { root: 0, quality: 'maj' }, romanNumeral: 'I' }],
};

const styles: PatternsResponse['styles'] = [
  {
    id: 'rock',
    name: 'ロック',
    description: '',
    accompaniments: ['eighth-pulse'],
    drums: ['rock'],
    feels: ['straight'],
    grooves: [
      {
        meter: { numerator: 4, denominator: 4 },
        accompaniment: 'eighth-pulse',
        drums: 'rock',
        feel: 'straight',
      },
    ],
  },
];

const json = (status: number, body: unknown) => new Response(JSON.stringify(body), { status });
const failure = () => json(400, { error: { code: 'invalid', message: '作れません' } });

/** A fetch whose answers the test hands out one at a time. */
function heldFetch() {
  const answers: ((r: Response) => void)[] = [];
  vi.stubGlobal(
    'fetch',
    vi.fn(() => new Promise<Response>((resolve) => answers.push(resolve))),
  );
  return (r: Response) => answers.shift()?.(r);
}

function setup() {
  const show = vi.fn();
  const hook = renderHook(() => useExampleChanges(timeline, show, styles));
  return { show, hook };
}

describe('useExampleChanges (review)', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('drops a style change and its note when an example opens meanwhile', async () => {
    const answer = heldFetch();
    const { show, hook } = setup();
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = hook.result.current.changeStyle('rock');
    });
    act(() => hook.result.current.forget()); // the same example opened again
    await act(async () => {
      answer(json(200, { timeline: { ...timeline, tempoBpm: 120 } }));
      await pending;
    });
    expect(show).not.toHaveBeenCalled();
    expect(hook.result.current.regroove.note).toBeNull();
  });

  it('drops the error of a key change that fails after another example opened', async () => {
    const answer = heldFetch();
    const { hook } = setup();
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = hook.result.current.changeKey(2);
    });
    act(() => hook.result.current.forget());
    await act(async () => {
      answer(failure());
      await pending;
    });
    expect(hook.result.current.keyChange.error).toBeNull();
  });

  it('forgets a key error when a style is chosen', async () => {
    const answer = heldFetch();
    const { hook } = setup();
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = hook.result.current.changeKey(2);
    });
    await act(async () => {
      answer(failure());
      await pending;
    });
    expect(hook.result.current.keyChange.error).toBe('キーを変えられませんでした');
    act(() => {
      pending = hook.result.current.changeStyle('rock');
    });
    expect(hook.result.current.keyChange.error).toBeNull();
    await act(async () => {
      answer(json(200, { timeline }));
      await pending;
    });
  });

  it('shows a change that finishes on the same example', async () => {
    const answer = heldFetch();
    const { show, hook } = setup();
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = hook.result.current.changeKey(2);
    });
    const inD = { ...timeline, key: { tonic: 2, mode: 'major' as const } };
    await act(async () => {
      answer(json(200, { timeline: inD }));
      await pending;
    });
    expect(show).toHaveBeenCalledWith(inD);
  });
});
