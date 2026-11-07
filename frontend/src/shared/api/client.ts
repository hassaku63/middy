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

// --- domain model (mirrors backend/internal/music) ---------------------------

/** 0 = C … 11 = B */
export type PitchClass = number;
/** MIDI note number, 60 = C4 */
export type Pitch = number;
export type Mode = 'major' | 'minor';
export type Key = { tonic: PitchClass; mode: Mode };
export type ChordQuality = 'maj' | 'min' | 'dim' | 'aug' | 'maj7' | 'min7' | 'dom7' | 'm7b5';
export type Chord = { root: PitchClass; quality: ChordQuality };
export type TimeSignature = { numerator: number; denominator: number };
export type NoteEvent = { pitch: Pitch; velocity: number; start: number; duration: number };
export type Track = { name: string; channel: number; program: number; notes: NoteEvent[] };
export type HarmonicFunction = 'T' | 'S' | 'D';
export type Cadence = 'authentic' | 'plagal' | 'half' | 'deceptive';
export type HarmonyAnalysis = {
  degree: number;
  function: HarmonicFunction;
  cadence?: Cadence;
  explanation: string;
};
export type HarmonyEvent = {
  start: number;
  duration: number;
  chord: Chord;
  romanNumeral: string;
  analysis?: HarmonyAnalysis;
};
export type Timeline = {
  ppq: number;
  tempoBpm: number;
  timeSignature: TimeSignature;
  key: Key;
  tracks: Track[];
  harmony: HarmonyEvent[];
};

// --- POST /timelines/from-progression ---------------------------------------

export type ProgressionLabel = { numeral: string; beats?: number };
export type FromProgressionRequest = {
  key: Key;
  tempoBpm: number;
  progression: ProgressionLabel[];
};
export type TimelineResponse = { timeline: Timeline };

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

/**
 * `fetch` must be called with the global as `this`; passing the bare
 * function around and invoking it as a property throws "Illegal invocation"
 * in browsers. Wrapping it once keeps every call site safe.
 */
const defaultFetch: typeof fetch = (input, init) => globalThis.fetch(input, init);

export async function fetchHealth(fetchImpl: typeof fetch = defaultFetch): Promise<Health> {
  return requestJson<Health>('/api/v1/health', { fetchImpl });
}

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

// --- POST /ideas -------------------------------------------------------------

export type ChatRole = 'user' | 'assistant';
export type ChatMessage = { role: ChatRole; content: string };
export type IdeasRequest = { messages: ChatMessage[] };
export type IdeaCandidate = {
  title: string;
  key: Key;
  tempoBpm: number;
  progression: string[];
  beatsPerChord: number;
  reason: string;
  timeline: Timeline;
};
export type IdeasResponse = { reply: string; mood: string[]; candidates: IdeaCandidate[] };

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
  const headers: Record<string, string> = { Accept: opts.accept ?? 'application/json' };
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
  return res;
}
