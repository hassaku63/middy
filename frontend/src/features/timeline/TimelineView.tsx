import type { Timeline } from '@/shared/api/client';
import { chordSymbol, pitchName } from '@/shared/lib/notation';
import { ticksPerBar, timelineLengthTicks } from '@/shared/lib/timing';
import { BAR_WIDTH, harmonyAt } from './layout';
import './TimelineView.css';

export type TimelineViewProps = {
  timeline: Timeline;
  /** Playhead position in ticks. */
  positionTick: number;
  /** Index into timeline.harmony the user pinned by clicking, if any. */
  selectedHarmony?: number | null;
  onSelectHarmony?: (index: number) => void;
};

const CHORD_LANE_HEIGHT = 44;
const ROW_HEIGHT = 10;
const LABEL_WIDTH = 40;
const MIN_ROWS = 12;

/**
 * Chord lane plus piano roll rendered as SVG. Highlights the chord and the
 * notes under the playhead; chord blocks are clickable so the explanation
 * panel can be pinned to one of them.
 */
export function TimelineView({
  timeline,
  positionTick,
  selectedHarmony = null,
  onSelectHarmony,
}: TimelineViewProps) {
  const { ppq, timeSignature, harmony, tracks } = timeline;
  const perBar = ticksPerBar(ppq, timeSignature);
  const lengthTicks = Math.max(timelineLengthTicks(timeline), perBar);
  const bars = Math.ceil(lengthTicks / perBar);
  const x = (tick: number) => LABEL_WIDTH + (tick / perBar) * BAR_WIDTH;
  const width = x(bars * perBar);

  const notes = tracks.flatMap((t) => t.notes);
  const pitches = notes.map((n) => n.pitch);
  const hi = pitches.length ? Math.max(...pitches) : 71;
  const lo = Math.min(pitches.length ? Math.min(...pitches) : 60, hi - MIN_ROWS + 1);
  const rows = hi - lo + 1;
  const rollTop = CHORD_LANE_HEIGHT;
  const y = (pitch: number) => rollTop + (hi - pitch) * ROW_HEIGHT;
  const height = rollTop + rows * ROW_HEIGHT;

  const activeHarmony = harmonyAt(harmony, positionTick);

  return (
    <svg
      className="timeline"
      role="img"
      aria-label="タイムライン"
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
    >
      {/* bar lines */}
      {Array.from({ length: bars + 1 }, (_, i) => (
        <line
          key={`bar-${i}`}
          className="timeline__bar"
          x1={x(i * perBar)}
          x2={x(i * perBar)}
          y1={0}
          y2={height}
        />
      ))}

      {/* chord lane */}
      <g aria-label="コード" role="listbox">
        {harmony.map((h, i) => {
          const active = i === activeHarmony;
          const selected = i === selectedHarmony;
          const label = `${h.romanNumeral} (${chordSymbol(h.chord)})`;
          return (
            <g
              key={i}
              role="option"
              aria-label={label}
              aria-current={active ? 'true' : undefined}
              aria-selected={selected}
              className={`timeline__chord${active ? ' is-active' : ''}${selected ? ' is-selected' : ''}`}
              onClick={() => onSelectHarmony?.(i)}
              tabIndex={onSelectHarmony ? 0 : undefined}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') onSelectHarmony?.(i);
              }}
            >
              <rect
                x={x(h.start)}
                y={2}
                width={x(h.duration) - LABEL_WIDTH}
                height={CHORD_LANE_HEIGHT - 4}
                rx={4}
              />
              <text x={x(h.start) + 6} y={18} className="timeline__numeral">
                {h.romanNumeral}
              </text>
              <text x={x(h.start) + 6} y={36} className="timeline__symbol">
                {chordSymbol(h.chord)}
              </text>
            </g>
          );
        })}
      </g>

      {/* piano roll */}
      <g aria-label="ピアノロール">
        {Array.from({ length: rows }, (_, i) => {
          const pitch = hi - i;
          const black = [1, 3, 6, 8, 10].includes(pitch % 12);
          return (
            <g key={`row-${pitch}`}>
              <rect
                className={`timeline__row${black ? ' is-black' : ''}`}
                x={LABEL_WIDTH}
                y={y(pitch)}
                width={width - LABEL_WIDTH}
                height={ROW_HEIGHT}
              />
              {pitch % 12 === 0 && (
                <text x={4} y={y(pitch) + ROW_HEIGHT - 2} className="timeline__pitch-label">
                  {pitchName(pitch)}
                </text>
              )}
            </g>
          );
        })}
        {notes.map((n, i) => {
          const sounding = positionTick >= n.start && positionTick < n.start + n.duration;
          return (
            <rect
              key={`note-${i}`}
              className={`timeline__note${sounding ? ' is-sounding' : ''}`}
              data-sounding={sounding ? 'true' : undefined}
              aria-label={`${pitchName(n.pitch)}`}
              x={x(n.start)}
              y={y(n.pitch) + 1}
              width={Math.max(2, x(n.duration) - LABEL_WIDTH - 1)}
              height={ROW_HEIGHT - 2}
              rx={2}
            />
          );
        })}
      </g>

      {/* playhead */}
      <line
        className="timeline__playhead"
        data-testid="playhead"
        x1={x(positionTick)}
        x2={x(positionTick)}
        y1={0}
        y2={height}
      />
    </svg>
  );
}
