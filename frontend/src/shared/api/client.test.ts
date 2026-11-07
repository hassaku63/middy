import { describe, expect, it, vi } from 'vitest';
import { createTimelineFromProgression, fetchHealth, type ApiError, type Timeline } from './client';

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
});

describe('createTimelineFromProgression', () => {
  const timeline: Timeline = {
    ppq: 480,
    tempoBpm: 90,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 0, mode: 'major' },
    tracks: [],
    harmony: [
      {
        start: 0,
        duration: 1920,
        chord: { root: 0, quality: 'maj' },
        romanNumeral: 'I',
        analysis: { degree: 1, function: 'T', explanation: 'I（C）はトニック' },
      },
    ],
  };

  it('posts the request as JSON and unwraps the timeline', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { timeline }));

    const got = await createTimelineFromProgression(
      { key: { tonic: 0, mode: 'major' }, tempoBpm: 90, progression: [{ numeral: 'I' }] },
      fetchMock,
    );

    expect(got).toEqual(timeline);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/v1/timelines/from-progression');
    expect(init.method).toBe('POST');
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
    expect(JSON.parse(init.body as string)).toEqual({
      key: { tonic: 0, mode: 'major' },
      tempoBpm: 90,
      progression: [{ numeral: 'I' }],
    });
  });

  it('surfaces validation errors from the envelope', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        jsonResponse(400, { error: { code: 'invalid_request', message: 'progression[1]: bad' } }),
      );

    await expect(
      createTimelineFromProgression(
        { key: { tonic: 0, mode: 'major' }, tempoBpm: 90, progression: [{ numeral: 'X' }] },
        fetchMock,
      ),
    ).rejects.toMatchObject({ status: 400, code: 'invalid_request' } satisfies Partial<ApiError>);
  });
});
