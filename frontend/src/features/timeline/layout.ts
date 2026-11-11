import type { HarmonyEvent, Timeline } from '@/shared/api/client';
import { ticksPerBar, timelineLengthTicks } from '@/shared/lib/timing';

/** Pixels per bar; every lane shares this scale so columns line up. */
export const BAR_WIDTH = 160;
export const CHORD_LANE_HEIGHT = 44;
export const ROW_HEIGHT = 10;
/** The narrowest label column: pitch names such as C#4. */
export const LABEL_WIDTH = 40;
/** The drum rows' labels: 9 px text, one em per character. */
const DRUM_LABEL_FONT = 9;
export const DRUM_ROW_HEIGHT = 12;
const MIN_ROWS = 12;
/** GM percussion channel (0-based); its notes are drawn in the drum lane. */
export const DRUM_CHANNEL = 9;

/**
 * The drum rows' names, in the words of the rhythm notes (#402 E,
 * backend/internal/music/pattern/nearest.go), so the lane and the text
 * name a drum alike.
 */
export const DRUM_NAMES: Record<number, string> = {
  35: 'キック',
  36: 'キック',
  37: 'サイドスティック',
  38: 'スネア',
  40: 'スネア',
  42: 'ハイハット',
  44: '足のハイハット', // the pedal hi-hat, told from the hands' closed hat (GM 42, #385)
  45: '低いタム',
  46: 'オープンのハイハット',
  47: '中くらいのタム',
  49: 'クラッシュ',
  50: '高いタム',
  51: 'ライド',
  53: 'ライドのベル',
};

/** A drum row's label. */
export function drumName(note: number): string {
  return DRUM_NAMES[note] ?? `#${note}`;
}

/** Index of the harmony event sounding at `tick`, or -1. */
export function harmonyAt(harmony: HarmonyEvent[], tick: number): number {
  return harmony.findIndex((h) => tick >= h.start && tick < h.start + h.duration);
}

export type Layout = {
  /** The label column: wide enough for the longest drum name in it. */
  labelWidth: number;
  /** Pixels per bar in the drawing: BAR_WIDTH, or wider to fill a box (fitBarWidth). */
  barWidth: number;
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

/**
 * Geometry shared by the chord lane, the piano roll and the playhead. A
 * bar is BAR_WIDTH wide unless a wider one is asked for, to fill a box
 * whose height is capped (fitBarWidth, #402 F); only time stretches, the
 * rows and the text keep their size.
 */
export function computeLayout(tl: Timeline, barWidth = BAR_WIDTH): Layout {
  const perBar = ticksPerBar(tl.ppq, tl.timeSignature);
  const lengthTicks = Math.max(timelineLengthTicks(tl), perBar);
  const bars = Math.ceil(lengthTicks / perBar);

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
  const labelWidth = Math.max(
    LABEL_WIDTH,
    ...drumNotes.map((n) => 4 + drumName(n).length * DRUM_LABEL_FONT + 4),
  );
  const x = (tick: number) => labelWidth + (tick / perBar) * barWidth;
  const hi = pitches.length ? Math.max(...pitches) : 71;
  const lo = Math.min(pitches.length ? Math.min(...pitches) : 60, hi - MIN_ROWS + 1);
  const rows = hi - lo + 1;
  const rollTop = CHORD_LANE_HEIGHT;
  const y = (pitch: number) => rollTop + (hi - pitch) * ROW_HEIGHT;
  const drumTop = rollTop + rows * ROW_HEIGHT + (drumNotes.length ? 4 : 0);

  return {
    labelWidth,
    barWidth,
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

/**
 * The bar width that fills a box of the given size at the drawing's own
 * height (#402 F): where the box is wider than the drawing at BAR_WIDTH,
 * the bars stretch to its width instead of the drawing shrinking to fit
 * the capped height and leaving the right of the box empty.
 */
export function fitBarWidth(natural: Layout, box: { width: number; height: number }): number {
  if (box.width <= 0 || box.height <= 0) return BAR_WIDTH;
  const width = (box.width / box.height) * natural.height; // drawing units across the box
  return Math.max(BAR_WIDTH, (width - natural.labelWidth) / natural.bars);
}
