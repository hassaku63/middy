import type { Rhythm } from '@/shared/api/client';

type Part = { label: string; name: (r: Rhythm | undefined) => string };

// What each part is when the timeline does not name it.
const PARTS: Part[] = [
  { label: '伴奏', name: (r) => r?.accompaniment?.name ?? 'ブロック' },
  { label: 'ドラム', name: (r) => r?.drums?.name ?? 'なし' },
  { label: 'ノリ', name: (r) => r?.feel?.name ?? 'ストレート' },
  // A style may voice the chords on its own instrument (BL-110).
  { label: '楽器', name: (r) => r?.style?.instrument ?? '型の楽器' },
];

/**
 * What changed when the example was played in another style (BL-087): the
 * harmony stays, so only the parts that differ are named, before → after.
 */
export function grooveChanges(
  styleName: string,
  before: Rhythm | undefined,
  after: Rhythm | undefined,
): string {
  const changed = PARTS.filter((p) => p.name(before) !== p.name(after)).map(
    (p) => `${p.label}: ${p.name(before)} → ${p.name(after)}`,
  );
  const head = `${styleName}で鳴らし直しました。和声はそのままです。`;
  return changed.length === 0
    ? `${head}伴奏・ドラム・ノリ・楽器も前と同じです。`
    : `${head}${changed.join('、')}。`;
}
