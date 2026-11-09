import type { useDegreeMapTrial } from './useDegreeMapTrial';

/**
 * The switches of the screen trials, shown with ?lab: today the degree map
 * (BL-099). They go away with the trials.
 */
export function LabControls({ degreeMap }: { degreeMap: ReturnType<typeof useDegreeMapTrial> }) {
  if (!degreeMap.lab) return null;
  return (
    <label className="app__lab">
      <input
        type="checkbox"
        checked={degreeMap.on}
        onChange={(e) => degreeMap.toggle(e.target.checked)}
      />
      度数の地図（試験）
    </label>
  );
}
