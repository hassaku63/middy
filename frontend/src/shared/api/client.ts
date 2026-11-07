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

export async function fetchHealth(fetchImpl: typeof fetch = fetch): Promise<Health> {
  return requestJson<Health>('/api/v1/health', { fetchImpl });
}

export async function createTimelineFromProgression(
  req: FromProgressionRequest,
  fetchImpl: typeof fetch = fetch,
): Promise<Timeline> {
  const res = await requestJson<TimelineResponse>('/api/v1/timelines/from-progression', {
    fetchImpl,
    method: 'POST',
    body: req,
  });
  return res.timeline;
}

type RequestOptions = { fetchImpl: typeof fetch; method?: 'GET' | 'POST'; body?: unknown };

async function requestJson<T>(path: string, opts: RequestOptions): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
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
  return (await res.json()) as T;
}
