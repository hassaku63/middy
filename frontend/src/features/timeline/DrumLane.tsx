import type { NoteEvent } from '@/shared/api/client';
import { DRUM_ROW_HEIGHT, LABEL_WIDTH, type Layout } from './layout';

type Props = {
  notes: NoteEvent[];
  layout: Layout;
  positionTick: number;
};

const DRUM_NAMES: Record<number, string> = {
  35: 'Kick',
  36: 'Kick',
  37: 'Stick',
  38: 'Snare',
  40: 'Snare',
  42: 'HH',
  44: 'HH',
  45: 'Tom L',
  46: 'HH open',
  47: 'Tom M',
  49: 'Crash',
  50: 'Tom H',
  51: 'Ride',
  53: 'Bell',
};

/** One row per percussion sound; hits are short ticks so the groove reads at a glance. */
export function DrumLane({ notes, layout, positionTick }: Props) {
  if (layout.drumNotes.length === 0) return null;
  const rows = new Map(layout.drumNotes.map((note, i) => [note, i])); // not indexOf per hit
  const rowY = (note: number) => layout.drumTop + (rows.get(note) ?? 0) * DRUM_ROW_HEIGHT;
  return (
    <g aria-label="ドラム" data-testid="drum-lane">
      {layout.drumNotes.map((note) => (
        <g key={`drum-row-${note}`}>
          <rect
            className="timeline__drum-row"
            x={LABEL_WIDTH}
            y={rowY(note)}
            width={layout.width - LABEL_WIDTH}
            height={DRUM_ROW_HEIGHT}
          />
          <text x={4} y={rowY(note) + DRUM_ROW_HEIGHT - 3} className="timeline__pitch-label">
            {DRUM_NAMES[note] ?? `#${note}`}
          </text>
        </g>
      ))}
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
