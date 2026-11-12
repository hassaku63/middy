import type { Rhythm, Timeline } from '@/shared/api/client';
import { harmonySymbol } from '@/shared/lib/notation';

/**
 * What colours the chords: the style's instrument and tensions, or the
 * tensions the candidate asked for (ADR-0037).
 */
export function colourNotes(rhythm: Rhythm): string[] {
  const style = rhythm.style;
  const out: string[] = [];
  if (style?.instrument)
    out.push(`${style.name}のスタイルなので、和音は${style.instrument}で鳴らす。`);
  if (style?.tensions) out.push(style.tensions);
  if (rhythm.tensions?.explanation) out.push(rhythm.tensions.explanation);
  return out;
}

/**
 * The engine's notes on the whole example as lines of text, for copying
 * (BL-300, #442): the chords' names in order, the whole and the flow, and
 * the rhythm. The one-chord notes follow the playhead and are left out.
 */
export function explanationLines(timeline: Timeline): string[] {
  const lines: string[] = [];
  if (timeline.harmony.length > 0) {
    lines.push(`コード: ${timeline.harmony.map(harmonySymbol).join(' - ')}`);
  }
  if (timeline.overview) {
    lines.push(`全体: ${timeline.overview.whole}`, `流れ: ${timeline.overview.flow}`);
  }
  const rhythm = timeline.rhythm;
  if (rhythm) {
    lines.push(...colourNotes(rhythm));
    for (const [label, part] of [
      ['伴奏', rhythm.accompaniment],
      ['ドラム', rhythm.drums],
    ] as const) {
      if (part) lines.push(`${label}: ${part.name}。${part.description}${part.explanation}`);
    }
    if (rhythm.feel) lines.push(`ノリ: ${rhythm.feel.name}。${rhythm.feel.explanation}`);
  }
  return lines;
}
