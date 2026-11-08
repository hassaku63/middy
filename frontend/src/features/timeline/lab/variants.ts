/**
 * Long-timeline layouts under trial (BL-080). The owner compares them with
 * ?timeline=<id> (and ?lab to show a switcher); the ones not adopted are
 * deleted together with this folder's files that only they use.
 */
export const TIMELINE_LAYOUTS = [
  { id: 'fit', name: '縮める（今の表示）' },
  { id: 'scroll', name: '横スクロール＋追従' },
  { id: 'rows', name: '4 小節ずつ折り返し' },
  { id: 'overview', name: '全体図＋拡大' },
] as const;

export type TimelineLayout = (typeof TIMELINE_LAYOUTS)[number]['id'];

export function isTimelineLayout(v: string | null): v is TimelineLayout {
  return TIMELINE_LAYOUTS.some((l) => l.id === v);
}

/** Bars shown per row (rows) and in the detail window (overview). */
export const BARS_PER_VIEW = 4;
