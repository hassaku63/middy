/**
 * Thin fetch wrapper for the backend API.
 *
 * Request/response types mirror docs/api/openapi.yaml. When the contract
 * grows beyond a handful of endpoints, replace the hand-written types with
 * generated ones (tracked in the backlog) rather than letting them drift.
 */

export type Health = {
  status: 'ok';
  version: string;
};

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

type ErrorEnvelope = { error: { code: string; message: string } };

export async function fetchHealth(fetchImpl: typeof fetch = fetch): Promise<Health> {
  return requestJson<Health>('/api/v1/health', fetchImpl);
}

async function requestJson<T>(path: string, fetchImpl: typeof fetch): Promise<T> {
  const res = await fetchImpl(path, { headers: { Accept: 'application/json' } });
  if (!res.ok) {
    let code = 'http_error';
    let message = `${res.status} ${res.statusText}`;
    try {
      const body = (await res.json()) as Partial<ErrorEnvelope>;
      if (body.error) {
        code = body.error.code;
        message = body.error.message;
      }
    } catch {
      // Non-JSON error body: keep the HTTP status message.
    }
    throw new ApiError(res.status, code, message);
  }
  return (await res.json()) as T;
}
