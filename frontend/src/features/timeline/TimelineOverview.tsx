import { TimelineView, type TimelineViewProps } from './TimelineView';
import { BAR_WIDTH, LABEL_WIDTH, computeLayout } from './layout';
import './TimelineOverview.css';

/** Bars in the detail below the overview. */
const BARS_PER_VIEW = 4;

type Props = Omit<TimelineViewProps, 'view' | 'highlight' | 'label' | 'decorative' | 'className'>;

/**
 * The timeline as the screen shows it (BL-080): a short one whole, a long
 * one as the whole piece small on top and four bars large below. The
 * detail shows the page with the pinned chord, else the one under the
 * playhead; clicking a chord in the overview pins it and so moves the
 * detail. The overview is a duplicate drawing, hidden from assistive tech.
 */
export function TimelineOverview(props: Props) {
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
    <div className="timeline-overview">
      <TimelineView {...props} decorative highlight={detail} className="timeline--overview" />
      <TimelineView {...props} view={detail} />
    </div>
  );
}
