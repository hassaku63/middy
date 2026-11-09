import { describe, expect, it, vi } from 'vitest';
import {
  exportMidi,
  fetchHealth,
  imagine,
  type ApiError,
  type IdeasResponse,
  type Timeline,
} from './client';

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('fetchHealth', () => {
  it('returns the parsed health payload', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse(200, { status: 'ok', version: 'dev' }));

    const health = await fetchHealth(fetchMock);

    expect(health).toEqual({ status: 'ok', version: 'dev' });
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/health', expect.anything());
  });

  it('throws ApiError using the error envelope', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse(503, { error: { code: 'unavailable', message: 'down' } }));

    await expect(fetchHealth(fetchMock)).rejects.toMatchObject({
      name: 'ApiError',
      status: 503,
      code: 'unavailable',
      message: 'down',
    } satisfies Partial<ApiError>);
  });

  it('falls back to the HTTP status when the error body is not JSON', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        new Response('boom', { status: 500, statusText: 'Internal Server Error' }),
      );

    await expect(fetchHealth(fetchMock)).rejects.toMatchObject({
      status: 500,
      code: 'http_error',
    });
  });

  it.each([null, { error: 'oops' }, { error: { code: 1, message: 'x' } }])(
    'falls back to the HTTP status for JSON that is not the error envelope: %j (review)',
    async (body) => {
      const fetchMock = vi
        .fn()
        .mockResolvedValue(
          new Response(JSON.stringify(body), { status: 502, statusText: 'Bad Gateway' }),
        );
      await expect(fetchHealth(fetchMock)).rejects.toMatchObject({
        status: 502,
        code: 'http_error',
        message: '502 Bad Gateway',
      });
    },
  );
});

describe('exportMidi', () => {
  it('posts the timeline and returns the binary body', async () => {
    const bytes = new Uint8Array([0x4d, 0x54, 0x68, 0x64]);
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        new Response(bytes, { status: 200, headers: { 'Content-Type': 'audio/midi' } }),
      );
    const timeline: Timeline = {
      ppq: 480,
      tempoBpm: 90,
      timeSignature: { numerator: 4, denominator: 4 },
      key: { tonic: 0, mode: 'major' },
      tracks: [],
      harmony: [],
    };

    const blob = await exportMidi(timeline, fetchMock);

    expect(new Uint8Array(await blob.arrayBuffer())).toEqual(bytes);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/v1/export/midi');
    expect((init.headers as Record<string, string>).Accept).toBe('audio/midi');
    expect(JSON.parse(init.body as string)).toEqual({ timeline });
  });
});

describe('imagine', () => {
  it('posts the conversation and returns the reply with candidates', async () => {
    const response: IdeasResponse = {
      reply: '静かな雨。',
      mood: ['静か'],
      candidates: [
        {
          title: '窓辺の雨',
          key: { tonic: 9, mode: 'minor' },
          tempoBpm: 72,
          progression: ['i', 'VI', 'III', 'VII'],
          beatsPerChord: 4,
          reason: '短調で。',
          timeline: {
            ppq: 480,
            tempoBpm: 72,
            timeSignature: { numerator: 4, denominator: 4 },
            key: { tonic: 9, mode: 'minor' },
            tracks: [],
            harmony: [],
          },
        },
      ],
    };
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, response));

    const got = await imagine([{ role: 'user', content: '雨の日曜' }], fetchMock);

    expect(got).toEqual(response);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/v1/ideas');
    expect(JSON.parse(init.body as string)).toEqual({
      messages: [{ role: 'user', content: '雨の日曜' }],
    });
  });

  it('maps 502 envelopes to ApiError', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        jsonResponse(502, { error: { code: 'llm_unavailable', message: 'down' } }),
      );
    await expect(imagine([{ role: 'user', content: 'x' }], fetchMock)).rejects.toMatchObject({
      status: 502,
      code: 'llm_unavailable',
    } satisfies Partial<ApiError>);
  });
});
