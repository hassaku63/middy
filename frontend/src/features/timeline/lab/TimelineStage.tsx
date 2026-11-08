import { useRef } from 'react';
import { TimelineView, type TimelineViewProps } from '../TimelineView';
import { BAR_WIDTH, LABEL_WIDTH, computeLayout } from '../layout';
import { useFollowPlayhead } from './useFollowPlayhead';
import { BARS_PER_VIEW, type TimelineLayout } from './variants';
import './lab.css';

type Props = Omit<
  TimelineViewProps,
  'view' | 'highlight' | 'label' | 'decorative' | 'className'
> & {
  layout: TimelineLayout;
};

/** Draws the timeline in the layout under trial (BL-080). */
export function TimelineStage({ layout, ...props }: Props) {
  switch (layout) {
    case 'scroll':
      return <ScrollTimeline {...props} />;
    case 'rows':
      return <RowsTimeline {...props} />;
    case 'overview':
      return <OverviewTimeline {...props} />;
    default:
      return <TimelineView {...props} />;
  }
}

type ViewProps = Omit<Props, 'layout'>;

/** Natural size (160 px a bar), scrolled sideways, following the playhead. */
function ScrollTimeline(props: ViewProps) {
  const ref = useRef<HTMLDivElement>(null);
  const geometry = computeLayout(props.timeline);
  useFollowPlayhead(ref, geometry.x(props.positionTick));
  return (
    <div ref={ref} className="lab-scroll" data-testid="timeline-scroll">
      <TimelineView {...props} className="timeline--natural" />
    </div>
  );
}

/** Wrapped like a lead sheet: four bars per row, rows stacked. */
function RowsTimeline(props: ViewProps) {
  const geometry = computeLayout(props.timeline);
  const rows = Math.ceil(geometry.bars / BARS_PER_VIEW);
  return (
    <div className="lab-rows">
      {Array.from({ length: rows }, (_, r) => {
        const first = r * BARS_PER_VIEW;
        const count = Math.min(BARS_PER_VIEW, geometry.bars - first);
        return (
          <TimelineView
            key={r}
            {...props}
            label={`タイムライン（${first + 1}〜${first + count} 小節）`}
            view={{ x: LABEL_WIDTH + first * BAR_WIDTH, width: BARS_PER_VIEW * BAR_WIDTH }}
          />
        );
      })}
    </div>
  );
}

/**
 * The whole piece small on top, four bars large below. The detail shows the
 * page with the pinned chord, else the one under the playhead; clicking a
 * chord in the overview pins it and so moves the detail.
 */
function OverviewTimeline(props: ViewProps) {
  const geometry = computeLayout(props.timeline);
  if (geometry.bars <= BARS_PER_VIEW) return <TimelineView {...props} />;
  const pinned =
    props.selectedHarmony != null ? props.timeline.harmony[props.selectedHarmony] : undefined;
  const anchor = pinned ? pinned.start : props.positionTick;
  const page = Math.floor(anchor / (geometry.perBar * BARS_PER_VIEW));
  const detail = {
    x: LABEL_WIDTH + page * BARS_PER_VIEW * BAR_WIDTH,
    width: BARS_PER_VIEW * BAR_WIDTH,
  };
  return (
    <div className="lab-overview">
      <TimelineView {...props} decorative highlight={detail} className="timeline--overview" />
      <TimelineView {...props} view={detail} />
    </div>
  );
}
