import { useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import { TimelineView, type TimelineViewProps } from './TimelineView';
import { BAR_WIDTH, LABEL_WIDTH, computeLayout, type Layout } from './layout';
import './TimelineOverview.css';

/** Bars in the detail below the overview. */
const BARS_PER_VIEW = 4;
const DETAIL_WIDTH = BARS_PER_VIEW * BAR_WIDTH;

type Props = Omit<
  TimelineViewProps,
  'view' | 'highlight' | 'label' | 'decorative' | 'className'
> & {
  /**
   * How the detail moves, under trial (BL-127). False (A): it shows the
   * four-bar page with the pinned chord, else the playhead, and switches
   * page by page. True (B): it keeps the playhead at its centre while
   * playing; stopped, it centres the pinned chord, else where a slider puts it.
   */
  follow?: boolean;
};

/**
 * The timeline as the screen shows it (BL-080): a short one whole, a long
 * one as the whole piece small on top and four bars large below. Clicking
 * a chord in the overview pins it and so moves the detail. The overview is
 * a duplicate drawing, hidden from assistive tech.
 */
export function TimelineOverview({ follow = false, ...props }: Props) {
  const [manual, setManual] = useState(0); // the detail's centre while stopped (B), in ticks
  const geometry = computeLayout(props.timeline);
  if (geometry.bars <= BARS_PER_VIEW) return <TimelineView {...props} />;
  const pinned =
    props.selectedHarmony != null ? props.timeline.harmony[props.selectedHarmony] : undefined;
  const detail = follow
    ? centredOn(geometry, followed(props, pinned, manual))
    : pageOf(geometry, pinned ? pinned.start : props.positionTick);
  const sliding = follow && props.playing === false && !pinned;
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

/** The tick B keeps at the centre of the detail. */
function followed(
  props: { positionTick: number; playing?: boolean | undefined },
  pinned: Timeline['harmony'][number] | undefined,
  manual: number,
) {
  if (props.playing !== false) return props.positionTick;
  return pinned ? pinned.start + pinned.duration / 2 : manual;
}

/** A: the four-bar page that holds tick. */
function pageOf(geometry: Layout, tick: number) {
  const page = Math.floor(tick / (geometry.perBar * BARS_PER_VIEW));
  return { x: LABEL_WIDTH + page * DETAIL_WIDTH, width: DETAIL_WIDTH };
}

/** B: four bars with tick at the centre, kept inside the piece. */
function centredOn(geometry: Layout, tick: number) {
  const x = geometry.x(tick) - DETAIL_WIDTH / 2;
  return {
    x: Math.min(Math.max(x, LABEL_WIDTH), geometry.width - DETAIL_WIDTH),
    width: DETAIL_WIDTH,
  };
}
