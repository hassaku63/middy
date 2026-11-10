import { useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import { TimelineView, type TimelineViewProps } from './TimelineView';
import { BAR_WIDTH, LABEL_WIDTH, computeLayout, type Layout } from './layout';
import './TimelineOverview.css';

/** Bars in the detail below the overview. */
const BARS_PER_VIEW = 4;
const DETAIL_WIDTH = BARS_PER_VIEW * BAR_WIDTH;

type Props = Omit<TimelineViewProps, 'view' | 'highlight' | 'label' | 'decorative' | 'className'>;

/**
 * The timeline as the screen shows it (BL-080): a short one whole, a long
 * one as the whole piece small on top and four bars large below. Clicking
 * a chord in the overview pins it and so moves the detail. The overview is
 * a duplicate drawing, hidden from assistive tech. The detail follows the
 * playhead, keeping it at the centre while playing; stopped, it centres the
 * pinned chord, else where a slider puts it (BL-127, the owner's choice).
 */
export function TimelineOverview(props: Props) {
  const [manual, setManual] = useState(0); // the detail's centre while stopped (B), in ticks
  const geometry = computeLayout(props.timeline);
  if (geometry.bars <= BARS_PER_VIEW) return <TimelineView {...props} />;
  const pinned =
    props.selectedHarmony != null ? props.timeline.harmony[props.selectedHarmony] : undefined;
  const detail = centredOn(geometry, followed(props, pinned, manual));
  const sliding = props.playing === false && !pinned;
  return (
    <div className="timeline-overview">
      <TimelineView {...props} decorative highlight={detail} className="timeline--overview" />
      <TimelineView {...props} view={detail} />
      {sliding && (
        <input
          type="range"
          className="timeline-overview__slider"
          aria-label="拡大の位置"
          min={0}
          max={geometry.bars * geometry.perBar}
          step={geometry.perBar / 4}
          value={manual}
          onChange={(e) => setManual(Number(e.target.value))}
        />
      )}
    </div>
  );
}

/** The tick kept at the centre of the detail. */
function followed(
  props: { positionTick: number; playing?: boolean | undefined },
  pinned: Timeline['harmony'][number] | undefined,
  manual: number,
) {
  if (props.playing !== false) return props.positionTick;
  return pinned ? pinned.start + pinned.duration / 2 : manual;
}

/** Four bars with tick at the centre, kept inside the piece. */
function centredOn(geometry: Layout, tick: number) {
  const x = geometry.x(tick) - DETAIL_WIDTH / 2;
  return {
    x: Math.min(Math.max(x, LABEL_WIDTH), geometry.width - DETAIL_WIDTH),
    width: DETAIL_WIDTH,
  };
}
