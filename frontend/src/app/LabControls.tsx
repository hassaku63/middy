import type { useLayoutTrial } from './useLayoutTrial';
import type { useZoomTrial } from './useZoomTrial';

/**
 * The switches of the screen trials, shown with ?lab: how the timeline's
 * detail moves (BL-127) and where the conversation sits (BL-184). They go
 * away with the trials.
 */
export function LabControls({
  zoom,
  layout,
}: {
  zoom: ReturnType<typeof useZoomTrial>;
  layout: ReturnType<typeof useLayoutTrial>;
}) {
  if (!zoom.lab) return null;
  return (
    <>
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
      <label className="app__lab">
        画面の配置（試験）
        <select
          value={layout.desk ? 'desk' : 'stacked'}
          onChange={(e) => layout.choose(e.target.value === 'desk')}
        >
          <option value="stacked">会話が上、結果が下</option>
          <option value="desk">会話を左の列に</option>
        </select>
      </label>
    </>
  );
}
