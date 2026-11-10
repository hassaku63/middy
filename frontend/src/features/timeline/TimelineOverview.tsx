import type { CSSProperties } from 'react';
import type { Timeline } from '@/shared/api/client';
import { TimelineView, type TimelineViewProps } from './TimelineView';
import { BAR_WIDTH, LABEL_WIDTH, computeLayout, type Layout } from './layout';
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
 * the owner's choice). Playing, the box turns a page when the playhead
 * nears an edge; stopped, it centres the pinned chord, else it stays where
 * it was scrolled. Clicking a chord in the strip pins it. The strip is a
 * duplicate drawing, hidden from assistive tech.
 */
export function TimelineOverview(props: Props) {
  const geometry = computeLayout(props.timeline);
  const long = geometry.bars > BARS_PER_VIEW;
  const pinned =
    props.selectedHarmony != null ? props.timeline.harmony[props.selectedHarmony] : undefined;
  const playing = props.playing !== false;
  const target = !long ? null : followed(geometry, props.positionTick, playing, pinned);
  const { ref, view, onScroll } = useTimelineScroll(
    geometry.width,
    target,
    playing ? 'page' : 'centre',
  );
  if (!long) return <TimelineView {...props} />;
  // The box shows four bars across its width, unless the height cap
  // (ADR-0041) makes the drawing smaller.
  const fourBars = geometry.height / (LABEL_WIDTH + BARS_PER_VIEW * BAR_WIDTH);
  return (
    <div className="timeline-overview">
      <div
        className="timeline-overview__scroll"
        ref={ref}
        onScroll={onScroll}
        style={{ '--timeline-four-bars': fourBars } as CSSProperties}
      >
        <TimelineView {...props} />
      </div>
      <TimelineView {...props} decorative stretch highlight={view} className="timeline--overview" />
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
