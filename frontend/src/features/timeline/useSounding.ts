import { useMemo } from 'react';
import type { NoteEvent } from '@/shared/api/client';

/**
 * The indexes of the notes sounding at tick, as a set that stays the same
 * object until a note starts or stops (#500): the note layers redraw at
 * the notes' edges, not on every move of the playhead.
 */
export function useSounding(notes: NoteEvent[], tick: number): ReadonlySet<number> {
  let key = '';
  notes.forEach((n, i) => {
    if (tick >= n.start && tick < n.start + n.duration) key += `${String(i)},`;
  });
  return useMemo(() => new Set(key.split(',').filter(Boolean).map(Number)), [key]);
}
