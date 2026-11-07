import { describe, expect, it, vi } from 'vitest';
import { fetchHealth, type ApiError } from './client';

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
