import type { ReactNode } from 'react';
import type { NoteEvent } from '@/shared/api/client';
import { pitchName } from '@/shared/lib/notation';
import { ROW_HEIGHT, type Layout } from './layout';

type Props = {
  notes: NoteEvent[];
  layout: Layout;
  positionTick: number;
  /** Drawn over the rows and under the notes (TimelineGrid). */
  grid?: ReactNode;
};

const BLACK_KEYS = new Set([1, 3, 6, 8, 10]);

/**
 * Pitch rows (with C labels) and one rectangle per note; sounding notes are
 * marked. A picture only: assistive technology gets the chords (ChordLane)
 * and the explanation instead of dozens of note names (BL-162).
 */
export function PianoRoll({ notes, layout, positionTick, grid }: Props) {
  return (
    <g aria-hidden="true">
      {Array.from({ length: layout.rows }, (_, i) => {
        const pitch = layout.hi - i;
        return (
          <g key={`row-${pitch}`}>
            <rect
              className={`timeline__row${BLACK_KEYS.has(pitch % 12) ? ' is-black' : ''}`}
              x={layout.labelWidth}
              y={layout.y(pitch)}
              width={layout.width - layout.labelWidth}
              height={ROW_HEIGHT}
            />
            {pitch % 12 === 0 && (
              <text x={4} y={layout.y(pitch) + ROW_HEIGHT - 2} className="timeline__pitch-label">
                {pitchName(pitch)}
              </text>
            )}
          </g>
        );
      })}
      {grid}
      {notes.map((n, i) => {
        const sounding = positionTick >= n.start && positionTick < n.start + n.duration;
        return (
          <rect
            key={`note-${i}`}
            className={`timeline__note${sounding ? ' is-sounding' : ''}`}
            data-sounding={sounding ? 'true' : undefined}
            data-pitch={pitchName(n.pitch)}
            x={layout.x(n.start)}
            y={layout.y(n.pitch) + 1}
            width={Math.max(2, layout.x(n.duration) - layout.labelWidth - 1)}
            height={ROW_HEIGHT - 2}
            rx={2}
          />
        );
      })}
    </g>
  );
}
