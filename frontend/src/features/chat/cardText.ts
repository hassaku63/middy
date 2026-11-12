import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { keyName } from '@/shared/lib/notation';
import { barsOf, chordNamesOf, numeralsOf } from './progression';

/** The card's spec items: key, tempo, length and bars ("A minor", "72 BPM", …). */
export function specOf(c: IdeaCandidate): string[] {
  const bars = barsOf(c);
  return [
    keyName(c.key),
    `${c.tempoBpm} BPM`,
    c.beats?.length ? `長さ: ${c.beats.join('・')} 拍` : `${c.beatsPerChord} 拍/コード`,
    ...(bars !== undefined ? [`${bars} 小節`] : []),
  ];
}

/**
 * Names for the candidate's patterns, from the catalogue when it is loaded.
 * Straight feel and 4/4 are the defaults and go unsaid.
 */
export function patternLabels(c: IdeaCandidate, patterns: PatternsResponse | undefined): string[] {
  const name = (list: { id: string; name: string }[] | undefined, id: string) =>
    list?.find((x) => x.id === id)?.name ?? id;
  const labels: string[] = [];
  if (c.style) labels.push(`スタイル: ${name(patterns?.styles, c.style)}`);
  if (c.accompaniment) labels.push(`伴奏: ${name(patterns?.accompaniments, c.accompaniment)}`);
  if (c.drums) labels.push(`ドラム: ${name(patterns?.drums, c.drums)}`);
  if (c.feel && c.feel !== 'straight') labels.push(`ノリ: ${name(patterns?.feels, c.feel)}`);
  if (c.meter && c.meter !== '4/4') labels.push(`拍子: ${c.meter}`);
  return labels;
}

/**
 * The card as lines of text (BL-298, BL-300): the words it answered, its
 * spec, progression, chord names, patterns and reason. The explanation
 * shows them, and copying the explanation writes them.
 */
export function cardLines(
  c: IdeaCandidate,
  image: string,
  patterns: PatternsResponse | undefined,
): string[] {
  const names = chordNamesOf(c);
  const labels = patternLabels(c, patterns);
  return [
    ...(image ? [`イメージ: ${image}`] : []),
    specOf(c).join(' · '),
    numeralsOf(c).join(' - '),
    ...(names.length > 0 ? [names.join(' - ')] : []),
    ...(labels.length > 0 ? [labels.join(' · ')] : []),
    c.reason,
  ];
}
