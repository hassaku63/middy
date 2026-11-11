import type { Layout } from './layout';

type Props = {
  layout: Layout;
  /** The lane's top and bottom, in drawing units. */
  top: number;
  bottom: number;
  /** Draw the beats as well as the bars (not on a thin overview strip). */
  beats?: boolean;
};

/**
 * The bars and beats across one lane, drawn over its rows and under its
 * notes, as a DAW's editor draws its grid (#435): a bar's line stronger
 * than a beat's, so where a bar starts shows apart from where a chord does.
 */
export function TimelineGrid({ layout, top, bottom, beats = true }: Props) {
  const perBar = Math.round(layout.perBar / layout.perBeat);
  return (
    <g aria-hidden="true" className="timeline__grid">
      {beats &&
        Array.from({ length: layout.bars * perBar }, (_, i) =>
          i % perBar === 0 ? null : (
            <line
              key={`beat-${i}`}
              className="timeline__beat"
              x1={layout.x(i * layout.perBeat)}
              x2={layout.x(i * layout.perBeat)}
              y1={top}
              y2={bottom}
            />
          ),
        )}
      {Array.from({ length: layout.bars + 1 }, (_, i) => (
        <line
          key={`bar-${i}`}
          className="timeline__bar-line"
          x1={layout.x(i * layout.perBar)}
          x2={layout.x(i * layout.perBar)}
          y1={top}
          y2={bottom}
        />
      ))}
    </g>
  );
}
