import type {
  FromProgressionRequest,
  PatternsResponse,
  PitchClass,
  Timeline,
} from '@/shared/api/client';
import { ticksPerBeat } from '@/shared/lib/timing';

/**
 * The request that renders the same example in another key (BL-078): the
 * timeline's own numerals, chord lengths, tempo, meter, mode and patterns,
 * with a new tonic. The engine guarantees the numerals read back as the same
 * chords (music.TestHarmonyLabelsReadBackAsTheSameChords).
 */
export function rebuildRequest(timeline: Timeline, tonic: PitchClass): FromProgressionRequest {
  const req = sameHarmony(timeline, tonic);
  const rhythm = timeline.rhythm;
  if (rhythm?.accompaniment) req.accompaniment = rhythm.accompaniment.id;
  // A grid is sent back as written; its drums id ("grid") is not in the catalogue (BL-123).
  if (rhythm?.drumGrid) req.drumGrid = rhythm.drumGrid;
  else if (rhythm?.drums) req.drums = rhythm.drums.id;
  if (rhythm?.feel) req.feel = rhythm.feel.id;
  if (rhythm?.style) req.style = rhythm.style.id; // keeps the style's instrument (BL-110)
  return req;
}

type Style = PatternsResponse['styles'][number];
export type Groove = Style['grooves'][number];

/** The style's groove in the timeline's meter (BL-087); undefined when the style lists none. */
export function grooveFor(style: Style, timeline: Timeline): Groove | undefined {
  const ts = timeline.timeSignature;
  return style.grooves.find(
    (g) => g.meter.numerator === ts.numerator && g.meter.denominator === ts.denominator,
  );
}

/**
 * The request that plays the same harmony in another style (BL-087): the
 * numerals, lengths, key, tempo and meter stay, and the groove is the one
 * the engine lists for the style in this meter.
 */
export function regrooveRequest(
  timeline: Timeline,
  style: string,
  groove: Groove,
): FromProgressionRequest {
  const req = sameHarmony(timeline, timeline.key.tonic);
  req.style = style;
  req.accompaniment = groove.accompaniment;
  req.feel = groove.feel;
  if (groove.drums) req.drums = groove.drums;
  return req;
}

/** The timeline's harmony as a request in `tonic`, without patterns. */
function sameHarmony(timeline: Timeline, tonic: PitchClass): FromProgressionRequest {
  const beat = ticksPerBeat(timeline.ppq, timeline.timeSignature);
  return {
    key: { tonic, mode: timeline.key.mode },
    tempoBpm: timeline.tempoBpm,
    timeSignature: timeline.timeSignature,
    progression: timeline.harmony.map((h) => ({
      numeral: h.romanNumeral,
      beats: Math.round(h.duration / beat),
      ...(h.push ? { push: h.push } : {}), // a pushed chord stays pushed (BL-085)
    })),
  };
}
