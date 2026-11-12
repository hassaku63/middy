import type { CSSProperties } from 'react';
import type { Timeline } from '@/shared/api/client';
import { TimelineLabels } from './TimelineLabels';
import { TimelineView, type TimelineViewProps } from './TimelineView';
import { BAR_WIDTH, computeLayout, type Layout } from './layout';
import { useReducedMotion } from './useReducedMotion';
import { useTimelineScroll } from './useTimelineScroll';
import './TimelineOverview.css';

/** Bars that fit the scrolling box's width, at most, when its height allows. */
const BARS_PER_VIEW = 4;

type Props = Omit<
  TimelineViewProps,
  'highlight' | 'label' | 'decorative' | 'className' | 'stretch'
>;

/**
 * The timeline as the screen shows it: a short one whole; a long one as
 * one drawing that scrolls sideways, with the whole piece as a thin strip
 * below that outlines the part in view, as an editor's overview (#360,
 * the owner's choice). Playing, the box keeps the playhead in the middle
 * of the drawing in view (BL-127, BL-326, ADR-0066), or turns a page when
 * the person asked the system for less motion; stopped, it centres the
 * pinned chord, else it stays where it was scrolled. Clicking a chord in
 * the strip pins it. The strip is a duplicate drawing, hidden from
 * assistive tech.
 */
export function TimelineOverview(props: Props) {
  const geometry = computeLayout(props.timeline);
  const long = geometry.bars > BARS_PER_VIEW;
  const pinned =
    props.selectedHarmony != null ? props.timeline.harmony[props.selectedHarmony] : undefined;
  const playing = props.playing !== false;
  // Any timeline may scroll (a narrow screen), so the box follows the playhead or the pinned chord.
  const target = followed(geometry, props.positionTick, playing, pinned);
  const lessMotion = useReducedMotion();
  const { ref, view, onScroll } = useTimelineScroll(
    geometry.width,
    target,
    playing && lessMotion ? 'page' : 'centre',
    geometry.labelWidth,
  );
  // One box for every timeline, so the tree does not change with the
  // width (ADR-0042): it scrolls sideways when the drawing is wider than
  // the box, with the label column held at its left (#402 A). A short
  // timeline fills the box's width and widens its bars under the height
  // cap (ADR-0057); it never shrinks below its own size, so on a narrow
  // screen it scrolls instead of shrinking its text. A long one shows four
  // bars across, at its own size at least, with the whole piece as a strip
  // below (#360).
  const vars = {
    '--timeline-aspect': geometry.width / geometry.height,
    '--timeline-natural-width': `${String(geometry.width)}px`,
    '--timeline-four-bars-width': `${String(geometry.labelWidth + BARS_PER_VIEW * BAR_WIDTH)}px`,
    '--timeline-four-bars': geometry.height / (geometry.labelWidth + BARS_PER_VIEW * BAR_WIDTH),
    '--timeline-label-ratio': geometry.labelWidth / geometry.height,
  } as CSSProperties;
  return (
    <div className="timeline-overview">
      <div
        className={`timeline-overview__scroll${long ? ' is-long' : ''}`}
        ref={ref}
        onScroll={onScroll}
        style={vars}
      >
        <TimelineLabels layout={geometry} />
        <TimelineView {...props} fit={!long && props.fit === true} />
      </div>
      {long && (
        <TimelineView
          {...props}
          fit={false}
          decorative
          stretch
          highlight={view}
          className="timeline--overview"
        />
      )}
    </div>
  );
}

/** The point kept in view, in drawing units: the playhead, or the pinned chord's middle. */
function followed(
  geometry: Layout,
  positionTick: number,
  playing: boolean,
  pinned: Timeline['harmony'][number] | undefined,
) {
  if (playing) return geometry.x(positionTick);
  return pinned ? geometry.x(pinned.start + pinned.duration / 2) : null;
}
