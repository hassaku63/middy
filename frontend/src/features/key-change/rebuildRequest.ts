import type { FromProgressionRequest, PitchClass, Timeline } from '@/shared/api/client';
import { ticksPerBeat } from '@/shared/lib/timing';

/**
 * The request that renders the same example in another key (BL-078): the
 * timeline's own numerals, chord lengths, tempo, meter, mode and patterns,
 * with a new tonic. The engine guarantees the numerals read back as the same
 * chords (music.TestHarmonyLabelsReadBackAsTheSameChords).
 */
export function rebuildRequest(timeline: Timeline, tonic: PitchClass): FromProgressionRequest {
  const beat = ticksPerBeat(timeline.ppq, timeline.timeSignature);
  const req: FromProgressionRequest = {
    key: { tonic, mode: timeline.key.mode },
    tempoBpm: timeline.tempoBpm,
    timeSignature: timeline.timeSignature,
    progression: timeline.harmony.map((h) => ({
      numeral: h.romanNumeral,
      beats: Math.round(h.duration / beat),
    })),
  };
  const rhythm = timeline.rhythm;
  if (rhythm?.accompaniment) req.accompaniment = rhythm.accompaniment.id;
  if (rhythm?.drums) req.drums = rhythm.drums.id;
  if (rhythm?.feel) req.feel = rhythm.feel.id;
  if (rhythm?.style) req.style = rhythm.style.id; // keeps the style's instrument (BL-110)
  return req;
}
