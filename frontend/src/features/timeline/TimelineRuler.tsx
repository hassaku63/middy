import type { Layout } from './layout';
import { RULER_HEIGHT } from './layout';

/**
 * Bar numbers above the chord lane, as a DAW's ruler (#435): bar 3 here is
 * the 3 of 3.1 the player shows. A picture: the player says the position.
 */
export function TimelineRuler({ layout }: { layout: Layout }) {
  return (
    <g aria-hidden="true" className="timeline__ruler">
      <rect
        className="timeline__ruler-ground"
        x={layout.labelWidth}
        y={0}
        width={layout.width - layout.labelWidth}
        height={RULER_HEIGHT}
      />
      {Array.from({ length: layout.bars + 1 }, (_, i) => (
        <line
          key={`bar-${i}`}
          className="timeline__bar-line"
          x1={layout.x(i * layout.perBar)}
          x2={layout.x(i * layout.perBar)}
          y1={0}
          y2={RULER_HEIGHT}
        />
      ))}
      {Array.from({ length: layout.bars }, (_, i) => (
        <text
          key={i}
          className="timeline__bar-number"
          x={layout.x(i * layout.perBar) + 4}
          y={RULER_HEIGHT - 5}
        >
          {i + 1}
        </text>
      ))}
    </g>
  );
}
