import type { FromProgressionRequest, PitchClass, Timeline } from '@/shared/api/client';

/**
 * The request that renders the same example in another key (BL-078): the
 * timeline's own numerals, chord lengths, tempo, meter, mode and patterns,
 * with a new tonic. The engine guarantees the numerals read back as the same
 * chords (music.TestHarmonyLabelsReadBackAsTheSameChords).
 */
export function rebuildRequest(timeline: Timeline, tonic: PitchClass): FromProgressionRequest {
  const req: FromProgressionRequest = {
    key: { tonic, mode: timeline.key.mode },
    tempoBpm: timeline.tempoBpm,
    timeSignature: timeline.timeSignature,
    progression: timeline.harmony.map((h) => ({
      numeral: h.romanNumeral,
      beats: Math.round(h.duration / timeline.ppq),
    })),
  };
  const rhythm = timeline.rhythm;
  if (rhythm?.accompaniment) req.accompaniment = rhythm.accompaniment.id;
  if (rhythm?.drums) req.drums = rhythm.drums.id;
  if (rhythm?.feel) req.feel = rhythm.feel.id;
  return req;
}
