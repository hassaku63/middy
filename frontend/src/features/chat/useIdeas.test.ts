import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { IdeasResponse } from '@/shared/api/client';
import { useIdeas } from './useIdeas';

const reply: IdeasResponse = {
  reply: '静かな雨。',
  mood: ['静か'],
  candidates: [],
};

function json(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status });
}

describe('useIdeas', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('sends the full history every turn and appends the reply', async () => {
    // A fresh Response per call: a body can only be read once.
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(() =>
      Promise.resolve(json(200, reply)),
    );
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useIdeas());

    await act(() => result.current.send('雨の日曜'));
    await waitFor(() => expect(result.current.busy).toBe(false));
    expect(result.current.turns).toEqual([
      { role: 'user', content: '雨の日曜' },
      { role: 'assistant', content: '静かな雨。', mood: ['静か'], candidates: [] },
    ]);

    await act(() => result.current.send('もう少し明るく'));
    await waitFor(() => expect(result.current.turns).toHaveLength(4));
    const secondInit = fetchMock.mock.calls[1]?.[1];
    const secondBody = JSON.parse(secondInit?.body as string) as { messages: unknown[] };
    expect(secondBody.messages).toEqual([
      { role: 'user', content: '雨の日曜' },
      { role: 'assistant', content: '静かな雨。' },
      { role: 'user', content: 'もう少し明るく' },
    ]);
  });

  it('translates provider errors into friendly text and keeps the user turn', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(json(502, { code: 'llm_unavailable', detail: 'x' })),
    );
    const { result } = renderHook(() => useIdeas());
    await act(() => result.current.send('雨'));
    await waitFor(() => expect(result.current.error).toMatch(/AI に接続できません/));
    expect(result.current.turns).toEqual([{ role: 'user', content: '雨' }]);

    act(() => result.current.reset());
    expect(result.current.turns).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it('asks the user to rephrase when the model declines', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(json(422, { code: 'llm_declined', detail: 'x' })),
    );
    const { result } = renderHook(() => useIdeas());
    await act(() => result.current.send('深夜に一人で勉強している集中した時間'));
    await waitFor(() => expect(result.current.error).toMatch(/別の言葉で/));
    expect(result.current.turns).toHaveLength(1);
  });

  it('reports each answer once, with the image it answers', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(json(200, reply))),
    );
    const onAnswer = vi.fn();
    const { result } = renderHook(() => useIdeas(onAnswer));

    await act(() => result.current.send('雨の日曜'));
    await act(() => result.current.send('もう少し明るく'));
    expect(onAnswer.mock.calls).toEqual([
      [{ image: '雨の日曜', candidates: [] }],
      [{ image: 'もう少し明るく', candidates: [] }],
    ]);
  });

  it('does not report an answer when the request fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(json(502, { code: 'llm_unavailable', detail: 'x' }))),
    );
    const onAnswer = vi.fn();
    const { result } = renderHook(() => useIdeas(onAnswer));
    await act(() => result.current.send('雨'));
    expect(onAnswer).not.toHaveBeenCalled();
  });

  it('says what went wrong in Japanese for any other failure (review)', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          json(500, { code: 'internal_error', detail: 'the examples could not be built' }),
        ),
    );
    const { result } = renderHook(() => useIdeas());
    await act(() => result.current.send('雨'));
    expect(result.current.error).toBe(
      'AI の答えから例を作れませんでした。もう一度送ってください。',
    );
  });

  it('does not report a fault in the answer handler as a network failure (review)', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json(200, reply)));
    const { result } = renderHook(() =>
      useIdeas(() => {
        throw new Error('handler');
      }),
    );
    await act(() => result.current.send('雨').catch(() => undefined));
    expect(result.current.error).toBeNull();
    expect(result.current.turns).toHaveLength(2);
    expect(result.current.busy).toBe(false);
  });

  it('ignores blank input', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useIdeas());
    await act(() => result.current.send('   '));
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
