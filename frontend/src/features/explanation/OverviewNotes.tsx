import type { Overview } from '@/shared/api/client';

/**
 * The example as a whole, above the per-chord notes (BL-098): what its mode
 * sounds like and how the progression runs. It does not follow playback.
 */
export function OverviewNotes({ overview }: { overview: Overview }) {
  return (
    <div className="explanation__overview" aria-label="全体と流れ">
      <p className="explanation__overview-line">
        <span className="explanation__overview-label">全体</span>
        {overview.whole}
      </p>
      <p className="explanation__overview-line">
        <span className="explanation__overview-label">流れ</span>
        {overview.flow}
      </p>
    </div>
  );
}
