import type { useZoomTrial } from './useZoomTrial';

/**
 * The switches of the screen trials, shown with ?lab: today how the
 * timeline's detail moves (BL-127). They go away with the trials.
 */
export function LabControls({ zoom }: { zoom: ReturnType<typeof useZoomTrial> }) {
  if (!zoom.lab) return null;
  return (
    <label className="app__lab">
      拡大の動き（試験）
      <select
        value={zoom.follow ? 'follow' : 'page'}
        onChange={(e) => zoom.choose(e.target.value === 'follow')}
      >
        <option value="page">4 小節ずつ切り替え</option>
        <option value="follow">再生位置に追随</option>
      </select>
    </label>
  );
}
