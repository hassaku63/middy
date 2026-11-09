import type { HarmonyEvent } from '@/shared/api/client';
import { harmonySymbol } from '@/shared/lib/notation';
import { CHORD_LANE_HEIGHT, LABEL_WIDTH, type Layout } from './layout';

type Props = {
  harmony: HarmonyEvent[];
  layout: Layout;
  activeIndex: number;
  selectedIndex: number | null;
  onSelect?: ((index: number) => void) | undefined;
  /** False for a duplicate drawing hidden from assistive tech: clickable, but no tab stop. */
  focusable?: boolean;
  /**
   * Mark the chord the explanation follows with ▶ and the pinned one with ■,
   * so pinning shows without colour (WCAG 1.4.1; the desk under trial, BL-185).
   */
  marks?: boolean | undefined;
};

/** One clickable block per harmony event, with roman numeral and chord symbol. */
export function ChordLane({
  harmony,
  layout,
  activeIndex,
  selectedIndex,
  onSelect,
  focusable = true,
  marks = false,
}: Props) {
  return (
    <g aria-label="コード" role="listbox">
      {harmony.map((h, i) => {
        const active = i === activeIndex;
        const selected = i === selectedIndex;
        const x = layout.x(h.start);
        const width = layout.x(h.duration) - LABEL_WIDTH;
        const mark = selected ? '■' : active && selectedIndex === null ? '▶' : null;
        return (
          <g
            key={i}
            role="option"
            aria-label={`${h.romanNumeral} ${harmonySymbol(h)}`}
            aria-current={active ? 'true' : undefined}
            aria-selected={selected}
            className={`timeline__chord${active ? ' is-active' : ''}${selected ? ' is-selected' : ''}`}
            onClick={() => onSelect?.(i)}
            tabIndex={onSelect && focusable ? 0 : undefined}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault(); // Space would also scroll the page
                onSelect?.(i);
              }
            }}
          >
            <rect x={x} y={2} width={width} height={CHORD_LANE_HEIGHT - 4} rx={4} />
            {/* One text with a space between the lines, so the visible text
                reads "V7 B7" like the option's name (WCAG 2.5.3, BL-162). */}
            <text>
              <tspan x={x + 6} y={18} className="timeline__numeral">
                {h.romanNumeral}
              </tspan>{' '}
              <tspan x={x + 6} y={36} className="timeline__symbol">
                {harmonySymbol(h)}
              </tspan>
            </text>
            {marks && mark && (
              <text
                x={x + width - 6}
                y={18}
                textAnchor="end"
                className="timeline__mark"
                aria-hidden="true"
              >
                {mark}
              </text>
            )}
          </g>
        );
      })}
    </g>
  );
}
