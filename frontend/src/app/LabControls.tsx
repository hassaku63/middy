import type { Layout, useLayoutTrial } from './useLayoutTrial';
import type { useZoomTrial } from './useZoomTrial';

/**
 * The switches of the screen trials, shown with ?lab: how the timeline's
 * detail moves (BL-127), where the conversation sits (BL-184) and where
 * its input sits (BL-218). They go away with the trials.
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
        <select value={layout.mode} onChange={(e) => layout.choose(e.target.value as Layout)}>
          <option value="stacked">会話が上、結果が下</option>
          <option value="desk">会話を左の列に</option>
          <option value="wide">会話を左の列に、タイムラインを横いっぱいに</option>
        </select>
      </label>
      {layout.desk && (
        <label className="app__lab">
          入力欄の位置（試験）
          <select
            value={layout.followInput ? 'follow' : 'pinned'}
            onChange={(e) => layout.chooseInput(e.target.value === 'follow')}
          >
            <option value="pinned">列の下に固定</option>
            <option value="follow">やり取りのすぐ下（空いていれば上に寄る）</option>
          </select>
        </label>
      )}
    </>
  );
}
