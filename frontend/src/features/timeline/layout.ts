import type { HarmonyEvent, Timeline } from '@/shared/api/client';
import { ticksPerBar, timelineLengthTicks } from '@/shared/lib/timing';

/** Pixels per bar; every lane shares this scale so columns line up. */
export const BAR_WIDTH = 160;
export const CHORD_LANE_HEIGHT = 44;
export const ROW_HEIGHT = 10;
export const LABEL_WIDTH = 40;
export const DRUM_ROW_HEIGHT = 12;
const MIN_ROWS = 12;
/** GM percussion channel (0-based); its notes are drawn in the drum lane. */
export const DRUM_CHANNEL = 9;

/** Index of the harmony event sounding at `tick`, or -1. */
export function harmonyAt(harmony: HarmonyEvent[], tick: number): number {
  return harmony.findIndex((h) => tick >= h.start && tick < h.start + h.duration);
}

export type Layout = {
  perBar: number;
  bars: number;
  width: number;
  height: number;
  /** distinct drum notes, highest first; empty when there is no drum track */
  drumNotes: number[];
  drumTop: number;
  /** highest and lowest pitch row shown */
  hi: number;
  lo: number;
  rows: number;
  rollTop: number;
  x: (tick: number) => number;
  y: (pitch: number) => number;
};

/** Geometry shared by the chord lane, the piano roll and the playhead. */
export function computeLayout(tl: Timeline): Layout {
  const perBar = ticksPerBar(tl.ppq, tl.timeSignature);
  const lengthTicks = Math.max(timelineLengthTicks(tl), perBar);
  const bars = Math.ceil(lengthTicks / perBar);
  const x = (tick: number) => LABEL_WIDTH + (tick / perBar) * BAR_WIDTH;

  const pitches = tl.tracks
    .filter((t) => t.channel !== DRUM_CHANNEL)
    .flatMap((t) => t.notes.map((n) => n.pitch));
  const drumNotes = [
    ...new Set(
      tl.tracks
        .filter((t) => t.channel === DRUM_CHANNEL)
        .flatMap((t) => t.notes.map((n) => n.pitch)),
    ),
  ].sort((a, b) => b - a);
  const hi = pitches.length ? Math.max(...pitches) : 71;
  const lo = Math.min(pitches.length ? Math.min(...pitches) : 60, hi - MIN_ROWS + 1);
  const rows = hi - lo + 1;
  const rollTop = CHORD_LANE_HEIGHT;
  const y = (pitch: number) => rollTop + (hi - pitch) * ROW_HEIGHT;
  const drumTop = rollTop + rows * ROW_HEIGHT + (drumNotes.length ? 4 : 0);

  return {
    perBar,
    bars,
    width: x(bars * perBar),
    height: drumTop + drumNotes.length * DRUM_ROW_HEIGHT,
    drumNotes,
    drumTop,
    hi,
    lo,
    rows,
    rollTop,
    x,
    y,
  };
}
