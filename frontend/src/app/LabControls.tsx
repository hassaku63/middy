import { TIMELINE_LAYOUTS, isTimelineLayout } from '@/features/timeline/lab/variants';
import type { useDegreeMapTrial } from './useDegreeMapTrial';
import type { useTimelineLayout } from './useTimelineLayout';

/**
 * The switches of the screen trials, shown with ?lab: the long-timeline
 * layout (BL-080) and the degree map (BL-099). They go away with the trials.
 */
export function LabControls({
  layout,
  degreeMap,
}: {
  layout: ReturnType<typeof useTimelineLayout>;
  degreeMap: ReturnType<typeof useDegreeMapTrial>;
}) {
  if (!layout.lab) return null;
  return (
    <>
      <label className="app__lab">
        表示（試験）
        <select
          value={layout.layout}
          onChange={(e) => {
            if (isTimelineLayout(e.target.value)) layout.choose(e.target.value);
          }}
        >
          {TIMELINE_LAYOUTS.map((l) => (
            <option key={l.id} value={l.id}>
              {l.name}
            </option>
          ))}
        </select>
      </label>
      <label className="app__lab">
        <input
          type="checkbox"
          checked={degreeMap.on}
          onChange={(e) => degreeMap.toggle(e.target.checked)}
        />
        度数の地図（試験）
      </label>
    </>
  );
}
