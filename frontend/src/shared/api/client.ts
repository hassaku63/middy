/**
 * Thin fetch wrapper for the backend API.
 *
 * Request/response types are generated from docs/api/openapi.yaml into
 * schema.gen.ts (`npm run gen:api`, ADR-0009); this file only gives them
 * short names. Never hand-write an API shape here.
 */

import type { components } from './schema.gen';

type Schemas = components['schemas'];

export type Health = Schemas['Health'];

// --- domain model (backend/internal/music, via the OpenAPI schemas) ---------

/** 0 = C … 11 = B */
export type PitchClass = Schemas['PitchClass'];
/** MIDI note number, 60 = C4 */
export type Pitch = Schemas['Pitch'];
export type Key = Schemas['Key'];
export type ChordQuality = Schemas['ChordQuality'];
export type Chord = Schemas['Chord'];
export type TimeSignature = Schemas['TimeSignature'];
export type NoteEvent = Schemas['NoteEvent'];
export type HarmonyAnalysis = Schemas['HarmonyAnalysis'];
export type HarmonicFunction = HarmonyAnalysis['function'];
export type Cadence = NonNullable<HarmonyAnalysis['cadence']>;
export type HarmonyEvent = Schemas['HarmonyEvent'];
export type RhythmPart = Schemas['RhythmPart'];
export type Rhythm = Schemas['Rhythm'];
export type Overview = Schemas['Overview'];
export type DegreeChord = Schemas['DegreeChord'];
export type Timeline = Schemas['Timeline'];

// --- POST /timelines/from-progression (changing an example's key; older projects) --

export type FromProgressionRequest = Schemas['FromProgressionRequest'];
type TimelineResponse = Schemas['TimelineResponse'];

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

/**
 * The code and developer message of an RFC 9457 problem detail, which the
 * API and nginx answer every error with (ADR-0030). Null for anything
 * else, such as a proxy's own page.
 */
function readError(body: unknown): { code: string; message: string } | null {
  if (typeof body !== 'object' || body === null) return null;
  if ('code' in body && typeof body.code === 'string' && 'detail' in body) {
    return { code: body.code, message: typeof body.detail === 'string' ? body.detail : '' };
  }
  return null;
}

/**
 * `fetch` must be called with the global as `this`; passing the bare
 * function around and invoking it as a property throws "Illegal invocation"
 * in browsers. Wrapping it once keeps every call site safe.
 */
const defaultFetch: typeof fetch = (input, init) => globalThis.fetch(input, init);

export async function fetchHealth(fetchImpl: typeof fetch = defaultFetch): Promise<Health> {
  return requestJson<Health>('/api/v1/health', { fetchImpl });
}

/** Renders a progression in a key without the LLM (used to change an example's key). */
export async function createTimelineFromProgression(
  req: FromProgressionRequest,
  fetchImpl: typeof fetch = defaultFetch,
): Promise<Timeline> {
  const res = await requestJson<TimelineResponse>('/api/v1/timelines/from-progression', {
    fetchImpl,
    method: 'POST',
    body: req,
  });
  return res.timeline;
}

// --- GET /patterns -----------------------------------------------------------

export type PatternsResponse = Schemas['PatternsResponse'];

export async function fetchPatterns(
  fetchImpl: typeof fetch = defaultFetch,
): Promise<PatternsResponse> {
  return requestJson<PatternsResponse>('/api/v1/patterns', { fetchImpl });
}

// --- POST /ideas -------------------------------------------------------------

export type ChatMessage = Schemas['ChatMessage'];
export type IdeasRequest = Schemas['IdeasRequest'];
export type IdeaCandidate = Schemas['IdeaCandidate'];
export type IdeasResponse = Schemas['IdeasResponse'];

/** One conversational turn of scenario S2; the server keeps no state. */
export async function imagine(
  messages: ChatMessage[],
  fetchImpl: typeof fetch = defaultFetch,
): Promise<IdeasResponse> {
  return requestJson<IdeasResponse>('/api/v1/ideas', {
    fetchImpl,
    method: 'POST',
    body: { messages } satisfies IdeasRequest,
  });
}

// --- POST /export/midi -------------------------------------------------------

/** Returns the timeline as a Standard MIDI File blob (audio/midi). */
export async function exportMidi(
  timeline: Timeline,
  fetchImpl: typeof fetch = defaultFetch,
): Promise<Blob> {
  const res = await request('/api/v1/export/midi', {
    fetchImpl,
    method: 'POST',
    body: { timeline },
    accept: 'audio/midi',
  });
  return res.blob();
}

type RequestOptions = {
  fetchImpl: typeof fetch;
  method?: 'GET' | 'POST';
  body?: unknown;
  accept?: string;
};

async function requestJson<T>(path: string, opts: RequestOptions): Promise<T> {
  const res = await request(path, opts);
  return (await res.json()) as T;
}

/** Performs the request and turns non-2xx responses into ApiError. */
async function request(path: string, opts: RequestOptions): Promise<Response> {
  // Errors come as problem details whatever the success type is (ADR-0030).
  const headers: Record<string, string> = {
    Accept: `${opts.accept ?? 'application/json'}, application/problem+json`,
  };
  const init: RequestInit = { method: opts.method ?? 'GET', headers };
  if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json';
    init.body = JSON.stringify(opts.body);
  }
  const res = await opts.fetchImpl(path, init);
  if (!res.ok) {
    let code = 'http_error';
    let message = `${res.status} ${res.statusText}`;
    try {
      const error = readError(await res.json());
      if (error) {
        code = error.code;
        message = error.message;
      }
    } catch {
      // Non-JSON error body: keep the HTTP status message.
    }
    throw new ApiError(res.status, code, message);
  }
  return res;
}
