import type { HarmonyEvent } from '@/shared/api/client';
import { chordSymbol } from '@/shared/lib/notation';
import { CHORD_LANE_HEIGHT, LABEL_WIDTH, type Layout } from './layout';

type Props = {
  harmony: HarmonyEvent[];
  layout: Layout;
  activeIndex: number;
  selectedIndex: number | null;
  onSelect?: ((index: number) => void) | undefined;
};

/** One clickable block per harmony event, with roman numeral and chord symbol. */
export function ChordLane({ harmony, layout, activeIndex, selectedIndex, onSelect }: Props) {
  return (
    <g aria-label="コード" role="listbox">
      {harmony.map((h, i) => {
        const active = i === activeIndex;
        const selected = i === selectedIndex;
        const x = layout.x(h.start);
        return (
          <g
            key={i}
            role="option"
            aria-label={`${h.romanNumeral} (${chordSymbol(h.chord)})`}
            aria-current={active ? 'true' : undefined}
            aria-selected={selected}
            className={`timeline__chord${active ? ' is-active' : ''}${selected ? ' is-selected' : ''}`}
            onClick={() => onSelect?.(i)}
            tabIndex={onSelect ? 0 : undefined}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') onSelect?.(i);
            }}
          >
            <rect
              x={x}
              y={2}
              width={layout.x(h.duration) - LABEL_WIDTH}
              height={CHORD_LANE_HEIGHT - 4}
              rx={4}
            />
            <text x={x + 6} y={18} className="timeline__numeral">
              {h.romanNumeral}
            </text>
            <text x={x + 6} y={36} className="timeline__symbol">
              {chordSymbol(h.chord)}
            </text>
          </g>
        );
      })}
    </g>
  );
}
