import type { ReactNode } from 'react';
import type { NoteEvent } from '@/shared/api/client';
import { DRUM_ROW_HEIGHT, drumName, type Layout } from './layout';

type Props = {
  notes: NoteEvent[];
  layout: Layout;
  positionTick: number;
  /** Drawn over the rows and under the notes (TimelineGrid). */
  grid?: ReactNode;
};

/** One row per percussion sound; hits are short ticks so the groove reads at a glance. */
export function DrumLane({ notes, layout, positionTick, grid }: Props) {
  if (layout.drumNotes.length === 0) return null;
  const rows = new Map(layout.drumNotes.map((note, i) => [note, i])); // not indexOf per hit
  const rowY = (note: number) => layout.drumTop + (rows.get(note) ?? 0) * DRUM_ROW_HEIGHT;
  return (
    <g aria-label="ドラム" data-testid="drum-lane">
      {layout.drumNotes.map((note) => (
        <g key={`drum-row-${note}`}>
          <rect
            className="timeline__drum-row"
            x={layout.labelWidth}
            y={rowY(note)}
            width={layout.width - layout.labelWidth}
            height={DRUM_ROW_HEIGHT}
          />
          <text x={4} y={rowY(note) + DRUM_ROW_HEIGHT - 4} className="timeline__pitch-label">
            {drumName(note)}
          </text>
        </g>
      ))}
      {grid}
      {notes.map((n, i) => {
        const sounding = positionTick >= n.start && positionTick < n.start + n.duration;
        return (
          <rect
            key={`drum-${i}`}
            className={`timeline__drum-hit${sounding ? ' is-sounding' : ''}`}
            x={layout.x(n.start)}
            y={rowY(n.pitch) + 2}
            width={4}
            height={DRUM_ROW_HEIGHT - 4}
            rx={1}
          />
        );
      })}
    </g>
  );
}
