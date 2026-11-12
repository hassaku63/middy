import type { CSSProperties, MouseEvent } from 'react';
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
 * pinned chord, else it stays where it was scrolled. Clicking the drawing
 * moves the playhead (BL-328, ADR-0067); clicking the strip only brings
 * the box there (BL-333, ADR-0069). Scrolling the box by hand or from the
 * strip while playing pauses the following until 「再生位置へ戻る」 or the
 * next play (BL-332, ADR-0068). The strip is a duplicate
 * drawing, hidden from assistive tech.
 */
export function TimelineOverview(props: Props) {
  const geometry = computeLayout(props.timeline);
  const long = geometry.bars > BARS_PER_VIEW;
  const pinned =
    props.selectedHarmony != null ? props.timeline.harmony[props.selectedHarmony] : undefined;
  const playing = props.playing !== false;
  // Any timeline may scroll (a narrow screen), so the box follows the
  // playhead or the pinned chord.
  const target = followed(geometry, props.positionTick, playing, pinned);
  const lessMotion = useReducedMotion();
  const { ref, view, onScroll, paused, resume, look } = useTimelineScroll(
    geometry.width,
    target,
    playing && lessMotion ? 'page' : 'centre',
    geometry.labelWidth,
    playing,
  );
  // The strip picks where to look, not where to play (ADR-0069): a click
  // on it brings the box there, as a DAW's overview does.
  const lookFromStrip = (e: MouseEvent<HTMLDivElement>) => {
    const strip = e.target instanceof Element ? e.target.closest('svg') : null;
    const box = strip?.getBoundingClientRect();
    if (!box?.width) return;
    look(((e.clientX - box.left) / box.width) * geometry.width);
  };
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
      {paused && (
        // Scrolled by hand while playing, the box stops following (BL-332,
        // ADR-0068, as a DAW's follow): this takes it up again.
        <button type="button" className="timeline-overview__follow" onClick={resume}>
          再生位置へ戻る
        </button>
      )}
      {long && (
        <div className="timeline-overview__strip" onClick={lookFromStrip}>
          <TimelineView
            {...props}
            // A place to look, not a chord to pin or a point to play.
            onSelectHarmony={undefined}
            onSeek={undefined}
            fit={false}
            decorative
            stretch
            highlight={view}
            className="timeline--overview"
          />
        </div>
      )}
    </div>
  );
}

/**
 * The point kept in view, in drawing units: the playhead while playing;
 * stopped, the pinned chord's middle, else the beginning once ■ took the
 * playhead back there (BL-334), else nothing (the box stays where it was).
 */
function followed(
  geometry: Layout,
  positionTick: number,
  playing: boolean,
  pinned: Timeline['harmony'][number] | undefined,
) {
  if (playing) return geometry.x(positionTick);
  if (pinned) return geometry.x(pinned.start + pinned.duration / 2);
  return positionTick === 0 ? geometry.x(0) : null;
}
