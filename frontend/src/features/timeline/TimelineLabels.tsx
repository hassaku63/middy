import { pitchName } from '@/shared/lib/notation';
import { DRUM_ROW_HEIGHT, ROW_HEIGHT, drumName, type Layout } from './layout';

type Props = { layout: Layout };

/**
 * The label column, drawn again over the timeline's own and held at the
 * left of the scrolling box (#402 A): on a narrow screen the drawing keeps
 * its size and scrolls sideways, and the pitch and drum names stay in
 * view. The timeline draws the same names, so this copy is hidden from
 * assistive tech.
 */
export function TimelineLabels({ layout }: Props) {
  const rows = Array.from({ length: layout.rows }, (_, i) => layout.hi - i);
  return (
    <svg
      className="timeline-labels"
      aria-hidden="true"
      width={layout.labelWidth}
      height={layout.height}
      viewBox={`0 0 ${layout.labelWidth} ${layout.height}`}
    >
      <rect className="timeline-labels__ground" width={layout.labelWidth} height={layout.height} />
      {rows
        .filter((pitch) => pitch % 12 === 0)
        .map((pitch) => (
          <text
            key={pitch}
            x={4}
            y={layout.y(pitch) + ROW_HEIGHT - 2}
            className="timeline__pitch-label"
          >
            {pitchName(pitch)}
          </text>
        ))}
      {layout.drumNotes.map((note, i) => (
        <text
          key={`drum-${note}`}
          x={4}
          y={layout.drumTop + i * DRUM_ROW_HEIGHT + DRUM_ROW_HEIGHT - 4}
          className="timeline__pitch-label"
        >
          {drumName(note)}
        </text>
      ))}
    </svg>
  );
}
